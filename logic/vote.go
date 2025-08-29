package logic

import (
	"errors"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"go.uber.org/zap"
	"strconv"
)

var (
	ErrVoteTimeExpire = errors.New("投票时间已过")
	ErrVoteRepeated   = errors.New("不允许重复投票")
)

// VoteForPost 为帖子投票的函数
func VoteForPost(userID int64, p *models.ParamVoteData) error {
	zap.L().Debug("VoteForPost",
		zap.Int64("userID", userID),
		zap.String("postID", p.PostID),
		zap.Int8("direction", p.Direction),
	)

	status, err := redis.VoteForPost(strconv.Itoa(int(userID)), p.PostID, float64(p.Direction))
	if err != nil {
		return err
	}

	// 根据 redis 返回来选择业务错误码
	switch status {
	case "time_expired":
		return ErrVoteTimeExpire
	case "repeated":
		return ErrVoteRepeated
	case "ok":
		return nil
	default:
		return errors.New("未知的投票状态")
	}
}
