package post

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/preciouswxe/EchoBoard_backend/controller"
	"github.com/preciouswxe/EchoBoard_backend/logic"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"go.uber.org/zap"
)

func CreateCommentHandler(c *gin.Context) {
	cm := new(models.Comment)
	if err := c.ShouldBindJSON(cm); err != nil {
		zap.L().Error("Create Comment with invalid param", zap.Error(err), zap.Int64("postIDStr", cm.PostID))
		controller.ResponseError(c, controller.CodeInvalidParam)
		return
	}
	userID, err := controller.GetCurrentUser(c)
	if err != nil {
		zap.L().Error("getCurrentUser failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeNeedLogin)
		return
	}
	cm.UserID = userID
	err = logic.CreateComment(cm)
	if err != nil {
		zap.L().Error("logic.CreateComment failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeServerBusy)
		return
	}
	controller.ResponseSuccess(c, nil)
}

func GetCommentListHandler(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		controller.ResponseError(c, controller.CodeInvalidParam)
		return
	}
	comments, err := logic.GetCommentListByPostID(postID)
	if err != nil {
		zap.L().Error("logic.GetCommentListByPostID failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeServerBusy)
		return
	}
	controller.ResponseSuccess(c, comments)
}