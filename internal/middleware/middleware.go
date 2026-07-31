package middleware

import (
	"net/http"

	"e_commerce_platform/internal/response"

	"github.com/gin-gonic/gin"
)

const CtxUserID = "userID"
const CtxRequestID = "requestID"

// RequestID 注入 request_id（3.3）。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 读 X-Request-ID 或生成 uuid，写入 context + 响应头
		c.Next()
	}
}

// JWTAuth 校验 Bearer Token（3.2）。
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = secret
		// TODO: Authorization: Bearer ... → auth.ParseToken → c.Set(CtxUserID, ...)
		response.Fail(c, http.StatusUnauthorized, 40100, "TODO: JWTAuth")
		c.Abort()
	}
}
