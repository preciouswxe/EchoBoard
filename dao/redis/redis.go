package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/preciouswxe/EchoBoard_backend/setting"
)

var rdb *redis.Client

func Init(cfg *setting.RedisConfig) (err error) {
	rdb = redis.NewClient(&redis.Options{
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
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to initialize Redis client: %v", err)
	}
	return err
}

func Close() {
	// 不暴露 db 变量，用函数
	_ = rdb.Close()
}

func GetClient() *redis.Client {
	return rdb
}
