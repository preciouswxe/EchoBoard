package mysql

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/models"
)

// CreatePost 创建帖子
func CreatePost(p *models.Post) (err error) {
	sqlStr := `insert into post(
	post_id, title, content, media, author_id, community_id)
	values (?,?,?,?,?,?)`

	_, err = db.Exec(sqlStr, p.ID, p.Title, p.Content, p.Media, p.AuthorID, p.CommunityID)
	return
}

// GetPostById 根据帖子 ID 获取单个帖子数据
func GetPostById(postID int64) (post *models.Post, err error) {
	post = new(models.Post)
	sqlStr := `select 
    post_id, title, content, media, author_id, community_id, status, like_num, collect_num, comment_num, create_time
	from post
	where post_id = ?`

	err = db.Get(post, sqlStr, postID)
	return
}

// GetPostList 查询帖子列表函数 (以时间排序 从新到旧返回)
func GetPostList(page, size int64) (posts []*models.Post, err error) {
	sqlStr := `select 
    post_id, title, content, media, author_id, community_id, status, like_num, collect_num, comment_num, create_time
	from post
	ORDER BY create_time
	DESC
	LIMIT ?,?`

	posts = make([]*models.Post, 0, 2)
	err = db.Select(&posts, sqlStr, (page-1)*size, size)
	return
}

// GetPostListByIDs 根据给定的 id 列表查询帖子数据
func GetPostListByIDs(ids []string) (postList []*models.Post, err error) {
	sqlStr := `select 
    post_id, title, content, media, author_id, community_id, like_num, collect_num, comment_num, create_time
	from post
	where post_id in (?)
	order by  FIND_IN_SET(post_id, ?)
	`
	// 前一个问号传入 ids, 后一个问号传入 ids 拼接的 string
	query, args, err := sqlx.In(sqlStr, ids, strings.Join(ids, ","))
	if err != nil {
		return nil, err
	}

	// 根据数据库类型调整占位符
	query = db.Rebind(query)

	err = db.Select(&postList, query, args...)
	return
}

// GetPostListByKeyWord 根据用户输入关键词搜索相关帖子（可优化推荐）
func GetPostListByKeyWord(keyWord string, page, size int64) (postList []*models.Post, err error) {
	sqlStr := `select 
	post_id, title, content, author_id, community_id, like_num, collect_num, comment_num, create_time
	from post
	where title like ?
	or content like ?
	limit ?, ?`
	postList = make([]*models.Post, 0, 2)
	err = db.Select(&postList, sqlStr, "%"+keyWord+"%", "%"+keyWord+"%", (page-1)*size, size)
	return
}

// GetPostUserRelation 获取帖子与当前用户的点赞、收藏、评论关系 [停用，改用redis]
func GetPostUserRelation(postID, userID int64) (isLiked, isCollected bool, err error) {
	var count int

	sqlLike := `select count(1) from post_like where post_id = ? and user_id = ?`
	if err = db.Get(&count, sqlLike, postID, userID); err != nil {
		return
	}
	if count > 1 {
		// 只可能有一条记录，不是则说明出问题了
		zap.L().Error("post_like data error: duplicate record", zap.Int64("postID", postID), zap.Int64("userID", userID))
	}
	isLiked = count == 1

	sqlCollect := `select count(1) from post_collect where post_id = ? and user_id = ?`
	if err = db.Get(&count, sqlCollect, postID, userID); err != nil {
		return
	}
	if count > 1 {
		// 只可能有一条记录，不是则说明出问题了
		zap.L().Error("post_collect data error: duplicate record", zap.Int64("postID", postID), zap.Int64("userID", userID))
	}
	isCollected = count == 1

	return
}