package kafka

import (
	"context"
	"time"

	"github.com/preciouswxe/EchoBoard_backend/setting"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	Writer *kafka.Writer
)

func Init(cfg *setting.KafkaConfig) error {
	// 解析压缩算法
	compression := parseCompression(cfg.Compression)
	// 解析确认级别
	requiredAcks := parseRequiredAcks(cfg.RequiredAcks)

	Writer = &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: requiredAcks,
		Compression:  compression,
		MaxAttempts:  cfg.MaxAttempts,
		BatchSize:    cfg.BatchSize,
		BatchTimeout: time.Duration(cfg.BatchTimeout) * time.Millisecond,
	}
	// 设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试连接
	conn, err := kafka.DialLeader(ctx, "tcp", cfg.Brokers[0], "test", 0)
	if err != nil {
		zap.L().Error("Failed to connect to Kafka", zap.Error(err))
		return err
	}
	conn.Close()

	zap.L().Info("Kafka initialized successfully",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("compression", cfg.Compression),
		zap.Int("max_attempts", cfg.MaxAttempts),
		zap.Int("batch_size", cfg.BatchSize),
	)
	return nil
}

func Close() error {
	if Writer != nil {
		return Writer.Close()
	}
	return nil
}

// parseCompression 解析压缩算法
func parseCompression(compression string) kafka.Compression {
	switch compression {
	case "gzip":
		return kafka.Gzip
	case "snappy":
		return kafka.Snappy
	case "lz4":
		return kafka.Lz4
	case "zstd":
		return kafka.Zstd
	case "none":
		return kafka.Compression(0) // 无压缩
	default:
		zap.L().Warn("Unknown compression, using snappy", zap.String("compression", compression))
		return kafka.Snappy
	}
}

// parseRequiredAcks 解析确认级别
func parseRequiredAcks(acks int) kafka.RequiredAcks {
	switch acks {
	case 0:
		return kafka.RequireNone // 不等待确认（最快，可能丢消息）
	case 1:
		return kafka.RequireOne  // 等待 leader 确认（推荐）
	case -1:
		return kafka.RequireAll  // 等待所有副本确认（最安全，最慢）
	default:
		zap.L().Warn("Unknown required_acks, using RequireOne", zap.Int("acks", acks))
		return kafka.RequireOne
	}
}