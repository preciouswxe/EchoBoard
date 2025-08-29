package controller

type ResCode int64

const (
	CodeSuccess ResCode = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeServerBusy

	CodeNeedLogin
	CodeInvalidToken
	CodeNotSameDevice
	CodeVoteTimeOut
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:         "操作成功",
	CodeInvalidParam:    "请求参数错误",
	CodeUserExist:       "用户名已存在",
	CodeUserNotExist:    "用户名不存在",
	CodeInvalidPassword: "用户名或密码错误",
	CodeServerBusy:      "服务繁忙",

	CodeNeedLogin:     "需要登录",
	CodeInvalidToken:  "无效的token",
	CodeNotSameDevice: "账号已在其他设备登陆",
	CodeVoteTimeOut:   "投票已经结束",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	// 不存在该类型错误, 默认返回
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
