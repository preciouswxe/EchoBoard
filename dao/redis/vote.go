package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"math"
	"time"
)

// 本项目使用简化版的投票分数
// 投一票就加 432 分 = 一天 86400秒/200  意味着 200 张赞成票可以给你的帖子续一天，来自《redis实战》
// 书里设计的评分公式大概是： 帖子得分 = 投票数 × 432 + 发表时间戳 （新帖子时间戳大）

/*
投票的几种情况：
	direction=1时，有两种情况：
		1. 之前没有投过票，现在投赞成票	--> 更新分数和投票纪录 差值的绝对值:1 +432
		2. 之前投反对票，现在改投赞成票	--> 更新分数和投票纪录 差值的绝对值:2 +432*2
	direction=0时，有两种情况：
		1. 之前投过赞成票，现在要取消投票 --> 更新分数和投票纪录 差值的绝对值:1 -432
		2. 之前投过反对票，现在要取消投票 --> 更新分数和投票纪录 差值的绝对值:1 +432
	direction=-1时，有两种情况：
		1. 之前没有投过票，现在投赞成票	--> 更新分数和投票纪录 差值的绝对值:1 -432
		2. 之前投赞成票，现在改投反对票	--> 更新分数和投票纪录 差值的绝对值:2 -432*2

投票的限制：
	每个帖子自发表之日起，一个星期之内允许用户投票，超过则不允许再投票。 可以持久化到数据库里
	 1. 到期之后将 redis 中保存的赞成票数及反对票数存储到 mysql 表中
	 2. 到期之后删除那个 KeyPostVotedZSetPF
*/

const (
	oneWeekInSeconds = 7 * 24 * 3600
	scorePerVote     = 432 // 每一票分数值
)

var (
	ErrVoteTimeExpire = errors.New("投票时间已过")
	ErrVoteRepeated   = errors.New("不允许重复投票")
)

// CreatePost 加入创建帖子的时间和分数
func CreatePost(postID int64) error {
	ctx := context.Background()

	pipeline := client.TxPipeline()
	// 以下两个属于同一事务，要么都完成要么都不完成
	// 帖子时间
	pipeline.ZAdd(ctx, getRedisKey(KeyPostTimeZSet), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	// 帖子分数
	pipeline.ZAdd(ctx, getRedisKey(KeyPostScoreZSet), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	// 执行事务
	_, err := pipeline.Exec(ctx)
	return err
}

// VoteForPost 为帖子投票
func VoteForPost(userID, postID string, value float64) error {
	ctx := context.Background()

	// 1.判断投票限制
	// 去 redis 取帖子发布时间 (判定是否还在一周内)
	postTime := client.ZScore(ctx, getRedisKey(KeyPostTimeZSet), postID).Val()
	if float64(time.Now().Unix())-postTime > oneWeekInSeconds {
		return ErrVoteTimeExpire
	}

	// 2.更新帖子的分数
	// 先查询当前用户对当前帖子的投票记录
	ovalue := client.ZScore(ctx, getRedisKey(KeyPostVotedZSetPF+postID), userID).Val()

	// 如果这一次投票的值和之前保存的值一致，就提示不允许重复
	if value == ovalue {
		return ErrVoteRepeated
	}

	var dir float64
	if value > ovalue {
		dir = 1
	} else {
		dir = -1
	}
	// 计算两次投票的差值
	diff := math.Abs(ovalue - value)

	// 更新和记录需要放到同一个 pipeline 事务中操作
	// 启动事务
	pipeline := client.TxPipeline()

	// 更新分数
	pipeline.ZIncrBy(ctx, getRedisKey(KeyPostScoreZSet), dir*diff*scorePerVote, postID)

	// 3.记录用户为该帖子投票的数据
	if value == 0 {
		pipeline.ZRem(ctx, getRedisKey(KeyPostVotedZSetPF+postID), userID).Result()
	} else {
		pipeline.ZAdd(ctx, getRedisKey(KeyPostVotedZSetPF+postID), &redis.Z{
			Score:  value,
			Member: userID,
		})
	}

	_, err := pipeline.Exec(ctx)
	return err
}
