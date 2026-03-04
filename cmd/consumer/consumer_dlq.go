package main

import (
	"context"
	"strconv"
	"time"

	"github.com/preciouswxe/EchoBoard_backend/setting"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// DLQ 死信队列常量
const (
	HeaderRetryCount = "x-retry-count"
	HeaderErrorMessage = "x-error-message"
	HeaderFailedAt = "x-failed-at"
)

var (
	dlqWriter *kafka.Writer
	MaxRetries int
)

// initDLQWriter 初始化死信队列生产者
func initDLQWriter() {
	cfg := setting.Conf.KafkaConfig
	dlqWriter = &kafka.Writer{
		Addr: kafka.TCP(cfg.Brokers...),
		Topic:        cfg.DLQTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
	}
	zap.L().Info("DLQ writer initialized", zap.String("topic", cfg.DLQTopic))
}

// getRetryCount 从消息 header 中获取重试次数
func getRetryCount(headers []kafka.Header) int {
	for _, h := range headers {
		if h.Key == HeaderRetryCount {
			count, err := strconv.Atoi(string(h.Value))
			if err == nil {
				return count
			}
		}
	}
	return 0
}

// setRetryCount 设置重试次数到 header
func setRetryCount(headers []kafka.Header, count int) []kafka.Header {
	newHeaders := make([]kafka.Header, 0, len(headers)+1) // Header是结构体，make只能传入切片, 初始长度 0, 预分配容量 len(headers)+1
	// 过滤掉旧的 retry header
	for _, h := range headers {
		if h.Key != HeaderRetryCount {
			newHeaders = append(newHeaders, h)
		}
	}
	// append 一个新的 retry header
	newHeaders = append(newHeaders, kafka.Header{
		Key: HeaderRetryCount,
		Value: []byte(strconv.Itoa(count)),
	})
	return newHeaders
}

// sendToDLQ 发送消息到死信队列
func sendToDLQ(ctx context.Context, msg kafka.Message, err error) {
	// 编写 headers
	headers := msg.Headers
	if headers == nil {
		headers = make([]kafka.Header, 0)
	}
	headers = append(headers,
		kafka.Header{
			Key: HeaderErrorMessage,
			Value: []byte(err.Error()),
		},
		kafka.Header{
			Key: HeaderFailedAt,
			Value: []byte(time.Now().Format(time.RFC3339)),
		},
	)

	// 定义消息
	dlqMsg := kafka.Message{
		Key: msg.Key,
		Value: msg.Value,
		Headers: headers,
	}

	// 发送消息
	if writeErr := dlqWriter.WriteMessages(ctx, dlqMsg); writeErr != nil {
		zap.L().Error("Failed to write to DLQ",
			zap.Error(writeErr),
			zap.String("original_error", err.Error()),
			zap.ByteString("message_value", msg.Value))
	}else {
		zap.L().Warn("Message sent to DLQ",
			zap.String("reason", err.Error()),
			zap.Int("retry_count", getRetryCount(msg.Headers)),
			zap.ByteString("message_value", msg.Value),
		)
	}

}


