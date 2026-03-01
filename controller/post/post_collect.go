package post

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/controller"
	"github.com/preciouswxe/EchoBoard_backend/logic"
)

func CollectPostHandler(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error("Collect Post with invalid param", zap.Error(err), zap.String("postIDStr", postIDStr))
		controller.ResponseError(c, controller.CodeInvalidParam)
		return
	}
	userID, err := controller.GetCurrentUser(c)
	if err != nil {
		zap.L().Error("getCurrentUser failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeNeedLogin)
		return
	}
	// 先检查是否已收藏
	isCollected, _ := redis.IsUserCollectedPost(userID, postID)
	if isCollected {
		c.JSON(200, gin.H{"msg": "已经点赞过了"})
		return
	}
	if err := logic.CollectPost(postID, userID); err != nil {
		zap.L().Error("logic.CollectPost failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeServerBusy)
		return
	}
	controller.ResponseSuccess(c, nil)
}

func CancelCollectPostHandler(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error("Cancel Collect Post with invalid param", zap.Error(err), zap.String("postIDStr", postIDStr))
		controller.ResponseError(c, controller.CodeInvalidParam)
		return
	}
	userID, err := controller.GetCurrentUser(c)
	if err != nil {
		zap.L().Error("getCurrentUser failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeNeedLogin)
		return
	}
	if err := logic.CancelCollectPost(postID, userID); err != nil {
		zap.L().Error("logic.CancelCollectPost failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeServerBusy)
		return
	}
	controller.ResponseSuccess(c, nil)
}