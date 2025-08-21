package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/preciouswxe/EchoBoard_backend/logic"
	"go.uber.org/zap"
)

// ---- 跟社区相关的 ----

// CommunityHandler
// @Summary 展示所有社区
// @Description 查询到所有的社区（community_id, community_name） 以列表形式返回
// @Tags 社区相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户 token 令牌"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponseCommunityAllList
// @Router /api/v1/community [get]
func CommunityHandler(c *gin.Context) {
	// 查询到所有的社区（community_id, community_name） 以列表形式返回
	data, err := logic.GetCommunityList()
	if err != nil {
		zap.L().Error("logic.GetCommunityList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

// CommunityDetailHandler
// @Summary 社区分类详情
// @Description 根据社区 id 获取社区详情
// @Tags 社区相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户 token 令牌"
// @Param id path int true "社区 id"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponseCommunityDetailList
// @Router /api/v1/community/{id} [get]
func CommunityDetailHandler(c *gin.Context) {
	// 获取社区 id 以及 参数校验
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 根据 id 获取社区详情
	data, err := logic.GetCommunityDetail(id)
	if err != nil {
		zap.L().Error("logic.GetCommunityDetail() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}
