package mysql

import (
	"github.com/preciouswxe/EchoBoard_backend/models"
	"github.com/preciouswxe/EchoBoard_backend/pkg/snowflake"
	"go.uber.org/zap"
)

// CreateComment 创建对应帖子的某用户的评论
func CreateComment(c *models.Comment) (err error) {
	c.CommentID = snowflake.GenID() // 评论 id 也由雪花算法生成

	tx, err := db.Begin()
	if err != nil {
		zap.L().Error("db.Begin failed", zap.Error(err))
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	sqlInsert := `insert into post_comment(comment_id, post_id, user_id, parent_id, content) values(?,?,?,?,?)`
	_, err = tx.Exec(sqlInsert, c.CommentID, c.PostID, c.UserID, c.ParentID, c.Content)
	if err != nil {
		zap.L().Error("tx.Exec CreateComment failed", zap.Error(err))
		return
	}

	sqlUpdate := `update post set comment_num = comment_num + 1 where post_id = ?`
	_, err = tx.Exec(sqlUpdate, c.PostID)
	if err != nil {
		zap.L().Error("tx.Exec update comment_num failed", zap.Error(err))
		return
	}

	zap.L().Info("CreateComment success")
	return
}

// GetCommentListByPostID 获取指定帖子的评论区
func GetCommentListByPostID(postID int64) (comments []*models.Comment, err error) {
	sqlStr := `select pc.comment_id, pc.post_id, pc.user_id, pc.parent_id, pc.content, pc.create_time ,
       u.username
	from post_comment pc
	left join user u on u.user_id = pc.user_id
	where post_id = ? and status = 1
	order by create_time ASC`
	comments = []*models.Comment{}
	err = db.Select(&comments, sqlStr, postID)
	if err != nil {
		zap.L().Error("db.Select GetCommentListByPostID failed", zap.Error(err))
	}
	return
}