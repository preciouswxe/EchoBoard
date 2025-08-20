package mysql

import (
	"github.com/jmoiron/sqlx"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"strings"
)

// CreatePost 创建帖子
func CreatePost(p *models.Post) (err error) {
	sqlStr := `insert into post(
	post_id, title, content, author_id, community_id)
	values (?,?,?,?,?)`

	_, err = db.Exec(sqlStr, p.ID, p.Title, p.Content, p.AuthorID, p.CommunityID)
	return
}

// GetPostById 根据帖子 ID 获取单个帖子数据
func GetPostById(postID int64) (post *models.Post, err error) {
	post = new(models.Post)
	sqlStr := `select 
    post_id, title, content, author_id, community_id, status, create_time
	from post
	where post_id = ?`

	err = db.Get(post, sqlStr, postID)
	return
}

// GetPostList 查询帖子列表函数 (以时间排序 从新到旧返回)
func GetPostList(page, size int64) (posts []*models.Post, err error) {
	sqlStr := `select 
    post_id, title, content, author_id, community_id, status, create_time
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
	sqlStr := `select post_id, title, content, author_id, community_id, create_time
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
