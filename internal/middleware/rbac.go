package middleware

import (
	"net/http"

	"e_commerce_platform/internal/response"

	"github.com/gin-gonic/gin"
)

// CtxRole JWT 解析后的角色，供 RequireRole 使用。
const CtxRole = "role"

// RequireRole RBAC 粗粒度：仅允许 listed roles（阶段 H · 3.2）。
// 依赖：JWTAuth 已 c.Set(CtxUserID) 且 c.Set(CtxRole, claims.Role)。
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO(H2):
		//  role := c.GetString(CtxRole)
		//  若不在 roles 内：Fail 403 + Abort
		_ = roles
		_ = http.StatusForbidden
		_ = response.Fail
		// 骨架阶段先放行，避免未接线时误伤；实现后删掉 Next 前的「默认放行」注释。
		c.Next()
	}
}
