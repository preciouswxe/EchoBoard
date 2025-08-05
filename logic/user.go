package logic

import (
	"context"
	"strconv"
	"time"

	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"bluebell/pkg/snowflake"
)

// SignUp 用户注册
func SignUp(p *models.ParamSignUp) (err error) {
	// 1. 判断用户存不存在
	if err = mysql.CheckUserExist(p.Username) ; err != nil {
		return err
	}

	// 2. 生成 UID
	userID := snowflake.GenID()
	// 构造一个 User 实例
	user := &models.User{
		UserID: userID,
		Username: p.Username,
		Password: p.Password,
	}

	// 3. 保存进数据库
	err = mysql.InsertUser(user)

	return err
}

// Login 用户登录
func Login(p *models.ParamLogin) (token string, err error) {
	user := &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	// 传递的是指针，就能拿到 user.UserID
	if err := mysql.Login(user); err != nil {
		return "", err
	}

	// 生成 JWT token
	token, err = jwt.GenToken(user.UserID, user.Username)
	if err != nil {
		return "", err
	}

	// 存入 redis, 覆盖之前的 token（单端登录关键）
	key := "login_token:" + strconv.FormatInt(user.UserID, 10)
	err = redis.GetClient().Set(
		context.Background(),
		key,
		token,
		time.Hour*1,
		).Err()
	if err != nil {
		return "", err
	}


	return token ,nil
}

// Logout 用户退出
func Logout(userID int64) error {
	key := "login_token:" + strconv.FormatInt(userID, 10)
	return redis.GetClient().Del(context.Background(), key).Err()
}