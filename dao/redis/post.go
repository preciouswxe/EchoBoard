package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/models"
)

// CreatePost 加入创建帖子的时间和分数
func CreatePost(postID, communityID int64) error {
	ctx := context.Background()
	postIDStr := strconv.FormatInt(postID, 10) // 转换为字符串

	pipeline := client.TxPipeline()
	// 以下属于同一事务，要么都完成要么都不完成
	// 1. 添加到时间排序 ZSet
	pipeline.ZAdd(ctx, getRedisKey(KeyPostTimeZSet), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	// 2. 添加到热度排序 ZSet（初始分数 = 0）
	pipeline.ZAdd(ctx, getRedisKey(KeyPostScoreZSet), &redis.Z{
		Score:  0,
		Member: postID,
	})

	// 3. 初始化帖子统计数据（Hash）: 点赞数、收藏数、评论数
	statsKey := getRedisKey(KeyPostStatsPF + postIDStr)
	pipeline.HSet(ctx, statsKey, map[string]interface{}{
		"like_num": 0,
		"collect_num": 0,
		"comment_num": 0,
	})

	// 4. 补充：把帖子 id 加到社区的 set
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(communityID)))
	pipeline.SAdd(ctx, cKey, postIDStr)

	// 执行事务
	_, err := pipeline.Exec(ctx)
	if err != nil {
		zap.L().Error("CreatePost in Redis failed",
			zap.Int64("post_id", postID),
			zap.Int64("community_id", communityID),
			zap.Error(err),
		)
		return err
	}
	return nil
}


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

