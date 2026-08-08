package oss

import (
	"fmt"
	"io"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/preciouswxe/EchoBoard_backend/setting"
	"go.uber.org/zap"
)

var (
	client       *aliyunoss.Client
	bucket       *aliyunoss.Bucket
	publicPrefix string // 对外访问前缀，如 https://echoboard.oss-cn-hangzhou.aliyuncs.com
)

func Init(cfg *setting.OssConfig) (err error) {
	client, err = aliyunoss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		zap.L().Error("oss.New failed", zap.Error(err))
		return err
	}
	bucket, err = client.Bucket(cfg.Bucket)
	if err != nil {
		zap.L().Error("client.Bucket failed", zap.Error(err))
		return err
	}

	if cfg.PublicDomain != "" {
		publicPrefix = "https://" + cfg.PublicDomain
	} else {
		publicPrefix = fmt.Sprintf("https://%s.%s", cfg.Bucket, cfg.Endpoint)
	}
	zap.L().Info("OSS init success", zap.String("prefix", publicPrefix))
	return nil
}

// Upload 上传到 OSS，objectKey 如 uploads/123456789012345.jpg，
// 返回可直接访问的完整 URL。bucket 需开放公共读。
func Upload(objectKey string, reader io.Reader) (url string, err error) {
	if bucket == nil {
		return "", fmt.Errorf("oss bucket not init")
	}
	if err = bucket.PutObject(objectKey, reader); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", publicPrefix, objectKey), nil
}
