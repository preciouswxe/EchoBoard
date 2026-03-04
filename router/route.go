package router

import (
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/preciouswxe/EchoBoard_backend/controller"
	postController "github.com/preciouswxe/EchoBoard_backend/controller/post"
	_ "github.com/preciouswxe/EchoBoard_backend/docs"
	"github.com/preciouswxe/EchoBoard_backend/logger"
	"github.com/preciouswxe/EchoBoard_backend/middlewares"
)

func SetupRouter(mode string) *gin.Engine {
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) // 设置成发布模式
	}

	r := gin.New()
	// 使用中间件
	//r.Use(logger.GinLogger(), logger.GinRecovery(true), middlewares.RateLimitMiddleware(2*time.Second, 10))
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	// 加载静态文件
	//r.LoadHTMLFiles("./templates/index.html")
	// 让 html 里的引用都映射到根目录下的 static
	//r.Static("/static", "./static")
	// 前端展示页
	//r.GET("/", func(c *gin.Context) {
	//	c.HTML(http.StatusOK, "index.html", nil)
	//})

	// 后端路由组
	v1 := r.Group("/api/v1")
	// 注册
	v1.POST("/signup", controller.SignUpHandler)
	// 登录
	v1.POST("/login", controller.LoginHandler)

	// 开启 JWT 认证
	v1.Use(middlewares.JWTAuthMiddleware())
	{
		// 1. 话题（社区）相关
		v1.GET("/community", controller.CommunityHandler)
		v1.GET("/community/:id", controller.CommunityDetailHandler)

		// 2. 帖子相关
		v1.POST("/post", postController.CreatePostHandler)
		v1.GET("/post/:id", postController.GetPostDetailHandler)
		//v1.GET("/posts", postController.GetPostListHandler)   // 暂时不用
		v1.GET("/posts2", postController.GetPostListHandler2) // 根据时间或分数获取帖子列表
		v1.GET("/posts/search", postController.GetPostsBySearchHandler)
		// 点赞
		v1.POST("/post/:id/like", postController.LikePostHandler)
		v1.DELETE("/post/:id/like", postController.UnlikePostHandler)
		// 收藏
		v1.POST("/post/:id/collect", postController.CollectPostHandler)
		v1.DELETE("/post/:id/collect", postController.CancelCollectPostHandler)
		// 评论
		v1.POST("/post/:id/comment", postController.CreateCommentHandler)
		v1.GET("/post/:id/comment", postController.GetCommentListHandler)
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
