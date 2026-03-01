package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/preciouswxe/EchoBoard_backend/pkg/es"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/controller"
	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"github.com/preciouswxe/EchoBoard_backend/logger"
	"github.com/preciouswxe/EchoBoard_backend/pkg/snowflake"
	"github.com/preciouswxe/EchoBoard_backend/router"
	"github.com/preciouswxe/EchoBoard_backend/setting"
)

// Go Web 开发较通用的脚手架模板

// @title EchoBoard
// @version 1.0
// @description EchoBoard 是一个轻量级、高性能的用户论坛系统，支持用户注册、登录和投票功能。

// @contact.name Lifridom
// @contact.url https://github.com/preciouswxe
// @contact.email wxe0750@qq.com

// @host 127.0.0.1:8081
// @BasePath /

func main() {
	// 1. 加载配置文件
	if err := setting.Init(); err != nil {
		fmt.Printf("init setting failed, err:%v\n", err)
		return
	}

	// 2. 初始化日志
	if err := logger.Init(setting.Conf.LogConfig, setting.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	// 把缓冲区的日志加进来
	defer zap.L().Sync()

	// 3. 初始化MySQL连接
	if err := mysql.Init(setting.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	defer mysql.Close()

	// 4. 初始化Redis连接
	if err := redis.Init(setting.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close()

	// 5. 初始化ES连接
	if err := es.Init(setting.Conf.EsConfig); err != nil {
		fmt.Printf("init elasticsearch failed, err:%v\n", err)
		return
	}

	// 6. 加载雪花算法
	if err := snowflake.Init(setting.Conf.AppConfig.StartTime, setting.Conf.AppConfig.MachineID); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}

	// 7. 初始化 gin 框架内置的校验器使用的翻译器
	if err := controller.InitTrans("zh"); err != nil {
		fmt.Printf("init trans failed, err:%v\n", err)
		return
	}

	// 8. 注册路由
	r := router.SetupRouter(setting.Conf.Mode)

	// 9. 启动服务（优雅关机）
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetString("app.host"), viper.GetInt("app.port")),
		Handler: r,
	}

	go func() {
		// 开启一个 goroutine 启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("listen: %s\n", zap.Error(err)) // 不是因为服务器关闭导致的错误
		}
	}()

	// 等待中断信号来优雅地关闭服务器，为关闭服务器操作设置一个 5 秒的超时
	// 创建一个接收信号的通道
	quit := make(chan os.Signal, 1)
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的 Ctrl+C 就是触发系统 SIGINT 信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify 把收到的 syscall.SIGINT或 syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此，当接收到上述两种信号时才会往下执行
	zap.L().Info("Shutdown Server ...")

	// 创建一个 5 秒超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 5 秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown ", zap.Error(err))
	}

	zap.L().Info("Server exiting")

}
