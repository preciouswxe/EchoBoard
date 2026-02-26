package models

import "time"

// Go 内存对齐

type Post struct {
	ID          int64     `json:"id,string" db:"post_id"`
	AuthorID    int64     `json:"author_id" db:"author_id"`
	CommunityID int64     `json:"community_id" db:"community_id" binding:"required"`
	Status      int32     `json:"status" db:"status"`
	Title       string    `json:"title" db:"title" binding:"required"`
	Content     string    `json:"content" db:"content" binding:"required"`
	CreateTime  time.Time `json:"create_time" db:"create_time"`

	LikeNum    int64 `json:"like_num" db:"like_num"`
	CollectNum int64 `json:"collect_num" db:"collect_num"`
	CommentNum int64 `json:"comment_num" db:"comment_num"`
}

// ApiPostDetail 帖子详情接口的结构体
type ApiPostDetail struct {
	AuthorName       string             `json:"author_name"`
	VoteNum          int64              `json:"vote_num"`
	IsLiked          bool               `json:"is_liked"` // 当前用户是否已点赞
	IsCollected      bool               `json:"is_collected"` // 当前用户是否已收藏
	*Post                               // 嵌入帖子结构体
	*CommunityDetail `json:"community"` // 嵌入社区信息
}

/*
	Request structs
*/

type PostCreateRequest struct {
	CommunityID int64  `json:"community_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
}

/*
	Response structs
*/
