package middlewares

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/preciouswxe/EchoBoard_backend/controller"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"github.com/preciouswxe/EchoBoard_backend/pkg/jwt"
)

// JWTAuthMiddleware 基于JWT的认证中间件
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 客户端携带 Token 有三种方式 1.放在请求头 2.放在请求体 3.放在 URI
		// 这里假设 Token 放在 Header 的 Authorization 中，并使用 Bearer 开头
		// Authorization: Bearer xxxx.xxx.xx
		// 这里的具体实现方式要依据你的实际业务情况决定
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

		// parts[1] 是获取到的 tokenString，我们使用之前定义好的解析 JWT 的函数来解析它
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		// Redis 中取出对应 UserID 的 token ,校验 token 是否一致（防止异地多端登录）
		key := "login_token:" + strconv.FormatInt(mc.UserID, 10)
		redistoken, err := redis.GetClient().Get(c, key).Result()
		if err != nil {
			controller.ResponseError(c, controller.CodeInvalidToken)
			c.Abort()
			return
		}

		//fmt.Printf("Authorization header: %s\n", authHeader)
		//fmt.Printf("Parsed UserID: %d, Token from header: %s\n", mc.UserID, parts[1])
		//fmt.Printf("Token from Redis: %s\n", redistoken)

		if redistoken != parts[1] {
			controller.ResponseError(c, controller.CodeNotSameDevice)
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
// 支持 “异地登录提醒”
// 支持强制下线 / 踢人
// 支持刷新 token 机制（access + refresh） liwenzhou也说了
