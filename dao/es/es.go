package es

import (
	"context"
	"fmt"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/preciouswxe/EchoBoard_backend/models"
	"github.com/preciouswxe/EchoBoard_backend/pkg/es"
	"go.uber.org/zap"
)

const PostIndex = "posts"

// ESPost 关于 post 的 es document 格式定义
type ESPost struct {
	PostID        string `json:"post_id"`
	AuthorName    string `json:"author_name"`
	CommunityName string `json:"community_name"`
	Title         string `json:"title"`
	Content       string `json:"content"`
}

// UpsertPost 在创建帖子的时候把 mysql 数据更新并上传到 es
func UpsertPost(post *models.Post, authorName, communityName string) error {
	doc := ESPost{
		PostID:        strconv.FormatInt(post.ID, 10),
		AuthorName:    authorName,
		CommunityName: communityName,
		Title:         post.Title,
		Content:       post.Content,
	}
	resp, err := es.Client.Index(PostIndex).
		Id(strconv.FormatInt(post.ID, 10)).
		Document(doc).
		Do(context.Background())
	if err != nil {
		zap.L().Error("es upsert failed",
			zap.Int64("post_id", post.ID),
			zap.Error(err),
		)
		return fmt.Errorf("es upsert failed: %w", err)
	}
	zap.L().Info("es upsert success", zap.Any("resp.result", resp.Result), zap.Int64("postID", post.ID))
	return nil
}

// SearchPosts 根据关键词在 es 进行搜索，拿到 postID 返回
func SearchPosts(keyword string, page, size int64) (postIDs []int64, total int64, err error) {
	// 偏移量
	from := int(page-1) * int(size)
	s := int(size)

	res, err := es.Client.
		Search().
		Index(PostIndex).
		Request(&search.Request{
			From: &from,
			Size: &s,
			Query: &types.Query{
				Bool: &types.BoolQuery{ // 组合多个条件查询
					Must: []types.Query{
						{
							MultiMatch: &types.MultiMatchQuery{
								Query: keyword,
								Fields: []string{"title^3", "content", "author_name^2", "community_name^2"},
							},
						},
					},
				},
			},
			}).
		Do(context.Background())
	if err != nil {
		return nil, 0, fmt.Errorf("")
	}

	total = res.Hits.Total.Value
	for _, hit := range res.Hits.Hits {
		postID, convErr := strconv.ParseInt(*hit.Id_, 10, 64)
		if convErr != nil {
			zap.L().Info("strconv.ParseInt failed", zap.Error(err))
			continue
		}
		postIDs = append(postIDs, postID)
	}
	return
}

// Query类似于写 json
//{
//  "query": {
//    "bool": {
//      "must": [
//        {
//          "multi_match": {
//            "query": "关键词",
//            "fields": ["title^3", "content", "author_name^2"]
//          }
//        }
//      ]
//    }
//  }
//}