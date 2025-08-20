package logic

import (
	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"github.com/preciouswxe/EchoBoard_backend/pkg/snowflake"
	"go.uber.org/zap"
)

func CreatePost(p *models.Post) (err error) {
	// 生成 post_id
	p.ID = snowflake.GenID()
	// 保存到数据库 返回
	err = mysql.CreatePost(p)
	if err != nil {
		return err
	}
	// redis 记录帖子创建时间
	err = redis.CreatePost(p.ID)
	return
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

func GetPostList(page, size int64) (data []*models.ApiPostDetail, err error) {
	// 获取一列帖子详情
	posts, err := mysql.GetPostList(page, size)
	if err != nil {
		return nil, err
	}

	// 初始化
	data = make([]*models.ApiPostDetail, 0, len(posts))

	// 列表需要循环读出具体信息
	for _, post := range posts {
		// 根据作者 id 查询作者信息
		user, err := mysql.GetUserById(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err),
			)
			continue
		}

		// 根据社区 id 查询社区详细信息
		commnunityDetail, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed",
				zap.Int64("community_id", post.CommunityID),
				zap.Error(err),
			)
			continue
		}

		// 拼接
		postdetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			Post:            post,
			CommunityDetail: commnunityDetail,
		}

		// 添加到返回里
		data = append(data, postdetail)
	}

	return
}

func GetPostList2(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 去 redis 查询 id 列表
	ids, err := redis.GetPostIDsInOrder(p)
	if err != nil {
		return
	}
	if len(ids) == 0 {
		zap.L().Warn("redis.GetPostIDsInOrder(p) return 0 data")
		return
	}

	// 根据 id 去数据库查询帖子详细信息
	// 返回的数据还要按照给定的 id 的顺序返回
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		return
	}

	// 提前查询好每篇帖子的 vote 数
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		return
	}

	// 列表需要循环读出具体信息
	for idx, post := range posts {
		// 根据作者 id 查询作者信息
		user, err := mysql.GetUserById(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err),
			)
			continue
		}

		// 根据社区 id 查询社区详细信息
		commnunityDetail, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed",
				zap.Int64("community_id", post.CommunityID),
				zap.Error(err),
			)
			continue
		}

		// 拼接
		postdetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: commnunityDetail,
		}

		// 添加到返回里
		data = append(data, postdetail)
	}

	return
}
