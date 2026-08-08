package controller

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/preciouswxe/EchoBoard_backend/pkg/oss"
	"github.com/preciouswxe/EchoBoard_backend/pkg/snowflake"
)

// 允许的媒体扩展名
var allowedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".mp4": true, ".webm": true, ".mov": true,
}

func mediaTypeByExt(ext string) string {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "image"
	case ".mp4", ".webm", ".mov":
		return "video"
	}
	return "other"
}

// UploadHandler 上传媒体文件（图片/视频）到 OSS，返回可直接访问的完整 URL。
// @Summary 上传媒体文件
// @Description 上传图片/视频到对象存储，返回可直接访问的 URL
// @Tags 媒体上传
// @Accept multipart/form-data
// @Produce application/json
// @Param Authorization header string true "Bearer 用户 token 令牌"
// @Param file formData file true "上传的文件（图片 ≤5MB，视频 ≤50MB）"
// @Security ApiKeyAuth
// @Success 200 {object} controller.ResponseData{data=object{url=string,media_type=string}}
// @Router /api/v1/upload [post]
func UploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParam, "缺少文件")
		return
	}

	// 1. 扩展名白名单 + 大小限制
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		ResponseErrorWithMsg(c, CodeInvalidParam, "不支持的文件类型")
		return
	}
	mediaType := mediaTypeByExt(ext)
	maxSize := int64(5 << 20) // 图片 5MB
	if mediaType == "video" {
		maxSize = 50 << 20 // 视频 50MB
	}
	if file.Size > maxSize {
		ResponseErrorWithMsg(c, CodeInvalidParam, "文件过大")
		return
	}

	src, err := file.Open()
	if err != nil {
		zap.L().Error("upload: open file failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	defer src.Close()

	// 2. 校验 magic bytes，防止伪装扩展名
	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	mime := http.DetectContentType(head[:n])
	if mediaType == "image" && !strings.HasPrefix(mime, "image/") {
		ResponseErrorWithMsg(c, CodeInvalidParam, "文件内容不是图片")
		return
	}
	if mediaType == "video" && !strings.HasPrefix(mime, "video/") {
		ResponseErrorWithMsg(c, CodeInvalidParam, "文件内容不是视频")
		return
	}

	// 3. 上传到 OSS（重新从头读）
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		zap.L().Error("upload: seek failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	objectKey := fmt.Sprintf("uploads/%d%s", snowflake.GenID(), ext)
	url, err := oss.Upload(objectKey, src)
	if err != nil {
		zap.L().Error("upload: oss upload failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, gin.H{
		"url":        url,
		"media_type": mediaType,
	})
}
