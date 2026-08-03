package middleware

import (
	"net/http"
	"strings"

	"e_commerce_platform/internal/auth"
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
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer") {
			response.Fail(c, http.StatusUnauthorized, 40100, "missing bearer token")
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")

		claims, err := auth.ParseToken(secret, tokenStr)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, 40101, "invalid token")
			c.Abort()
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Next()
	}
}
