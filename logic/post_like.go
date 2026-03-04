package logic

import (
	"go.uber.org/zap"

	dao_kafka "github.com/preciouswxe/EchoBoard_backend/dao/kafka"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
)

func LikePost(postID, userID int64) (err error) {
	// 1. 先写 Redis（快速响应）
	if err = redis.LikePost(userID, postID); err != nil {
		zap.L().Error("redis.LikePost failed", zap.Error(err))
		return
	}



	// 2. 异步发送到 Kafka（不阻塞）- mysql持久化由消费者端进行
	go func() {
		if err := dao_kafka.PublishInteractionEvent(ActionLike, postID, userID); err != nil {
			zap.L().Error("Failed to publish like event", zap.Error(err))
		}
	}()
	return
}

func UnlikePost(postID, userID int64) (err error) {
	// 1. 先写 Redis
	if err = redis.UnlikePost(userID, postID); err != nil {
		zap.L().Error("redis.UnlikePost failed", zap.Error(err))
		return
	}

	// 2. 异步发送到 Kafka（不阻塞）
	go func() {
		if err := dao_kafka.PublishInteractionEvent(ActionUnlike, postID, userID); err != nil {
			zap.L().Error("Failed to publish unlike event", zap.Error(err))
		}
	}()
	return
}