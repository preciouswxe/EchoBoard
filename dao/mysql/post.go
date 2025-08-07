package mysql

import "github.com/preciouswxe/EchoBoard_backend/models"

func CreatePost(p *models.Post) (err error) {
	sqlStr := `insert into post(
	post_id, title, content, author_id, community_id)
	values (?,?,?,?,?)`

	_, err = db.Exec(sqlStr, p.ID, p.Title, p.Content, p.AuthorID, p.CommunityID)
	return
}

func GetPostById(postID int64) (post *models.Post, err error) {
	post = new(models.Post)
	sqlStr := `select 
    post_id, title, content, author_id, community_id, status, create_time
	from post
	where post_id = ?`

	err = db.Get(post, sqlStr, postID)
	return
}
