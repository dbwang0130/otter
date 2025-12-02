package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("无效的token")
	ErrExpiredToken = errors.New("token已过期")
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	IsAdmin  bool   `json:"is_admin"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成Access Token（短期，用于API访问）
func GenerateAccessToken(userID uint, username string, isAdmin bool, secret string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		IsAdmin:  isAdmin,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken 生成Refresh Token（长期，用于刷新access token）
// 返回随机字符串token，不包含JWT claims
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateToken 兼容旧接口，生成Access Token
func GenerateToken(userID uint, username string, isAdmin bool, secret string, expiration time.Duration) (string, error) {
	return GenerateAccessToken(userID, username, isAdmin, secret, expiration)
}

// ValidateToken 验证JWT token
func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(secret), nil
	})

	if err != nil {
		// 检查是否是token过期错误
		if err.Error() == "token is expired" || err.Error() == "token has invalid claims: token is expired" {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
