package logic

import (
	"strconv"

	dao_es "github.com/preciouswxe/EchoBoard_backend/dao/es"
	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"github.com/preciouswxe/EchoBoard_backend/pkg/snowflake"
	"go.uber.org/zap"
)

func CreatePost(p *models.Post) (err error) {
	// 生成 post_id
	p.ID = snowflake.GenID()
	// 保存到数据库
	err = mysql.CreatePost(p)
	if err != nil {
		return err
	}
	// redis 记录帖子创建时间
	err = redis.CreatePost(p.ID, p.CommunityID)
	// 异步写入 es
	go func(postCopy *models.Post) {
		user, err := mysql.GetUserById(p.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById failed", zap.Error(err))
			return
		}
		commnunityDetail, err := mysql.GetCommunityDetailByID(p.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID failed",
				zap.Int64("community_id", p.CommunityID),
				zap.Error(err),
			)
			return
		}

		if err := dao_es.UpsertPost(p, user.Username, commnunityDetail.Name); err != nil {
			zap.L().Error("dao_es.UpsertPost failed", zap.Error(err))
		}
	}(p)
	return
}

// GetPostById 根据帖子 id 和用户 id 查询帖子详情数据
func GetPostById(postID, userID int64) (data *models.ApiPostDetail, err error) {
	// 获取帖子详情
	post, err := mysql.GetPostById(postID)
	if err != nil {
		zap.L().Error("mysql.GetPostById(postID) failed",
			zap.Int64("postID", postID),
			zap.Error(err),
		)
		return
	}

	user, communityDetail, err := getPostInfo(post)
	if err != nil {
		zap.L().Error("getPostInfo failed", zap.Error(err))
		return nil, err
	}

	// 拼接并返回数据
	data = &models.ApiPostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: communityDetail,
	}
	return
}

// GetPostList 获取帖子列表
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
		user, communityDetail, err := getPostInfo(post)
		if err != nil {
			zap.L().Error("getPostInfo failed", zap.Error(err))
			return nil, err
		}

		// 拼接
		postdetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			Post:            post,
			CommunityDetail: communityDetail,
		}

		// 添加到返回里
		data = append(data, postdetail)
	}

	return
}

func GetPostList2(p *models.ParamPostList, userID int64) (data []*models.ApiPostDetail, err error) {
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
		user, communityDetail, err := getPostInfo(post)
		if err != nil {
			zap.L().Error("getPostInfo failed", zap.Error(err))
			return nil, err
		}

		// 拼接
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: communityDetail,
		}

		data = append(data, postDetail)
	}

	return
}

func GetCommunityPostList(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 去 redis 查询 id 列表
	ids, err := redis.GetCommunityPostIDsInOrder(p)
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
		user, communityDetail, err := getPostInfo(post)
		if err != nil {
			zap.L().Error("getPostInfo failed", zap.Error(err))
			return nil, err
		}

		// 拼接
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: communityDetail,
		}

		data = append(data, postDetail)
	}

	return
}

// GetPostListNew 将两个查询逻辑合二为一
func GetPostListNew(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 根据请求参数的不同，执行不同的逻辑
	if p.CommunityID == 0 {
		// 说明查询所有
		data, err = GetPostList2(p)
	} else {
		// 根据社区 id 查询
		data, err = GetCommunityPostList(p)
	}

	if err != nil {
		zap.L().Error("GetPostListNew failed", zap.Error(err))
		return nil, err
	}

	return
}

// GetPostListBySearch 用户输入关键词获取搜索结果
func GetPostListBySearch(p *models.ParamSearchPostList) (data []*models.ApiPostDetail, err error) {
	// 走 ES 搜索
	postIDs, _, err := dao_es.SearchPosts(p.KeyWord, p.Page, p.Size)
	if err != nil {
		zap.L().Error("dao_es.SearchPosts failed", zap.String("keyword", p.KeyWord), zap.Error(err))
		return nil, err
	}

	if len(postIDs)==0 {
		zap.L().Info("len(postIDs)==0, return make([]*models.ApiPostDetail, 0), nil")
		return make([]*models.ApiPostDetail, 0), nil
	}

	// int64 转 string
	ids := make([]string, 0, len(postIDs))
	for _, id := range postIDs {
		ids = append(ids, strconv.FormatInt(id, 10))
	}

	// 复用已有的 GetPostListByIDs
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		zap.L().Error("mysql.GetPostListByIDs failed",zap.Error(err))
		return nil, err
	}
	for _, post := range posts {
		user, communityDetail, err := getPostInfo(post)
		if err != nil {
			zap.L().Error("getPostInfo failed", zap.Error(err))
			return nil, err
		}
		// 拼接
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			Post:            post,
			CommunityDetail: communityDetail,
		}

		data = append(data, postDetail)
	}

	return
}

// GetPostListBySearchByLike 纯 mysql like搜索
func GetPostListBySearchByLike(p *models.ParamSearchPostList) (data []*models.ApiPostDetail, err error) {
	posts, err := mysql.GetPostListByKeyWord(p.KeyWord, p.Page, p.Size)
	if err != nil {
		zap.L().Error("mysql.GetPostListByKeyWord failed",
			zap.String("keyword", p.KeyWord),
			zap.Error(err),
		)
		return nil, err
	}

	if len(posts) == 0 {
		return make([]*models.ApiPostDetail, 0), nil
	}

	for _, post := range posts {
		user, communityDetail, err := getPostInfo(post)
		if err != nil {
			zap.L().Error("getPostInfo failed", zap.Error(err))
			return nil, err
		}

		// 拼接
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			Post:            post,
			CommunityDetail: communityDetail,
		}

		data = append(data, postDetail)
	}

	return
}