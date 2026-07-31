package auth

import (
	"fmt"
	"time"
)

// HashPassword 注册时哈希（3.2：禁止明文）。
func HashPassword(plain string) (string, error) {
	_ = plain
	// TODO: bcrypt.GenerateFromPassword
	return "", fmt.Errorf("TODO: HashPassword")
}

// CheckPassword 登录校验。
func CheckPassword(hash, plain string) bool {
	_ = hash
	_ = plain
	// TODO: bcrypt.CompareHashAndPassword
	return false
}

// Claims JWT 载荷（自己定字段：uid / role / exp）。
type Claims struct {
	UserID uint   `json:"uid"`
	Role   string `json:"role"`
	// TODO: 嵌入 jwt.RegisteredClaims
}

// SignToken 签发 JWT。
func SignToken(secret string, userID uint, role string, ttl time.Duration) (string, error) {
	_ = secret
	_ = userID
	_ = role
	_ = ttl
	// TODO: jwt.NewWithClaims + SignedString
	return "", fmt.Errorf("TODO: SignToken")
}

// ParseToken 校验 JWT。
func ParseToken(secret, tokenStr string) (*Claims, error) {
	_ = secret
	_ = tokenStr
	// TODO: jwt.ParseWithClaims
	return nil, fmt.Errorf("TODO: ParseToken")
}
