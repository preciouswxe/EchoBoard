package logic

import (
	"math"
	"time"
)

const (
	wilsonZ       = 1.96 // 95%置信区间：我有 95% 把握，这篇帖子真实点赞率至少是多少
	halfLifeHours = 72.0
)

/*
  10 赞 / 10 曝光：表面 100%，但样本太少，Wilson 分不会很高
  100 赞 / 120 曝光：样本多且质量高，Wilson 分很高
  100 赞 / 1000 曝光：点赞率低，Wilson 分较低
*/

// TimeDecay 返回帖子随发布时间衰减后的系数。
// 72 小时后衰减为原来的一半。

func WilsonScore(likes, views int64) float64 {
	if views <= 0 || likes <= 0 {
		return 0
	}
	if likes > views {
		likes = views
	}
	n := float64(views)
	p := float64(likes) / n
	z2 := wilsonZ * wilsonZ
	return (p + z2/(2*n) - wilsonZ*math.Sqrt((p*(1-p)+z2/(4*n))/n)) / (1 + z2/n)
}

func TimeDecay(createTime time.Time, now time.Time) float64 {
	ageHours := now.Sub(createTime).Hours()
	if ageHours < 0 {
		ageHours = 0
	}
	return math.Exp(-math.Ln2 * ageHours / halfLifeHours)
}

func HotScore(likes, views, collects, comments int64, createTime time.Time, now time.Time) float64 {
	wilson := WilsonScore(likes, views)

	// 收藏、评论代表更深度的互动，但不混进 Wilson 的点赞率。
	bonus := 0.15*math.Log1p(float64(collects)) + 0.10*math.Log1p(float64(comments))

	// 0.05 保证零互动的新帖仍有冷启动展示机会。
	return (0.05 + wilson + bonus) * TimeDecay(createTime, now)
}
