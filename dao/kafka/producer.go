package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/models"
	pkg_Kafka "github.com/preciouswxe/EchoBoard_backend/pkg/kafka"
)

const (
	TopicPostInteraction = "post_interaction"
)

// PublishInteractionEvent 发布用户互动事件到 Kafka
func PublishInteractionEvent(action string, postID , userID int64) error {
	// 自定义事件内容填充
	event := models.PostInteractionEvent{
		Action: action,
		PostID: postID,
		UserID: userID,
		TimeStamp: time.Now().Unix(),
	}
	// 序列化为 JSON
	msgBytes, err := json.Marshal(event)
	if err != nil {
		zap.L().Error("Failed to marshal interaction event", zap.Error(err))
		return err
	}
	// 构造 Kafka 消息
	msg := kafka.Message{
		Topic: TopicPostInteraction,
		Key: []byte(action), // 使用 action 作为 key
		Value: msgBytes,
		Time: time.Now(),
	}
	// 发送消息（带超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pkg_Kafka.Writer.WriteMessages(ctx, msg); err != nil {
		zap.L().Error("Failed to send message to Kafka",
			zap.String("action", action),
			zap.Int64("user_id", userID),
			zap.Int64("post_id", postID),
			zap.Error(err),
		)
		return err
	}
	zap.L().Info("Message sent to Kafka",
		zap.String("action", action),
		zap.Int64("user_id", userID),
		zap.Int64("post_id", postID),
	)
	return nil
}