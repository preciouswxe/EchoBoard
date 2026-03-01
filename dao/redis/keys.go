package redis

// redis key 注意使用命名空间的方式，方便查询和拆分

const (
	Prefix             = "EchoBoard:"

	KeyPostTimeZSet    = "post:time"   // zset;帖子及发帖时间
	KeyPostScoreZSet   = "post:score"  // zset;帖子及投票的分数
	KeyPostStatsPF     = "post:stats:" // hash:帖子的点赞、收藏、评论数的 map
	// 点赞和收藏相关 key
	KeyPostLikedSetPF     = "post:like"     // set;点赞该帖子的用户集合 post:like:{post_id} -> Set(user_id)
	KeyPostCollectedSetPF = "post:collect"  // set;收藏该帖子的用户集合 post:collect:{post_id} -> Set(user_id)

	KeyCommunitySetPF = "community:" // set;保存每个话题分区下帖子的 id
)

// getRedisKey 给 redis key 加上前缀
func getRedisKey(key string) string {
	return Prefix + key
}
