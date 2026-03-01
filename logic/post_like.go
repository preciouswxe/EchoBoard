package logic

import (
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
)

func LikePost(postID, userID int64) (err error) {
	// 1. 先写 Redis（快速响应）
	if err = redis.LikePost(userID, postID); err != nil {
		zap.L().Error("redis.LikePost failed", zap.Error(err))
		return
	}

	// 2. 再写 MySQL（持久化）
	// TODO: 后续改用消息队列异步写入
	if err = mysql.LikePost(postID, userID); err != nil {
		zap.L().Error("mysql.LikePost failed", zap.Error(err))
		// 如果 MySQL 失败，需要回滚 Redis 吗？
		// 方案1：不回滚，定时任务同步
		// 方案2：回滚 Redis（redis.UnlikePost）
		return
	}
	return
}

func UnlikePost(postID, userID int64) (err error) {
	// 1. 先写 Redis
	if err = redis.UnlikePost(userID, postID); err != nil {
		zap.L().Error("redis.UnlikePost failed", zap.Error(err))
		return
	}

	// 2. 再写 MySQL
	// TODO: 后续改用消息队列异步写入
	if err = mysql.UnlikePost(postID, userID); err != nil {
		zap.L().Error("mysql.UnlikePost failed", zap.Error(err))
		return
	}
	return
}