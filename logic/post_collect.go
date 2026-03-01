package logic

import (
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
)

func CollectPost(postID, userID int64) (err error) {
	// 1. 先写 Redis
	if err = redis.CollectPost(userID, postID); err != nil {
		zap.L().Error("redis.CollectPost failed", zap.Error(err))
		return
	}

	// 2. 再写 MySQL
	// TODO: 后续改用消息队列异步写入
	if err = mysql.CollectPost(postID, userID); err != nil {
		zap.L().Error("mysql.CollectPost failed", zap.Error(err))
		return
	}
	return
}

func CancelCollectPost(postID, userID int64) (err error) {
	// 1. 先写 Redis
	if err = redis.CancelCollectPost(userID, postID); err != nil {
		zap.L().Error("redis.CancelCollectPost failed", zap.Error(err))
		return
	}

	// 2. 再写 MySQL
	// TODO: 后续改用消息队列异步写入
	if err = mysql.CancelCollectPost(postID, userID); err != nil {
		zap.L().Error("mysql.CancelCollectPost failed", zap.Error(err))
		return
	}
	return
}