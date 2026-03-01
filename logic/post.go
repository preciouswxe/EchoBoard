package logic

import (
	"strconv"

	dao_es "github.com/preciouswxe/EchoBoard_backend/dao/es"
	dao_redis "github.com/preciouswxe/EchoBoard_backend/dao/redis"
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
	// redis 记录帖子创建信息
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

	isLiked, isCollected, err := mysql.GetPostUserRelation(post.ID, userID) // 注意这里传入的是当前用户ID，不是作者ID
	if err != nil {
		zap.L().Error("mysql.GetPostUserRelation failed", zap.Error(err))
	}

	// 拼接并返回数据
	data = &models.ApiPostDetail{
		AuthorName:      user.Username,
		IsLiked:         isLiked,
		IsCollected:     isCollected,
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

// GetPostList2 进阶版 - 获取帖子列表
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


	// 列表需要循环读出具体信息
	for _, post := range posts {
		user, communityDetail, err := getPostInfo(post)
		if err != nil {
			zap.L().Error("getPostInfo failed", zap.Error(err))
			return nil, err
		}

		isLiked, isCollected, err := mysql.GetPostUserRelation(post.ID, userID) // 注意这里传入的是当前用户ID，不是作者ID
		if err != nil {
			zap.L().Error("mysql.GetPostUserRelation failed", zap.Error(err))
		}

		// 拼接
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			IsLiked:         isLiked,
			IsCollected:     isCollected,
			Post:            post,
			CommunityDetail: communityDetail,
		}

		data = append(data, postDetail)
	}

	return
}

// GetCommunityPostList 根据社区返回帖子列表
func GetCommunityPostList(p *models.ParamPostList, userID int64) (data []*models.ApiPostDetail, err error) {
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


	// 列表需要循环读出具体信息
	for _, post := range posts {
		user, communityDetail, err := getPostInfo(post)
		if err != nil {
			zap.L().Error("getPostInfo failed", zap.Error(err))
			return nil, err
		}

		isLiked, isCollected, err := mysql.GetPostUserRelation(post.ID, userID) // 注意这里传入的是当前用户ID，不是作者ID
		if err != nil {
			zap.L().Error("mysql.GetPostUserRelation failed", zap.Error(err))
		}

		// 拼接
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			IsLiked:         isLiked,
			IsCollected:     isCollected,
			Post:            post,
			CommunityDetail: communityDetail,
		}

		data = append(data, postDetail)
	}

	return
}

// GetPostListNew 将两个查询逻辑合二为一
func GetPostListNew(p *models.ParamPostList, userID int64) (data []*models.ApiPostDetail, err error) {
	// 根据请求参数的不同，执行不同的逻辑
	if p.CommunityID == 0 {
		// 说明查询所有
		data, err = GetPostList2(p, userID)
	} else {
		// 根据社区 id 查询
		data, err = GetCommunityPostList(p, userID)
	}

	if err != nil {
		zap.L().Error("GetPostListNew failed", zap.Error(err))
		return nil, err
	}

	return
}

// GetPostListBySearch 用户输入关键词获取搜索结果 - 接 ES
func GetPostListBySearch(p *models.ParamSearchPostList, userID int64) (searchData []*models.ApiPostDetail, recommendData []*models.ApiPostDetail, total int64, err error) {
	// 1. 走 ES 搜索
	postIDs, total, err := dao_es.SearchPosts(p.KeyWord, p.Page, p.Size)
	if err != nil {
		zap.L().Error("dao_es.SearchPosts failed", zap.String("keyword", p.KeyWord), zap.Error(err))
		return nil, nil, 0, err
	}

	// 2. 处理搜索结果
	searchData = make([]*models.ApiPostDetail, 0)
	var mainCommunityID int64 = 0 // 用于推荐

	if len(postIDs) > 0 {
		searchData, err = getPostDetailsByIDs(postIDs, userID)
		if err != nil {
			zap.L().Error("getPostDetailsByIDs for search failed", zap.Error(err))
			return nil, nil, total, err
		}

		// 获取第一个搜索结果的社区ID（用于推荐同社区内容）
		if len(searchData) > 0 && searchData[0].CommunityDetail != nil {
			mainCommunityID = searchData[0].CommunityDetail.ID
		}
	}

	// 3. 获取推荐（优先推荐同社区）
	recommendData = make([]*models.ApiPostDetail, 0)
	recommendIDStrs, err := dao_redis.GetRecommendPostIDs(8, postIDs, mainCommunityID)
	if err != nil {
		zap.L().Warn("dao_redis.GetRecommendPostIDs failed", zap.Error(err))
		return searchData, recommendData, total, nil
	}

	// 转换为int64
	var recommendIDs []int64
	for _, idStr := range recommendIDStrs {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		recommendIDs = append(recommendIDs, id)
	}

	if len(recommendIDs) > 0 {
		recommendData, err = getPostDetailsByIDs(recommendIDs, userID)
		if err != nil {
			zap.L().Error("getPostDetailsByIDs for recommend failed", zap.Error(err))
		}
	}

	return searchData, recommendData, total, nil
}

// getPostDetailsByIDs 根据帖子ID列表获取详细信息（复用逻辑）
func getPostDetailsByIDs(postIDs []int64, userID int64) ([]*models.ApiPostDetail, error) {
	if len(postIDs) == 0 {
		zap.L().Info("len(postIDs)==0, return make([]*models.ApiPostDetail, 0), nil")
		return make([]*models.ApiPostDetail, 0), nil
	}
	// int64 转 string
	ids := make([]string, 0, len(postIDs))
	for _, id := range postIDs {
		ids = append(ids, strconv.FormatInt(id, 10))
	}
	// 从 MySQL 获取帖子信息
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		zap.L().Error("mysql.GetPostListByIDs failed", zap.Error(err))
		return nil, err
	}

	data := make([]*models.ApiPostDetail, 0, len(posts))
	for _, post := range posts {
		user, communityDetail, err := getPostInfo(post)
		if err != nil {
			zap.L().Error("getPostInfo failed", zap.Error(err))
			return nil, err
		}

		isLiked, isCollected, err := mysql.GetPostUserRelation(post.ID, userID)
		if err != nil {
			zap.L().Error("mysql.GetPostUserRelation failed", zap.Error(err))
		}

		// 拼接
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			Post:            post,
			IsLiked:         isLiked,
			IsCollected:     isCollected,
			CommunityDetail: communityDetail,
		}

		data = append(data, postDetail)
	}

	return data, nil
}

// GetPostListBySearchByLike 纯 mysql like搜索 [停用]
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