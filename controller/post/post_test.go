package post

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/preciouswxe/EchoBoard_backend/controller"

	"github.com/stretchr/testify/assert"
)

func TestCreatePostHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	url := "/api/v1/post"
	r.POST(url, CreatePostHandler)

	body := `{
		"community_id": 1,
		"title": "test",
		"content": "just a test ~"
	}`

	// 发送请求 body要转换成 io.reader
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(body)))

	// 处理请求
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	// 判断响应的内容是不是按预期返回了需要登陆的错误
	// 方法1：判断响应内容是不是包含指定字符串
	assert.Contains(t, w.Body.String(), "需要登录")

	// 方法2：将响应的内容反序列化到ResponsData 然后判断字段与预期是否一致
	res := new(controller.ResponseData)
	if err := json.Unmarshal(w.Body.Bytes(), res); err != nil {
		t.Fatalf("json.Unmarshal w.Body failed, err : %v\n", err)
	}
	assert.Equal(t, res.Code, controller.CodeNeedLogin)
}
