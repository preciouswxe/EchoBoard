package logic

import (
	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"github.com/preciouswxe/EchoBoard_backend/pkg/snowflake"
	"go.uber.org/zap"
)

func CreatePost(p *models.Post) (err error) {
	// 生成 post_id
	p.ID = snowflake.GenID()
	// 保存到数据库 返回
	return mysql.CreatePost(p)
}

func GetPostById(postID int64) (data *models.ApiPostDetail, err error) {
	// 获取帖子详情
	post, err := mysql.GetPostById(postID)
	if err != nil {
		zap.L().Error("mysql.GetPostById(postID) failed",
			zap.Int64("postID", postID),
			zap.Error(err),
		)
		return
	}

	// 根据作者 id 查询作者信息
	user, err := mysql.GetUserById(post.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
			zap.Int64("author_id", post.AuthorID),
			zap.Error(err),
		)
		return
	}

	// 根据社区 id 查询社区详细信息
	commnunityDetail, err := mysql.GetCommunityDetailByID(post.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed",
			zap.Int64("community_id", post.CommunityID),
			zap.Error(err),
		)
		return
	}

	// 拼接并返回数据
	data = &models.ApiPostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: commnunityDetail,
	}
	return
}
