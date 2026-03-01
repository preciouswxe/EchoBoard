package mysql

import (
	"go.uber.org/zap"
)

// CollectPost 收藏帖子
func CollectPost(postID, userID int64) (err error) {
	// 开启事务，要么全部完成，要么部分失败就都回滚
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

	sqlInsert := `insert ignore into post_collect(post_id, user_id) values(?, ?)` // insert ignore: 遇到唯一索引冲突不报错，直接忽略
	result, err := tx.Exec(sqlInsert, postID, userID)
	if err != nil {
		zap.L().Error("tx.Exec CollectPost failed", zap.Error(err))
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		zap.L().Info("result.RowsAffected : found rows == 0")
		return
	}

	sqlUpdate := `update post set collect_num = collect_num + 1 where post_id = ?`
	_, err = tx.Exec(sqlUpdate, postID)
	if err != nil {
		zap.L().Error("tx.Exec(sqlUpdate, postID) failed", zap.Error(err))
		return
	}
	zap.L().Info("CollectPost success")
	return
}

// CancelCollectPost 取消收藏帖子
func CancelCollectPost(postID, userID int64) (err error) {
	// 开启事务，要么全部完成，要么部分失败就都回滚
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

	sqlDelete := `delete from post_collect where post_id = ? and user_id = ?`
	result, err := tx.Exec(sqlDelete, postID, userID)
	if err != nil {
		zap.L().Error("tx.Exec(sqlDelete, postID, userID) failed", zap.Error(err))
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		zap.L().Info("result.RowsAffected : found rows == 0")
		return
	}

	sqlUpdate := `update post set collect_num = GREATEST(collect_num-1, 0) where post_id = ?`
	_, err = tx.Exec(sqlUpdate, postID)
	if err != nil {
		zap.L().Error("tx.Exec CancelCollectPost failed", zap.Error(err))
		return
	}
	zap.L().Info("CancelCollectPost success")
	return
}