package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

var mySecret = []byte("夏天夏天悄悄过去")

const (
	// TokenTypeAccess 短效访问令牌
	TokenTypeAccess = "access"
	// TokenTypeRefresh 长效刷新令牌
	TokenTypeRefresh = "refresh"
)

// MyClaims 自定义声明结构体并内嵌 jwt.StandardClaims
// jwt 包自带的 jwt.StandardClaims 只包含了官方字段
// 我们这里需要额外记录 字段，所以要自定义结构体
// 如果想要保存更多信息，都可以添加到这个结构体中
type MyClaims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"` // access / refresh，区分两种令牌
	JTI       string `json:"jti"`        // JWT ID，唯一标识，后续可用于黑名单撤销
	jwt.StandardClaims
}

// newJTI 生成随机的 jti，用作令牌唯一标识
func newJTI() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}

// genToken 生成指定类型、指定有效期的令牌
func genToken(userID int64, username, tokenType string, expire time.Duration) (string, error) {
	// 创建一个我们自己的声明的数据
	c := MyClaims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		JTI:       newJTI(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expire).Unix(),
			Issuer:    "github.com/preciouswxe/EchoBoard_backend",
		},
	}
	// 使用指定的签名方法先创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的 secret 真正执行签名，并获得完整的编码后的字符串 token
	return token.SignedString(mySecret)
}

// GenTokenPair 生成 access + refresh 双令牌
// access 短效（auth.access_expire，分钟），refresh 长效（auth.refresh_expire，小时）
func GenTokenPair(userID int64, username string) (accessToken, refreshToken string, err error) {
	accessToken, err = genToken(userID, username, TokenTypeAccess,
		time.Duration(viper.GetInt("auth.access_expire"))*time.Minute)
	if err != nil {
		return "", "", err
	}
	refreshToken, err = genToken(userID, username, TokenTypeRefresh,
		time.Duration(viper.GetInt("auth.refresh_expire"))*time.Hour)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func ParseToken(tokenString string) (*MyClaims, error) {
	// 解析 token
	var mc = new(MyClaims)
	token, err := jwt.ParseWithClaims(tokenString, mc, func(token *jwt.Token) (interface{}, error) {
		return mySecret, nil
	})
	if err != nil {
		return nil, err
	}
	// 校验 token
	if token.Valid {
		return mc, nil
	}
	return nil, errors.New("invalid token")
}
