package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"strconv"
	"time"
)

// getIDsFromKey 包内用于计算索引起始点
func getIDsFromKey(key string, page, size int64) ([]string, error) {
	start := (page - 1) * size
	end := start + size - 1

	// ZRevRange 按分数从大到小的顺序查询指定数量的元素
	ctx := context.Background()
	return client.ZRevRange(ctx, key, start, end).Result()
}

// GetPostIDsInOrder 按序从 redis 获取帖子 ID
func GetPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	// 从 redis 获取 id, 默认时间
	// 1.根据用户请求中携带的 order 参数确定要查询的 rediskey
	key := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		key = getRedisKey(KeyPostScoreZSet)
	}

	// 2.确定查询的索引起始点
	return getIDsFromKey(key, p.Page, p.Size)
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

// GetCommunityPostIDsInOrder 按社区查询 ids
// 在 Redis 里动态生成「某个社区下的时间/分数排序帖子列表」，并做缓存（60 秒），然后分页返回帖子 ID。
func GetCommunityPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	ctx := context.Background()

	orderKey := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		orderKey = getRedisKey(KeyPostScoreZSet)
	}

	// 使用 zinterstore 把分区的帖子 set 与 帖子分数的 zset 生成一个新的 zset
	// 针对新的 zset 按之前逻辑取数据

	// 社区的 key
	ckey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(p.CommunityID)))

	// 利用缓存 key 减少 zinterstore 执行的次数
	key := orderKey + strconv.Itoa(int(p.CommunityID))

	// 判断缓存是否存在
	if client.Exists(ctx, key).Val() < 1 {
		// 不存在该 key 需要计算
		pipeline := client.Pipeline()
		// zinterstore 聚合时选择两边最大值
		pipeline.ZInterStore(ctx, key, &redis.ZStore{
			Keys:      []string{ckey, orderKey}, // 交集：社区帖子 + 全局排序
			Aggregate: "MAX",                    // 分数取最大（其实就是沿用全局的分数）
		})
		// 设置超时时间
		pipeline.Expire(ctx, key, 60*time.Second)
		// 拼接完后执行
		_, err := pipeline.Exec(ctx)
		if err != nil {
			return nil, err
		}
	}

	// 存在的话直接根据 key 查询 ids
	return getIDsFromKey(key, p.Page, p.Size)
}
