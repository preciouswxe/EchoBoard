package logic

import (
	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"go.uber.org/zap"
)

func getPostInfo(post *models.Post) (*models.User, *models.CommunityDetail, error) {
	// 根据作者 id 查询作者信息
	user, err := mysql.GetUserById(post.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
			zap.Int64("author_id", post.AuthorID),
			zap.Error(err),
		)
		return nil, nil, err
	}

	// 根据社区 id 查询社区详细信息
	commnunityDetail, err := mysql.GetCommunityDetailByID(post.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed",
			zap.Int64("community_id", post.CommunityID),
			zap.Error(err),
		)
		return nil, nil, err
	}

	return user, commnunityDetail, nil
}