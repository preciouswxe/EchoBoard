package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/preciouswxe/EchoBoard_backend/setting"
)

var (
	client *redis.Client
	Nil    = redis.Nil
)

func Init(cfg *setting.RedisConfig) (err error) {
	client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.Host,
			cfg.Port,
		),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// redis v8 正确用法：传 context
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to initialize Redis client: %v", err)
	}
	return err
}

func Close() {
	// 不暴露 db 变量，用函数
	_ = client.Close()
}

func GetClient() *redis.Client {
	return client
}
