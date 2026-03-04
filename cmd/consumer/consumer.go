package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/preciouswxe/EchoBoard_backend/models"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	dao_kafka "github.com/preciouswxe/EchoBoard_backend/dao/kafka"
	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"github.com/preciouswxe/EchoBoard_backend/logger"
	"github.com/preciouswxe/EchoBoard_backend/setting"
)

const (
	ActionLike          = "like"
	ActionUnlike        = "unlike"
	ActionCollect       = "collect"
	ActionCancelCollect = "cancel_collect"
)

// processEvent 处理消息对应的业务逻辑
func processEvent(event models.PostInteractionEvent) error {
	switch event.Action {
	case ActionLike:
		return mysql.LikePost(event.PostID, event.UserID)
	case ActionUnlike:
		return mysql.UnlikePost(event.PostID, event.UserID)
	case ActionCollect:
		return mysql.CollectPost(event.PostID, event.UserID)
	case ActionCancelCollect:
		return mysql.CancelCollectPost(event.PostID, event.UserID)
	default:
		return fmt.Errorf("unknown action: %s", event.Action)
	}
}

// consumePostInteraction 消费帖子互动事件落库
func consumePostInteraction(ctx context.Context) {
	cfg := setting.Conf.KafkaConfig
	// 创建 Kafka Reader（消费者）
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          dao_kafka.TopicPostInteraction,
		GroupID:        cfg.ConsumerGroup,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		CommitInterval: time.Second,       // 每秒自动提交 offset
		StartOffset:    kafka.LastOffset, // 生产环境从最新开始消费
	})
	defer reader.Close()

	zap.L().Info("Consumer started",
		zap.String("topic", dao_kafka.TopicPostInteraction),
		zap.String("group_id", cfg.ConsumerGroup),
	)

	// 循环读取消息
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				zap.L().Info("Consumer stopped by context")
				break
			}
			zap.L().Error("Failed to read message", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		handlePostInteractionEvent(ctx, msg)
	}
}

// handlePostInteractionEvent 处理读取到的消息
func handlePostInteractionEvent(ctx context.Context, msg kafka.Message) {
	// 反序列化 json 消息
	var event models.PostInteractionEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		zap.L().Error("Failed to unmarshal event", zap.Error(err), zap.ByteString("value", msg.Value))
		// 反序列化失败,直接进死信队列,不重试
		sendToDLQ(ctx, msg, fmt.Errorf("unmarshal failed: %w", err))
		return
	}
	zap.L().Debug("Processing event",
		zap.String("action", event.Action),
		zap.Int64("user_id", event.UserID),
		zap.Int64("post_id", event.PostID),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
	)
	// 获取当前重试次数
	retryCount := getRetryCount(msg.Headers)
	// 执行业务逻辑：根据 action 写入 mysql
	err := processEvent(event)
	if err != nil {
		zap.L().Error("Failed to persist event to MySQL",
			zap.String("action", event.Action),
			zap.Int64("user_id", event.UserID),
			zap.Int64("post_id", event.PostID),
			zap.Error(err),
		)
		// 判断是否需要重试
		if retryCount < MaxRetries {
			// 重试：延迟后重新处理
			backoffDuration := time.Duration( 1<<uint(retryCount) ) * time.Second // 指数退避 1 2 4 8...
			zap.L().Info("Scheduling retry",
				zap.Int("retry_count", retryCount+1),
				zap.Duration("backoff", backoffDuration))
			time.Sleep(backoffDuration)
			// 更新重试次数
			msg.Headers = setRetryCount(msg.Headers, retryCount+1)
			handlePostInteractionEvent(ctx, msg)
		} else{
			// 超过最大重试次数，进死信队列
			sendToDLQ(ctx, msg, fmt.Errorf("max retries exceeded: %w", err))
		}
	} else {
		zap.L().Info("Event persisted to MySQL successfully",
			zap.String("action", event.Action),
			zap.Int64("user_id", event.UserID),
			zap.Int64("post_id", event.PostID),
			zap.Int64("offset", msg.Offset),
		)
	}
}

func main() {
	// 1. 加载配置文件
	if err := setting.Init(); err != nil {
		fmt.Printf("init setting failed, err:%v\n", err)
		return
	}

	// 2. 初始化日志
	if err := logger.Init(setting.Conf.LogConfig, setting.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	// 把缓冲区的日志加进来
	defer zap.L().Sync()

	// 3. 初始化 MySQL 连接
	if err := mysql.Init(setting.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	defer mysql.Close()

	// 4. 初始化 DLQ Writer
	initDLQWriter()
	defer dlqWriter.Close()

	// 5. 从配置读取 MaxRetries
	MaxRetries = setting.Conf.KafkaConfig.MaxRetries
	if MaxRetries == 0 {
		MaxRetries = 3 // 默认值
	}

	zap.L().Info("Consumer starting...",
		zap.Strings("brokers", setting.Conf.KafkaConfig.Brokers),
		zap.String("consumer_group", setting.Conf.KafkaConfig.ConsumerGroup),
		zap.String("dlq_topic", setting.Conf.KafkaConfig.DLQTopic),
		zap.Int("max_retries", MaxRetries),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumePostInteraction(ctx)

	// 5. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("Consumer shutting down...")
	time.Sleep(2 * time.Second) // 等待正在处理的消息完成
	zap.L().Info("Consumer shut down successfully~")
}