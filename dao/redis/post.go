package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/preciouswxe/EchoBoard_backend/models"
)

// GetPostIDsInOrder 按序从 redis 获取帖子 ID
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

// GetPostVoteData 根据 ids 查询每篇帖子的投赞成票的数据
func GetPostVoteData(ids []string) (data []int64, err error) {
	ctx := context.Background()
	//data = make([]int64, 0, len(ids))
	//for _, id := range ids {
	//	// 获取键名
	//	key := getRedisKey(KeyPostVotedZSetPF + id)
	//	// 查找 key 中分数是 1 的元素数量 -> 统计每篇帖子的赞成票的数量
	//	v := client.ZCount(ctx, key, "1", "1").Val()
	//	data = append(data, v)
	//}

	// 使用 pipeline 一次发送多条命令减少 RTT
	pipeline := client.Pipeline()
	for _, id := range ids {
		key := getRedisKey(KeyPostVotedZSetPF + id)
		pipeline.ZCount(ctx, key, "1", "1")
	}
	cmders, err := pipeline.Exec(ctx)
	if err != nil {
		return
	}
	data = make([]int64, 0, len(ids))
	for _, cmder := range cmders {
		// 转换成 int 类型
		v := cmder.(*redis.IntCmd).Val()
		data = append(data, v)
	}

	return
}
