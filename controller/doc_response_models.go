package controller

import (
	"github.com/preciouswxe/EchoBoard_backend/models"
)

// 专门用来放接口文档用到的 model
// 因为我们的接口文档返回的数据格式是一致的，但是具体的 data 类型不一致

type _ResponseCommon struct {
	Code    ResCode `json:"code"`    // 通用业务响应状态码
	Message string  `json:"message"` // 提示信息
}

type _ResponseUserLogin struct {
	Code    ResCode       `json:"code"`    // 通用业务响应状态码
	Message string        `json:"message"` // 提示信息
	Data    UserLoginData `json:"data"`    // 返回登陆的 token 等数据
}

type _ResponsePostList struct {
	Code    ResCode                 `json:"code"`    // 业务响应状态码
	Message string                  `json:"message"` // 提示信息
	Data    []*models.ApiPostDetail `json:"data"`    // 数据
}

type _ResponseCommunityAllList struct {
	Code    int                 `json:"code"`    // 业务响应状态码
	Message string              `json:"message"` // 提示信息
	Data    []*models.Community `json:"data"`    // 数据
}

type _ResponseCommunityDetailList struct {
	Code    int                       `json:"code"`    // 业务响应状态码
	Message string                    `json:"message"` // 提示信息
	Data    []*models.CommunityDetail `json:"data"`    // 数据
}

/*
	应用结构体
*/

type UserLoginData struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Token    string `json:"token"`
}
