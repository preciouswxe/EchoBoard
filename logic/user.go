package logic

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/spf13/viper"

	"github.com/preciouswxe/EchoBoard_backend/dao/mysql"
	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"github.com/preciouswxe/EchoBoard_backend/pkg/jwt"
	"github.com/preciouswxe/EchoBoard_backend/pkg/snowflake"
)

// SignUp 用户注册
func SignUp(p *models.ParamSignUp) (err error) {
	// 1. 判断用户存不存在
	if err = mysql.CheckUserExist(p.Username); err != nil {
		return err
	}

	// 2. 生成 UID
	userID := snowflake.GenID()
	// 构造一个 User 实例
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}

	// 3. 保存进数据库
	err = mysql.InsertUser(user)

	return err
}

// Login 用户登录
func Login(p *models.ParamLogin) (user *models.User, err error) {
	user = &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	// 传递的是指针，就能拿到 user.UserID
	if err := mysql.Login(user); err != nil {
		return nil, err
	}

	// 生成 access + refresh 双令牌
	accessToken, refreshToken, err := jwt.GenTokenPair(user.UserID, user.Username)
	if err != nil {
		return nil, err
	}

	// 存入 redis, 覆盖之前的 refresh token（单端登录关键）
	key := "refresh_token:" + strconv.FormatInt(user.UserID, 10)
	err = redis.GetClient().Set(
		context.Background(),
		key,
		refreshToken,
		time.Duration(viper.GetInt("auth.refresh_expire"))*time.Hour,
	).Err()
	if err != nil {
		return nil, err
	}

	user.Token = accessToken
	user.RefreshToken = refreshToken
	return user, nil
}

// RefreshTokens 刷新令牌：校验 refresh token 与 Redis 中保存的一致后，换发新的双令牌
func RefreshTokens(userID int64, username, oldRefresh string) (accessToken, refreshToken string, err error) {
	key := "refresh_token:" + strconv.FormatInt(userID, 10)
	stored, err := redis.GetClient().Get(context.Background(), key).Result()
	if err != nil {
		return "", "", err
	}
	// 防重放：refresh token 只能换发一次，与 Redis 中保存的不一致则拒绝
	if stored != oldRefresh {
		return "", "", errors.New("refresh token 无效")
	}
	accessToken, refreshToken, err = jwt.GenTokenPair(userID, username)
	if err != nil {
		return "", "", err
	}
	// 轮换：更新 Redis 中的 refresh token，旧 refresh 立即失效
	err = redis.GetClient().Set(
		context.Background(),
		key,
		refreshToken,
		time.Duration(viper.GetInt("auth.refresh_expire"))*time.Hour,
	).Err()
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// Logout 用户退出
func Logout(userID int64) error {
	key := "refresh_token:" + strconv.FormatInt(userID, 10)
	return redis.GetClient().Del(context.Background(), key).Err()
}