// LikePost 点赞帖子（先写入redis）
func LikePost(userID, postID int64) (err error){
	ctx := context.Background()
	postIDStr := strconv.FormatInt(postID, 10)
	userIDStr := strconv.FormatInt(userID, 10)

	likeKey := getRedisKey(KeyPostLikedSetPF + postIDStr)
	statsKey := getRedisKey(KeyPostStatsPF + postIDStr)

	// 检查是否已经点赞
	isLiked, err := client.SIsMember(ctx, likeKey, userIDStr).Result()
	if err != nil {
		zap.L().Error("LikePost in Redis (client.SIsMember) failed",
			zap.Int64("post_id", postID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return err
	}
	if isLiked {
		zap.L().Warn("LikePost: already liked", zap.Int64("post_id", postID), zap.Int64("user_id", userID))
		return
	}

	pipeline := client.TxPipeline()
	// 1. 添加到点赞集合
	pipeline.SAdd(ctx, likeKey, userIDStr)
	// 2. 点赞数 +1
	pipeline.HIncrBy(ctx, statsKey, "like_num", 1)
	// 3. 更新热度分数（每个点赞 +3 分）
	pipeline.ZIncrBy(ctx, getRedisKey(KeyPostScoreZSet), 3, postIDStr)

	_, err = pipeline.Exec(ctx)
	if err != nil {
		zap.L().Error("LikePost in Redis failed",
			zap.Int64("post_id", postID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return err
	}
	return nil

	// TODO: [消息队列] 发送点赞事件到 MQ，异步持久化到 MySQL
	// 消息内容: {user_id, post_id, action: "like", timestamp}
	// 接入步骤:
	//   1. 引入 RabbitMQ/Kafka 客户端
	//   2. 定义消息体: type InteractionEvent struct {...}
	//   3. 发送: mq.Publish("post_interaction", event)
	//   4. 消费者: 从队列取消息 -> 写入 MySQL post_like 表
}

// UnlikePost 取消点赞帖子（先写入redis）
func UnlikePost(userID, postID int64) (err error){
	ctx := context.Background()
	postIDStr := strconv.FormatInt(postID, 10)
	userIDStr := strconv.FormatInt(userID, 10)

	likeKey := getRedisKey(KeyPostLikedSetPF + postIDStr)
	statsKey := getRedisKey(KeyPostStatsPF + postIDStr)

	// 检查是否已经点赞
	isLiked, err := client.SIsMember(ctx, likeKey, userIDStr).Result()
	if err != nil {
		zap.L().Error("UnlikePost in Redis (client.SIsMember) failed",
			zap.Int64("post_id", postID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return err
	}
	if !isLiked {
		zap.L().Warn("UnlikePost: have not liked", zap.Int64("post_id", postID), zap.Int64("user_id", userID))
		return
	}

	pipeline := client.TxPipeline()
	// 1. 移出点赞集合
	pipeline.SRem(ctx, likeKey, userIDStr)
	// 2. 点赞数 +1
	pipeline.HIncrBy(ctx, statsKey, "like_num", -1)
	// 3. 更新热度分数（每个点赞 +3 分）
	pipeline.ZIncrBy(ctx, getRedisKey(KeyPostScoreZSet), -3, postIDStr)

	_, err = pipeline.Exec(ctx)
	if err != nil {
		zap.L().Error("UnlikePost in Redis failed",
			zap.Int64("post_id", postID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return err
	}
	return nil

	// TODO: [消息队列] 发送取消点赞事件到 MQ
}

// IsUserLikedPost 检查用户是否已经点赞
func IsUserLikedPost(userID, postID int64) (bool, error) {
	ctx := context.Background()
	postIDStr := strconv.FormatInt(postID, 10)
	userIDStr := strconv.FormatInt(postID, 10)

	likeKey := getRedisKey(KeyPostLikedSetPF + postIDStr)
	return client.SIsMember(ctx, likeKey, userIDStr).Result()
}

// CollectPost 用户收藏帖子
func CollectPost(userID, postID int64) (err error) {
	ctx := context.Background()
	postIDStr := strconv.FormatInt(postID, 10)
	userIDStr := strconv.FormatInt(userID, 10)

	// Key: post:collect:{post_id} 存储收藏该帖子的用户集合
	collectKey := getRedisKey(KeyPostCollectedSetPF + postIDStr)
	statsKey := getRedisKey(KeyPostStatsPF + postIDStr)

	// 检查是否已收藏
	isCollected, err := client.SIsMember(ctx, collectKey, userIDStr).Result()
	if err != nil {
		zap.L().Error("CollectPost in Redis (client.SIsMember) failed",
			zap.Int64("post_id", postID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return err
	}
	if isCollected {
		zap.L().Warn("CollectPost: already collected", zap.Int64("post_id", postID), zap.Int64("user_id", userID))
		return
	}

	pipeline := client.TxPipeline()

	// 1. 添加到收藏集合
	pipeline.SAdd(ctx, collectKey, userIDStr)
	// 2. 收藏数 +1
	pipeline.HIncrBy(ctx, statsKey, "collect_num", 1)
	// 3. 更新热度分数（收藏权重更高：+5 分）
	pipeline.ZIncrBy(ctx, getRedisKey(KeyPostScoreZSet), 5, postIDStr)

	_, err = pipeline.Exec(ctx)
	if err != nil {
		zap.L().Error("CollectPost failed",
			zap.Int64("user_id", userID),
			zap.Int64("post_id", postID),
			zap.Error(err),
		)
		return err
	}

	// TODO: [消息队列] 发送收藏事件到 MQ，异步持久化到 MySQL

	return nil
}

// CancelCollectPost 用户取消收藏
func CancelCollectPost(userID, postID int64) (err error) {
	ctx := context.Background()
	postIDStr := strconv.FormatInt(postID, 10)
	userIDStr := strconv.FormatInt(userID, 10)

	collectKey := getRedisKey(KeyPostCollectedSetPF + postIDStr)
	statsKey := getRedisKey(KeyPostStatsPF + postIDStr)

	// 检查是否已收藏
	isCollected, err := client.SIsMember(ctx, collectKey, userIDStr).Result()
	if err != nil {
		zap.L().Error("CancelCollectPost in Redis (client.SIsMember) failed",
			zap.Int64("post_id", postID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return err
	}
	if !isCollected {
		zap.L().Warn("CancelCollectPost: have not collected", zap.Int64("post_id", postID), zap.Int64("user_id", userID))
		return
	}

	pipeline := client.TxPipeline()
	// 1. 从收藏集合移除
	pipeline.SRem(ctx, collectKey, userIDStr)
	// 2. 收藏数 -1
	pipeline.HIncrBy(ctx, statsKey, "collect_num", -1)
	// 3. 更新热度分数
	pipeline.ZIncrBy(ctx, getRedisKey(KeyPostScoreZSet), -5, postIDStr)

	_, err = pipeline.Exec(ctx)
	if err != nil {
		zap.L().Error("CancelCollectPost failed",
			zap.Int64("user_id", userID),
			zap.Int64("post_id", postID),
			zap.Error(err),
		)
		return err
	}

	// TODO: [消息队列] 发送取消收藏事件到 MQ

	return nil
}

// IsUserCollectedPost 检查用户是否收藏了该帖子
func IsUserCollectedPost(userID, postID int64) (bool, error) {
	ctx := context.Background()
	postIDStr := strconv.FormatInt(postID, 10)
	userIDStr := strconv.FormatInt(userID, 10)

	collectKey := getRedisKey(KeyPostCollectedSetPF + postIDStr)
	return client.SIsMember(ctx, collectKey, userIDStr).Result()
}

// GetRecommendPostIDs 从Redis获取热门推荐（优先同社区）
func GetRecommendPostIDs(size int, excludeIDs []int64, communityID int64) ([]string, error) {
	ctx := context.Background()

	excludeMap := make(map[int64]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	var result []string

	// 1. 如果指定了社区，优先从同社区推荐（占70%）
	if communityID > 0 {
		communitySize := int(float64(size) * 0.7)
		if communitySize < 1 {
			communitySize = size / 2 // 至少推荐一半同社区
		}

		communityPosts, err := getRecommendFromCommunity(ctx, communityID, communitySize*2, excludeMap)
		if err != nil {
			zap.L().Warn("getRecommendFromCommunity failed", zap.Error(err))
		} else {
			// 添加同社区的推荐
			for _, idStr := range communityPosts {
				if len(result) >= communitySize {
					break
				}
				result = append(result, idStr)
			}
		}
	}

	// 2. 从全站热门补充剩余数量
	needCount := size - len(result)
	if needCount > 0 {
		// 更新排除列表（加上已推荐的同社区帖子）
		for _, idStr := range result {
			id, _ := strconv.ParseInt(idStr, 10, 64)
			excludeMap[id] = true
		}

		globalPosts, err := getRecommendFromGlobal(ctx, needCount*2, excludeMap)
		if err != nil {
			zap.L().Error("getRecommendFromGlobal failed", zap.Error(err))
		} else {
			result = append(result, globalPosts...)
		}
	}

	// 3. 确保不超过请求数量
	if len(result) > size {
		result = result[:size]
	}

	return result, nil
}

// getRecommendFromCommunity 从指定社区获取热门帖子
func getRecommendFromCommunity(ctx context.Context, communityID int64, fetchSize int, excludeMap map[int64]bool) ([]string, error) {
	// 使用 zinterstore 获取某个社区的热门帖子（和 GetCommunityPostIDsInOrder 逻辑类似）
	scoreKey := getRedisKey(KeyPostScoreZSet)
	communityKey := getRedisKey(KeyCommunitySetPF + strconv.FormatInt(communityID, 10))

	// 临时 key 用于存储交集结果
	tempKey := getRedisKey("temp:recommend:community:" + strconv.FormatInt(communityID, 10))

	// 计算交集（社区帖子 ∩ 热度排序）
	_, err := client.ZInterStore(ctx, tempKey, &redis.ZStore{
		Keys:      []string{communityKey, scoreKey},
		Aggregate: "MAX",
	}).Result()
	if err != nil {
		return nil, err
	}

	// 设置过期时间
	client.Expire(ctx, tempKey, 60*time.Second)

	// 获取热门帖子
	postIDs, err := client.ZRevRange(ctx, tempKey, 0, int64(fetchSize)).Result()
	if err != nil {
		return nil, err
	}

	// 过滤排除的ID
	var result []string
	for _, idStr := range postIDs {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		if !excludeMap[id] {
			result = append(result, idStr)
		}
	}

	return result, nil
}

// getRecommendFromGlobal 从全站获取热门帖子
func getRecommendFromGlobal(ctx context.Context, fetchSize int, excludeMap map[int64]bool) ([]string, error) {
	key := getRedisKey(KeyPostScoreZSet)

	postIDs, err := client.ZRevRange(ctx, key, 0, int64(fetchSize)).Result()
	if err != nil {
		return nil, err
	}

	var result []string
	for _, idStr := range postIDs {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		if !excludeMap[id] {
			result = append(result, idStr)
		}
	}

	return result, nil
}