package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 注册时哈希（3.2：禁止明文）。
func HashPassword(plain string) (string, error) {
	// TODO: bcrypt.GenerateFromPassword
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword 登录校验。
func CheckPassword(hash, plain string) bool {
	// TODO: bcrypt.CompareHashAndPassword
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// Claims JWT 载荷（自己定字段：uid / role / exp）。
type Claims struct {
	UserID uint   `json:"uid"`
	Role   string `json:"role"`
	// TODO: 嵌入 jwt.RegisteredClaims
	jwt.RegisteredClaims
}

// SignToken 签发 JWT。
func SignToken(secret string, userID uint, role string, ttl time.Duration) (string, error) {
	// TODO: jwt.NewWithClaims + SignedString
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 校验 JWT。
func ParseToken(secret, tokenStr string) (*Claims, error) {
	// TODO: jwt.ParseWithClaims
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
