package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/preciouswxe/EchoBoard_backend/controller"
	"github.com/preciouswxe/EchoBoard_backend/pkg/jwt"
)

// JWTAuthMiddleware 基于JWT的认证中间件
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 客户端携带 Token 有三种方式 1.放在请求头 2.放在请求体 3.放在 URI
		// 这里假设 Token 放在 Header 的 Authorization 中，并使用 Bearer 开头
		// Authorization: Bearer xxxx.xxx.xx
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			controller.ResponseError(c, controller.CodeNeedLogin)
			c.Abort()
			return
		}

		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		// parts[1] 是 access token，解析它
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}
		// 只允许 access token 访问受保护接口；refresh token 只能走 /refresh
		if mc.TokenType != jwt.TokenTypeAccess {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		// 将当前请求的 userID 信息保存到请求的上下文 c 上
		c.Set(controller.CtxUserIDKey, mc.UserID)

		// 后续的处理函数可以用过 c.Get(CtxUserIDKey) 来获取当前请求的用户信息
		c.Next()
	}
}

// ✅ 可选增强（未来可以加）
// 记录登录设备信息（IP、UA）
// 支持 "异地登录提醒"
// 支持强制下线 / 踢人（Redis token 黑名单）
