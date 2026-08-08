package redis

import (
	"context"
	"errors"
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
		"view_num":    0,
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
// FeedCandidate contains the real-time values used to rank a Feed post.
type FeedCandidate struct {
	PostID     int64
	CreateTime time.Time
	LikeNum    int64
	CollectNum int64
	CommentNum int64
	ViewNum    int64
}

func GetFeedCandidates(limit int64) ([]FeedCandidate, error) {
	if limit <= 0 {
		return []FeedCandidate{}, nil
	}
	if limit > 500 {
		limit = 500
	}

	ctx := context.Background()
	end := limit - 1
	recent, err := client.ZRevRange(ctx, getRedisKey(KeyPostTimeZSet), 0, end).Result()
	if err != nil {
		return nil, err
	}
	hot, err := client.ZRevRange(ctx, getRedisKey(KeyPostScoreZSet), 0, end).Result()
	if err != nil {
		return nil, err
	}

	ids := make(map[string]struct{}, len(recent)+len(hot))
	for _, id := range recent {
		ids[id] = struct{}{}
	}
	for _, id := range hot {
		ids[id] = struct{}{}
	}

	pipe := client.Pipeline()
	createdCmds := make(map[string]*redis.FloatCmd, len(ids))
	statsCmds := make(map[string]*redis.StringStringMapCmd, len(ids))
	for id := range ids {
		createdCmds[id] = pipe.ZScore(ctx, getRedisKey(KeyPostTimeZSet), id)
		statsCmds[id] = pipe.HGetAll(ctx, getRedisKey(KeyPostStatsPF+id))
	}
	if _, err = pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	parseCount := func(stats map[string]string, field string) int64 {
		count, _ := strconv.ParseInt(stats[field], 10, 64)
		return count
	}
	result := make([]FeedCandidate, 0, len(ids))
	for id := range ids {
		postID, parseErr := strconv.ParseInt(id, 10, 64)
		if parseErr != nil {
			continue
		}
		createdAt, createdErr := createdCmds[id].Result()
		if createdErr != nil || createdAt <= 0 {
			continue
		}
		stats, _ := statsCmds[id].Result()
		result = append(result, FeedCandidate{
			PostID:     postID,
			CreateTime: time.Unix(int64(createdAt), 0),
			LikeNum:    parseCount(stats, "like_num"),
			CollectNum: parseCount(stats, "collect_num"),
			CommentNum: parseCount(stats, "comment_num"),
			ViewNum:    parseCount(stats, "view_num"),
		})
	}
	return result, nil
}

func UpdateFeedScores(scores map[int64]float64) error {
	if len(scores) == 0 {
		return nil
	}
	items := make([]*redis.Z, 0, len(scores))
	for postID, score := range scores {
		items = append(items, &redis.Z{Member: postID, Score: score})
	}
	return client.ZAdd(context.Background(), getRedisKey(KeyPostScoreZSet), items...).Err()
}

func RecordFeedImpressions(postIDs []int64) error {
	if len(postIDs) == 0 {
		return nil
	}
	ctx := context.Background()
	pipe := client.Pipeline()
	for _, postID := range postIDs {
		pipe.HIncrBy(ctx, getRedisKey(KeyPostStatsPF+strconv.FormatInt(postID, 10)), "view_num", 1)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func IncrementPostCommentCount(postID int64) error {
	return client.HIncrBy(context.Background(), getRedisKey(KeyPostStatsPF+strconv.FormatInt(postID, 10)), "comment_num", 1).Err()
}

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
}

// GetPostLikeCount 获取帖子点赞数
func GetPostLikeCount(postID int64) (int64, error) {
	ctx := context.Background()
	postIDStr := strconv.FormatInt(postID, 10)
	key := getRedisKey(KeyPostLikedSetPF + postIDStr)

	count, err := client.SCard(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil // key 不存在返回 0
	}
	return count, err
}

// IsUserLikedPost 检查用户是否已经点赞
func IsUserLikedPost(userID, postID int64) (bool, error) {
	ctx := context.Background()
	postIDStr := strconv.FormatInt(postID, 10)
	userIDStr := strconv.FormatInt(userID, 10)

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

	return nil
}

// GetPostCollectCount 获取帖子收藏数
func GetPostCollectCount(postID int64) (int64, error) {
	ctx := context.Background()
	postIDStr := strconv.FormatInt(postID, 10)
	key := getRedisKey(KeyPostCollectedSetPF + postIDStr)

	count, err := client.SCard(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return count, err
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

type PostStatsAndRelation struct {
	LikeCount      int64
	CollectCount int64
	IsLiked        bool
	IsCollected     bool
}

// GetPostsStatsAndRelationBatch 批量获取多个帖子的统计数据和用户关系(用于列表查询优化)
// 使用 Redis Pipeline 一次性查询所有数据,避免 N+1 查询问题
func GetPostsStatsAndRelationBatch(postIDs []int64, userID int64) (map[int64]*PostStatsAndRelation, error) {
	if len(postIDs) == 0 {
		return make(map[int64]*PostStatsAndRelation), nil
	}

	ctx := context.Background()
	pipe := client.Pipeline()

	// 从 Hash 中批量查询每个帖子的点赞数和收藏数
	statsCmds := make(map[int64]*redis.StringStringMapCmd)
	for _, postID := range postIDs {
		postIDStr := strconv.FormatInt(postID, 10)
		statsKey := getRedisKey(KeyPostStatsPF + postIDStr)

		// HGetAll 获取整个 Hash（like_num, collect_num, comment_num）
		statsCmds[postID] = pipe.HGetAll(ctx, statsKey)
	}

	// 批量查询用户是否点赞/收藏了这些帖子
	isLikedCmds := make(map[int64]*redis.BoolCmd)
	isCollectedCmds := make(map[int64]*redis.BoolCmd)
	userIDStr := strconv.FormatInt(userID, 10)
	for _, postID := range postIDs {
		postIDStr := strconv.FormatInt(postID, 10)
		likeKey := getRedisKey(KeyPostLikedSetPF + postIDStr)
		collectKey := getRedisKey(KeyPostCollectedSetPF + postIDStr)

		isLikedCmds[postID] = pipe.SIsMember(ctx, likeKey, userIDStr)
		isCollectedCmds[postID] = pipe.SIsMember(ctx, collectKey, userIDStr)
	}

	// 执行所有命令 (一次性网络请求)
	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	// 组装结果
	result := make(map[int64]*PostStatsAndRelation)
	for _, postID := range postIDs {
		stats, _ := statsCmds[postID].Result()
		likeCount, _ := strconv.ParseInt(stats["like_num"], 10, 64)
		collectCount, _ := strconv.ParseInt(stats["collect_num"], 10, 64)

		isLiked, _ := isLikedCmds[postID].Result()
		isCollected, _ := isCollectedCmds[postID].Result()

		result[postID] = &PostStatsAndRelation{
			LikeCount:    likeCount,
			CollectCount: collectCount,
			IsLiked:      isLiked,
			IsCollected:  isCollected,
		}
	}
	return result, nil
}
