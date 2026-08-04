package auth

import (
	"fmt"
	"time"
)

// SignRefreshToken 签发较长寿命的 refresh token（阶段 H · 3.2）。
// 建议：独立 claims 类型或 RegisteredClaims.Subject="refresh"；可带 jti 便于黑名单。
func SignRefreshToken(secret string, userID uint, role string, ttl time.Duration) (string, error) {
	// TODO(H2): 类似 SignToken，ttl 更长（如 7d）；claims 标记 token 类型
	_ = secret
	_ = userID
	_ = role
	_ = ttl
	return "", fmt.Errorf("TODO(H2): SignRefreshToken")
}

// RefreshAccessToken 校验 refresh，换发新的 access token。
func RefreshAccessToken(secret, refreshToken string, accessTTL time.Duration) (access string, err error) {
	// TODO(H2):
	//  1. Parse 并确认是 refresh
	//  2. SignToken 发短 access
	_ = secret
	_ = refreshToken
	_ = accessTTL
	return "", fmt.Errorf("TODO(H2): RefreshAccessToken")
}
