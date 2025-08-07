package logic

import (
	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"github.com/preciouswxe/EchoBoard_backend/pkg/snowflake"
)

func CreatePost(p *models.Post) (err error) {
	// 生成 post_id
	p.ID = snowflake.GenID()
	// 保存到数据库 返回
	return mysql.CreatePost(p)
}

func GetPostById(postID int64) (data *models.Post, err error) {
	return mysql.GetPostById(postID)
}
