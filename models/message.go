package models

// PostInteractionEvent 消息队列 kafka 发布的 Message Value部分的结构
type PostInteractionEvent struct {
	Action string `json:"action"`  // 点赞与收藏操作："like", "unlike", "collect", "cancel_collect"
	PostID int64 `json:"post_id"`
	UserID int64 `json:"user_id"`
	TimeStamp int64 `json:"timestamp"`
}