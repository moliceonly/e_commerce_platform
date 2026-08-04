package middleware

import (
	"net/http"
	"strings"
	"time"

	"e_commerce_platform/internal/applog"
	"e_commerce_platform/internal/auth"
	"e_commerce_platform/internal/errcode"
	"e_commerce_platform/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CtxUserID = "userID"
const CtxRequestID = "requestID"

// RequestID 注入 request_id（3.3 / G：写入 context 供 slog 使用）。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(CtxRequestID, requestID)
		c.Header("X-Request-ID", requestID)

		// TODO(G): 把 request_id 放进标准 context，供 service 打日志
		ctx := applog.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// AccessLog 结构化访问日志（阶段 G），替代或补充 gin.Logger。
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO(G):
		start := time.Now()
		c.Next()
		applog.FromContext(c.Request.Context()).Info(
			"http", "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "latency", time.Since(start).String(),
		)
	}
}

// JWTAuth 校验 Bearer Token（3.2）。
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Authorization: Bearer ... → auth.ParseToken → c.Set(CtxUserID, ...)
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer") {
			response.Fail(c, http.StatusUnauthorized, errcode.ErrAuthTokenExpired, "missing bearer token")
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")

		claims, err := auth.ParseToken(secret, tokenStr)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, errcode.ErrAuthInvalidCreds, "invalid token")
			c.Abort()
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}
