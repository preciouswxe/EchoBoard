package post

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/controller"
	"github.com/preciouswxe/EchoBoard_backend/logic"
	"github.com/preciouswxe/EchoBoard_backend/models"
)

// GetFeedHandler returns the cursor-based recommendation feed.
func GetFeedHandler(c *gin.Context) {
	p := &models.ParamFeedList{Size: 20}
	if err := c.ShouldBindQuery(p); err != nil || p.Size < 1 || p.Size > 50 {
		controller.ResponseError(c, controller.CodeInvalidParam)
		return
	}
	userID, err := controller.GetCurrentUser(c)
	if err != nil {
		controller.ResponseError(c, controller.CodeNeedLogin)
		return
	}
	data, err := logic.GetFeed(userID, p.Cursor, p.Size)
	if err != nil {
		zap.L().Error("logic.GetFeed failed", zap.Error(err))
		controller.ResponseError(c, controller.CodeInvalidParam)
		return
	}
	controller.ResponseSuccess(c, data)
}

// ReportFeedImpressionsHandler records cards that actually entered the viewport.
func ReportFeedImpressionsHandler(c *gin.Context) {
	p := new(models.FeedImpressionRequest)
	if err := c.ShouldBindJSON(p); err != nil {
		controller.ResponseError(c, controller.CodeInvalidParam)
		return
	}
	if err := logic.RecordFeedImpressions(p.PostIDs); err != nil {
		controller.ResponseError(c, controller.CodeInvalidParam)
		return
	}
	controller.ResponseSuccess(c, nil)
}
