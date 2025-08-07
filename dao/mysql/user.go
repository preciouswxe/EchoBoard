package mysql

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"github.com/preciouswxe/EchoBoard_backend/models"
)

// 把每一步数据库操作封装成函数
// 等待 logic 层根据业务需求调用

const secret = "Lifridom"

// CheckUserExist 检查指定用户名的用户是否存在
func CheckUserExist(username string) error {
	sqlStr := `select count(user_id) from user where username = ?`
	var count int
	if err := db.Get(&count, sqlStr, username); err != nil {
		return err
	}
	if count > 0 {
		return ErrorUserExist
	}
	return nil
}

// InsertUser 插入一条新的用户
func InsertUser(user *models.User) (err error) {
	// 对密码进行加密
	user.Password = encryptPassword(user.Password)
	// 执行 SQL 语句入库
	sqlStr := `insert into user(user_id, username, password) values (?,?,?)`
	_, err = db.Exec(sqlStr, user.UserID, user.Username, user.Password)
	return
}

// encryptPassword 加密密码
func encryptPassword(oPassword string) string {
	h := md5.New()
	// 加盐
	h.Write([]byte(secret))
	// 返回成字符串
	return hex.EncodeToString(h.Sum([]byte(oPassword)))
}

// Login 用户登录
func Login(user *models.User) error {
	oPassword := user.Password
	// 查询对应记录是否存在
	sqlStr := `select user_id, username, password from user where username = ?`
	err := db.Get(user, sqlStr, user.Username)
	if err == sql.ErrNoRows {
		return ErrorUserNotExist
	}
	if err != nil {
		// 查询数据库失败
		return err
	}
	// 判断密码是否正确
	password := encryptPassword(oPassword)
	if password != user.Password {
		return ErrorInvalidPassword
	}
	return nil
}
