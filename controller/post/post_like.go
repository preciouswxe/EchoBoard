package post

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/controller"
	"github.com/preciouswxe/EchoBoard_backend/logic"
)

func LikePostHandler(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error("Like Post with invalid param", zap.Error(err), zap.String("postIDStr", postIDStr))
		controller.ResponseError(c, controller.CodeInvalidParam)
		return
	}
	userID, err := controller.GetCurrentUser(c)
	if err != nil {
		zap.L().Error("getCurrentUser failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeNeedLogin)
		return
	}
	// 先检查是否已点赞（可选，因为 LikePost 内部也会检查）
	isLiked, _ := redis.IsUserLikedPost(userID, postID)
	if isLiked {
		c.JSON(200, gin.H{"msg": "已经点赞过了"})
		return
	}

	err = logic.LikePost(postID, userID)
	if err != nil {
		zap.L().Error("logic.LikePost failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeServerBusy)
		return
	}
	controller.ResponseSuccess(c, nil)
}

func UnlikePostHandler(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error("Unlike Post with invalid param", zap.Error(err), zap.String("postIDStr", postIDStr))
		controller.ResponseError(c, controller.CodeInvalidParam)
		return
	}
	userID, err := controller.GetCurrentUser(c)
	if err != nil {
		zap.L().Error("getCurrentUser failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeNeedLogin)
		return
	}
	err = logic.UnlikePost(postID, userID)
	if err != nil {
		zap.L().Error("logic.UnlikePost failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeServerBusy)
		return
	}
	controller.ResponseSuccess(c, nil)
}