package middleware

import "github.com/gin-gonic/gin"

// CORS 跨域（阶段 H · 3.2）。
// local/staging 可放行前端 Origin；prod 应用白名单，慎用 Allow-Origin: *。
func CORS(allowOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO(H2):
		//  1. 读 Origin，匹配 allowOrigins（空则默认本地调试）
		//  2. 设置 Access-Control-Allow-Origin / Headers / Methods / Credentials
		//  3. OPTIONS 预检直接 204 Abort
		_ = allowOrigins
		c.Next()
	}
}
