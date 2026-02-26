package es

import (
	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"go.uber.org/zap"
)

// SyncPostsToES 将 MySQL 中所有帖子同步到 ES（一次性使用）
func SyncPostsToES() {
	page := int64(1)
	size := int64(100)  // 每批100条

	for {
		posts, err := mysql.GetPostList(page, size)
		if err != nil {
			zap.L().Error("SyncPostsToES mysql.GetPostList failed", zap.Error(err))
			return
		}

		if len(posts) == 0 {
			break  // 没有更多数据了
		}

		for _, post := range posts {
			user, err := mysql.GetUserById(post.AuthorID)
			if err != nil {
				zap.L().Error("SyncPostsToES GetUserById failed", zap.Error(err))
				continue
			}
			commnunityDetail, err := mysql.GetCommunityDetailByID(post.CommunityID)
			if err != nil {
				zap.L().Error("mysql.GetCommunityDetailByID failed",
					zap.Int64("community_id", post.CommunityID),
					zap.Error(err),
				)
				return
			}
			if err := UpsertPost(post, user.Username, commnunityDetail.Name); err != nil {
				zap.L().Error("SyncPostsToES UpsertPost failed", zap.Error(err))
				continue
			}
		}

		zap.L().Info("SyncPostsToES batch done", zap.Int64("page", page))
		page++
	}

	zap.L().Info("SyncPostsToES all done")
}
