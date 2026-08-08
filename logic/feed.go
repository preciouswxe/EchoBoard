package logic

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/preciouswxe/EchoBoard_backend/dao/redis"
	"github.com/preciouswxe/EchoBoard_backend/models"
)

const feedCandidateLimit int64 = 500

type feedCursor struct {
	Score      float64 `json:"s"`
	PostID     int64   `json:"p"`
	SnapshotAt int64   `json:"t"`
}

type rankedFeedPost struct {
	PostID int64
	Score  float64
}

type FeedResult struct {
	List       []*models.ApiPostDetail `json:"list"`
	NextCursor string                  `json:"next_cursor"`
	HasMore    bool                    `json:"has_more"`
}

func GetFeed(userID int64, cursorValue string, size int) (*FeedResult, error) {
	if size <= 0 {
		size = 20
	}
	if size > 50 {
		size = 50
	}

	now := time.Now()
	var cursor *feedCursor
	if cursorValue != "" {
		decoded, err := decodeFeedCursor(cursorValue)
		if err != nil {
			return nil, err
		}
		cursor = &decoded
		now = time.Unix(cursor.SnapshotAt, 0)
	}

	candidates, err := redis.GetFeedCandidates(feedCandidateLimit)
	if err != nil {
		return nil, err
	}

	scores := make(map[int64]float64, len(candidates))
	ranked := make([]rankedFeedPost, 0, len(candidates))
	for _, candidate := range candidates {
		score := HotScore(candidate.LikeNum, candidate.ViewNum, candidate.CollectNum, candidate.CommentNum, candidate.CreateTime, now)
		scores[candidate.PostID] = score
		ranked = append(ranked, rankedFeedPost{PostID: candidate.PostID, Score: score})
	}
	if err := redis.UpdateFeedScores(scores); err != nil {
		return nil, err
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].PostID > ranked[j].PostID
		}
		return ranked[i].Score > ranked[j].Score
	})

	selected := make([]rankedFeedPost, 0, size+1)
	for _, item := range ranked {
		if cursor != nil && (item.Score > cursor.Score || (item.Score == cursor.Score && item.PostID >= cursor.PostID)) {
			continue
		}
		selected = append(selected, item)
		if len(selected) == size+1 {
			break
		}
	}

	hasMore := len(selected) > size
	if hasMore {
		selected = selected[:size]
	}
	postIDs := make([]int64, len(selected))
	for i, item := range selected {
		postIDs[i] = item.PostID
	}
	details, err := getPostDetailsByIDs(postIDs, userID)
	if err != nil {
		return nil, err
	}

	result := &FeedResult{List: details, HasMore: hasMore}
	if hasMore && len(selected) > 0 {
		last := selected[len(selected)-1]
		result.NextCursor, err = encodeFeedCursor(feedCursor{Score: last.Score, PostID: last.PostID, SnapshotAt: now.Unix()})
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func RecordFeedImpressions(postIDStrings []string) error {
	postIDs := make([]int64, 0, len(postIDStrings))
	seen := make(map[int64]struct{}, len(postIDStrings))
	for _, value := range postIDStrings {
		postID, err := strconv.ParseInt(value, 10, 64)
		if err != nil || postID <= 0 {
			return fmt.Errorf("invalid post id")
		}
		if _, exists := seen[postID]; exists {
			continue
		}
		seen[postID] = struct{}{}
		postIDs = append(postIDs, postID)
	}
	return redis.RecordFeedImpressions(postIDs)
}

func encodeFeedCursor(cursor feedCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeFeedCursor(value string) (feedCursor, error) {
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return feedCursor{}, fmt.Errorf("invalid feed cursor")
	}
	var cursor feedCursor
	if err := json.Unmarshal(payload, &cursor); err != nil || cursor.PostID <= 0 || cursor.SnapshotAt <= 0 {
		return feedCursor{}, fmt.Errorf("invalid feed cursor")
	}
	return cursor, nil
}
