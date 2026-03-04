package logic

import (
	"go.uber.org/zap"

	dao_kafka "github.com/preciouswxe/EchoBoard_backend/dao/kafka"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
)

func CollectPost(postID, userID int64) (err error) {
	// 1. 先写 Redis
	if err = redis.CollectPost(userID, postID); err != nil {
		zap.L().Error("redis.CollectPost failed", zap.Error(err))
		return
	}

	// 2. 异步发送到 Kafka（不阻塞）- mysql持久化由消费者端进行
	go func() {
		if err := dao_kafka.PublishInteractionEvent(ActionCollect, postID, userID); err != nil {
			zap.L().Error("Failed to publish collect event", zap.Error(err))
			// Kafka 失败,回滚 Redis
		}
	}()
	return
}

func CancelCollectPost(postID, userID int64) (err error) {
	// 1. 先写 Redis
	if err = redis.CancelCollectPost(userID, postID); err != nil {
		zap.L().Error("redis.CancelCollectPost failed", zap.Error(err))
		return
	}

	// 2. 异步发送到 Kafka（不阻塞）
	go func() {
		if err := dao_kafka.PublishInteractionEvent(ActionCancelCollect, postID, userID); err != nil {
			zap.L().Error("Failed to publish cancel collect event", zap.Error(err))
		}
	}()
	return
}