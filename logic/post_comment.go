package logic

import (
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"github.com/preciouswxe/EchoBoard_backend/models"

)

func CreateComment(c *models.Comment) (err error) {
	if err = mysql.CreateComment(c); err != nil {
		zap.L().Error("mysql.CreateComment failed", zap.Error(err))
		return
	}
	if redisErr := redis.IncrementPostCommentCount(c.PostID); redisErr != nil {
		zap.L().Warn("increment Feed comment count failed", zap.Error(redisErr))
	}
	return
}

func GetCommentListByPostID(postID int64) (comments []*models.Comment, err error) {
	comments, err = mysql.GetCommentListByPostID(postID)
	if err != nil {
		zap.L().Error("mysql.GetCommentListByPostID failed", zap.Error(err))
		return
	}
	return
}
