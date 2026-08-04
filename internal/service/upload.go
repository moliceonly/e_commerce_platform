package service

import (
	"context"
	"fmt"
	"mime/multipart"
)

// UploadService 本地文件上传（阶段 H · 3.3）。
// 进阶可换成 MinIO/OSS，对外仍返回 URL。
type UploadService struct {
	Dir     string // 如 data/uploads
	BaseURL string // 如 http://127.0.0.1:8080/static
}

// SaveAvatar 保存用户头像，返回可访问 URL。
func (s *UploadService) SaveAvatar(ctx context.Context, userID uint, fh *multipart.FileHeader) (url string, err error) {
	// TODO(H3):
	//  1. 校验大小/扩展名（jpg/png）
	//  2. 存到 s.Dir/{userID}_avatar.ext
	//  3. 返回 s.BaseURL + "/" + filename
	_ = ctx
	_ = userID
	_ = fh
	return "", fmt.Errorf("TODO(H3): SaveAvatar")
}
