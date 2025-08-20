package redis

import (
	"context"
	"github.com/preciouswxe/EchoBoard_backend/models"
)

func GetPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	// 从 redis 获取 id, 默认时间
	// 1.根据用户请求中携带的 order 参数确定要查询的 rediskey
	key := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		key = getRedisKey(KeyPostScoreZSet)
	}

	// 2.确定查询的索引起始点
	start := (p.Page - 1) * p.Size
	end := start + p.Size - 1

	// 3. ZRevRange 按分数从大到小的顺序查询指定数量的元素
	ctx := context.Background()
	return client.ZRevRange(ctx, key, start, end).Result()
}
