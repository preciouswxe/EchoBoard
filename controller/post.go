package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/preciouswxe/EchoBoard_backend/logic"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"go.uber.org/zap"
)

// CreatePostHandler 创建帖子
// @Summary 创建帖子
// @Description 根据传入 body 内容建帖
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户 token 令牌"
// @Param post body models.PostCreateRequest true "帖子参数（均必填）"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponseCommon
// @Router /api/v1/post [post]
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

// GetPostDetailHandler 获取单个帖子详情
// @Summary 获取单个帖子详情接口
// @Description 从 url 获取帖子 id 并返回单个帖子详情
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户 token 令牌"
// @Param id path int true "帖子 id"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponsePostList
// @Router /api/v1/post/{id} [get]
func GetPostDetailHandler(c *gin.Context) {
	// 获取参数（从 url 路径直接获取帖子 id ）
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
// @Summary 获取帖子列表
// @Description 从 url 获取帖子 id 并返回帖子详情
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户 token 令牌"
// @Param page query int false "页码，默认1"
// @Param size query int false "页容量，默认10"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponsePostList
// @Router /api/v1/posts [get]
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

// GetPostListHandler2 升级版获取帖子列表接口
// 根据前端传来的参数动态去获取帖子列表
// 按创建时间排序 或者 按分数排序
// 1. 获取 query string 参数
// 2. 去 redis 查询 id 列表
// 3. 根据 id 去数据库查询帖子详细信息
// @Summary 升级版帖子列表接口
// @Description 可按社区按时间或分数排序查询帖子列表接口
// @Tags 帖子相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户 token 令牌"
// @Param page query int false "页码，默认1"
// @Param size query int false "页容量，默认10"
// @Param order query string false "排序方式，time 或 score"
// @Param community_id query int false "社区ID"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponsePostList
// @Router /api/v1/posts2 [get]
func GetPostListHandler2(c *gin.Context) {
	// GET 请求参数: /api/v1/post2?page=1&size=10&order=time
	// 初始化结构体时指定初始参数
	p := &models.ParamPostList{
		Page:  1,
		Size:  10,
		Order: models.OrderTime,
	}

	// 获取参数
	// c.ShouldBind()	  根据请求的数据类型选择相应的方法去获取数据
	// c.ShouldBindJSON() 如果请求中携带的是 json 格式的数据，采用这个方法获取数据
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("GetPostListHandler2 with invalid params", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 获取数据
	data, err := logic.GetPostListNew(p)

	if err != nil {
		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 返回响应
	ResponseSuccess(c, data)
}

// GetCommunityPostListHandler 根据社区去查询帖子列表
//func GetCommunityPostListHandler(c *gin.Context) {
//	p := &models.ParamCommunityPostList{
//		ParamPostList: &models.ParamPostList{
//			Page:  1,
//			Size:  10,
//			Order: models.OrderTime,
//		},
//	}
//
//	// 获取参数
//	// c.ShouldBind()	  根据请求的数据类型选择相应的方法去获取数据
//	// c.ShouldBindJSON() 如果请求中携带的是 json 格式的数据，采用这个方法获取数据
//	if err := c.ShouldBindQuery(p); err != nil {
//		zap.L().Error("GetCommunityPostListHandler with invalid params", zap.Error(err))
//		ResponseError(c, CodeInvalidParam)
//		return
//	}
//
//	// 获取数据
//
//	if err != nil {
//		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
//		ResponseError(c, CodeServerBusy)
//		return
//	}
//
//	// 返回响应
//	ResponseSuccess(c, data)
//}
