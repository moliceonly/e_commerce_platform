package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// CORS 跨域（阶段 H · 3.2）。
// local/staging 可放行前端 Origin；prod 应用白名单，慎用 Allow-Origin: *。
func CORS(allowOrigins []string) gin.HandlerFunc {
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"http://localhost:3000"}
	}
	return func(c *gin.Context) {
		// TODO(H2):
		//  1. 读 Origin，匹配 allowOrigins（空则默认本地调试）
		//  2. 设置 Access-Control-Allow-Origin / Headers / Methods / Credentials
		//  3. OPTIONS 预检直接 204 Abort

		origin := c.GetHeader("Origin")
		allowed := slices.Contains(allowOrigins, origin)

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
