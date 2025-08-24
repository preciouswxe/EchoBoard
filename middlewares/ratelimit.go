package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"github.com/juju/ratelimit"
)

func RateLimitMiddleware(fillInterval time.Duration, cap int64) func(c *gin.Context) {
	bucket := ratelimit.NewBucket(fillInterval, cap)
	return func(c *gin.Context) {
		// 如果令牌不足就抛弃该请求，返回服务繁忙
		if bucket.TakeAvailable(1) == 0 {
			c.String(http.StatusTooManyRequests, "服务请求繁忙……请稍后再试。")
			c.Abort()
			return
		}
		// 有令牌则放行
		c.Next()
	}
}
