package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/preciouswxe/EchoBoard_backend/logic"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"go.uber.org/zap"
	"strconv"
)

// CreatePostHandler 创建帖子
func CreatePostHandler(c *gin.Context) {
	p := new(models.Post)
	// 获取参数及参数校验 (validator)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Debug("c.ShouldBindJSON(p) error", zap.Any("err", err))
		zap.L().Error("create post with invalid param")
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 从 context 取到当前发送请求的用户的 ID
	userID, err := getCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	// 发帖人和登录用户 ID 保持一致
	p.AuthorID = userID
	// 创建帖子
	if err := logic.CreatePost(p); err != nil {
		zap.L().Error("logic.CreatePost(p) failed:", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 返回响应
	ResponseSuccess(c, nil)
}

// GetPostDetailHandler 获取帖子详情
func GetPostDetailHandler(c *gin.Context) {
	// 获取参数（从 url 获取帖子 id ）
	postIDStr := c.Param("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		zap.L().Error("get post detail with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 根据 id 取出帖子数据
	data, err := logic.GetPostById(postID)
	if err != nil {
		zap.L().Error("logic.GetPostById(postID) failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 返回响应
	ResponseSuccess(c, data)
}

// GetPostListHandler 获取帖子列表
func GetPostListHandler(c *gin.Context) {
	// 获取分页参数
	page, size := getPageInfo(c)
	// 获取数据
	data, err := logic.GetPostList(page, size)
	if err != nil {
		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 返回响应
	ResponseSuccess(c, data)
}
