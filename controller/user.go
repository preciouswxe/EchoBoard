package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"github.com/preciouswxe/EchoBoard_backend/logic"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"github.com/preciouswxe/EchoBoard_backend/pkg/jwt"
	"go.uber.org/zap"
)

// SignUpHandler
// @Summary 用户注册
// @Description 用户注册
// @Tags 用户相关接口
// @Accept application/json
// @Produce application/json
// @Param user body models.ParamSignUp true "注册参数：账号、密码（均必填）"
// @Success 200 {object} _ResponseCommon
// @Router /api/v1/signup [post]
func SignUpHandler(c *gin.Context) {
	// 1. 获取参数和参数校验
	p := new(models.ParamSignUp)
	if err := c.ShouldBindJSON(p); err != nil {
		// JSON 请求参数有误，直接返回响应
		zap.L().Error("SignUp with invalid param", zap.Error(err))

		// 判断 err 是不是 validator.ValidatorErrors 类型
		err, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}

		// 返回带信息的错误 全局翻译器翻译后去除无用结构体名
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(err.Translate(trans)))
		return
	}

	// 2. 业务处理
	if err := logic.SignUp(p); err != nil {
		zap.L().Error("logic.SignUp failed", zap.Error(err))
		// 错误分类比较
		if errors.Is(err, mysql.ErrorUserNotExist) {
			ResponseError(c, CodeUserExist)
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}

	// 3. 返回响应
	ResponseSuccess(c, nil)
}

// LoginHandler
// @Summary 用户登录
// @Description 用户登录
// @Tags 用户相关接口
// @Accept application/json
// @Produce application/json
// @Param user body models.ParamLogin true "登陆参数：账号、密码（均必填）"
// @Success 200 {object} _ResponseUserLogin
// @Router /api/v1/login [post]
func LoginHandler(c *gin.Context) {
	// 1. 获取参数和参数校验
	p := new(models.ParamLogin)
	if err := c.ShouldBindJSON(p); err != nil {
		// JSON 参数请求有误，直接返回响应
		zap.L().Error("Login with invalid param: ", zap.Error(err))

		// 判断 errs 是不是 validator.ValidatorErrors 类型
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}

		// 返回带信息的错误 全局翻译器翻译后去除无用结构体名
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
	}
	// 2. 业务处理
	user, err := logic.Login(p)
	if err != nil {
		zap.L().Error("logic.login failed", zap.String("username", p.Username), zap.Error(err))
		if errors.Is(err, mysql.ErrorUserNotExist) {
			ResponseError(c, CodeUserNotExist)
			return
		}
		ResponseError(c, CodeInvalidPassword)
		return
	}

	// 3. 返回响应
	ResponseSuccess(c, gin.H{
		"user_id":       fmt.Sprintf("%d", user.UserID),
		"user_name":     user.Username,
		"access_token":  user.Token,
		"refresh_token": user.RefreshToken,
	})
}

// RefreshHandler
// @Summary 刷新令牌
// @Description 用 refresh token 换发新的 access + refresh 双令牌（轮换机制）
// @Tags 用户相关接口
// @Accept application/json
// @Produce application/json
// @Param body body models.ParamRefresh true "刷新令牌参数"
// @Success 200 {object} _ResponseUserLogin
// @Router /api/v1/refresh [post]
func RefreshHandler(c *gin.Context) {
	// 1. 获取参数和参数校验
	p := new(models.ParamRefresh)
	if err := c.ShouldBindJSON(p); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 2. 解析并校验 refresh token（必须是 refresh 类型）
	mc, err := jwt.ParseToken(p.RefreshToken)
	if err != nil || mc.TokenType != jwt.TokenTypeRefresh {
		ResponseError(c, CodeInvalidToken)
		return
	}

	// 3. 换发新双令牌（校验 Redis 一致性，防重放 + 单端登录）
	accessToken, refreshToken, err := logic.RefreshTokens(mc.UserID, mc.Username, p.RefreshToken)
	if err != nil {
		ResponseError(c, CodeInvalidToken)
		return
	}

	// 4. 返回响应
	ResponseSuccess(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// LogoutHandler
// @Summary 用户登出
// @Description 用户登出，撤销 Redis 中保存的登录令牌，使其立即失效
// @Tags 用户相关接口
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer 用户 token 令牌"
// @Security ApiKeyAuth
// @Success 200 {object} _ResponseCommon
// @Router /api/v1/logout [post]
func LogoutHandler(c *gin.Context) {
	// 1. 获取当前登录用户
	userID, err := GetCurrentUser(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	// 2. 删除 Redis 中保存的登录令牌（幂等，重复登出也返回成功）
	if err := logic.Logout(userID); err != nil {
		zap.L().Error("logic.Logout failed", zap.Int64("user_id", userID), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3. 返回成功
	ResponseSuccess(c, nil)
}
