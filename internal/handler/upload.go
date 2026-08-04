package handler

import (
	"net/http"

	"e_commerce_platform/internal/middleware"
	"e_commerce_platform/internal/response"
	"e_commerce_platform/internal/service"

	"github.com/gin-gonic/gin"
)

// UploadHandler 用户资料上传（阶段 H · 3.3）。
type UploadHandler struct{ Svc *service.UploadService }

// Avatar POST multipart field name = "file"
func (h *UploadHandler) Avatar(c *gin.Context) {
	// TODO(H3):
	//  uid := c.GetUint(middleware.CtxUserID)
	//  fh, err := c.FormFile("file")
	//  url, err := h.Svc.SaveAvatar(...)
	_ = middleware.CtxUserID
	response.Fail(c, http.StatusNotImplemented, 50100, "TODO(H3): avatar upload")
}

// AuthRefreshHandler Token 刷新（阶段 H · 3.2）。
type AuthRefreshHandler struct {
	JWTSecret string
	// 也可挂 *service.AuthService，把 Refresh 下沉到 service
}

func (h *AuthRefreshHandler) Refresh(c *gin.Context) {
	// TODO(H2): bind refresh_token → auth.RefreshAccessToken → OK({token})
	response.Fail(c, http.StatusNotImplemented, 50101, "TODO(H2): refresh token")
}
