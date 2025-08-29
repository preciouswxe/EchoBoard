package router

import (
	"github.com/gin-gonic/gin"
	"github.com/preciouswxe/EchoBoard_backend/controller"
	"github.com/preciouswxe/EchoBoard_backend/logger"
	"github.com/preciouswxe/EchoBoard_backend/middlewares"
	"net/http"
	"time"

	"github.com/gin-contrib/pprof"
	_ "github.com/preciouswxe/EchoBoard_backend/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) // 设置成发布模式
	}

	r := gin.New()

	r.Use(logger.GinLogger(), logger.GinRecovery(true), middlewares.RateLimitMiddleware(2*time.Second, 10))

	// 加载静态文件
	r.LoadHTMLFiles("./templates/index.html")
	// 让 html 里的引用都映射到根目录下的 static
	r.Static("/static", "./static")
	// 前端展示页
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// 后端路由组
	v1 := r.Group("/api/v1")
	// 注册
	v1.POST("/signup", controller.SignUpHandler)
	// 登录
	v1.POST("/login", controller.LoginHandler)

	// 开启 JWT 认证
	v1.Use(middlewares.JWTAuthMiddleware())
	{
		v1.GET("/community", controller.CommunityHandler)
		v1.GET("/community/:id", controller.CommunityDetailHandler)

		v1.POST("/post", controller.CreatePostHandler)
		v1.GET("/post/:id", controller.GetPostDetailHandler)
		v1.GET("/posts", controller.GetPostListHandler)

		// 根据时间或分数获取帖子列表
		v1.GET("/posts2", controller.GetPostListHandler2)

		v1.POST("/vote", controller.PostVoteController)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册 pprof 相关路由
	pprof.Register(r)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})

	return r
}
