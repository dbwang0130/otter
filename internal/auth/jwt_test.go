package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	testSecret     = "test-secret-key-for-jwt-testing"
	testExpiration = 15 * time.Minute
)

// TestGenerateAccessToken_Success 测试生成Access Token成功
func TestGenerateAccessToken_Success(t *testing.T) {
	userID := uint(1)
	username := "testuser"
	isAdmin := false

	token, err := GenerateAccessToken(userID, username, isAdmin, testSecret, testExpiration)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 50) // JWT token通常比较长
}

// TestGenerateAccessToken_WithAdmin 测试生成管理员Access Token
func TestGenerateAccessToken_WithAdmin(t *testing.T) {
	userID := uint(1)
	username := "admin"
	isAdmin := true

	token, err := GenerateAccessToken(userID, username, isAdmin, testSecret, testExpiration)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证token可以解析
	claims, err := ValidateToken(token, testSecret)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.True(t, claims.IsAdmin)
}

// TestGenerateAccessToken_DifferentUsers 测试不同用户生成不同的token
func TestGenerateAccessToken_DifferentUsers(t *testing.T) {
	token1, err1 := GenerateAccessToken(1, "user1", false, testSecret, testExpiration)
	token2, err2 := GenerateAccessToken(2, "user2", false, testSecret, testExpiration)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, token1, token2) // 不同用户应该生成不同的token
}

// TestValidateToken_Success 测试验证token成功
func TestValidateToken_Success(t *testing.T) {
	userID := uint(1)
	username := "testuser"
	isAdmin := false

	token, err := GenerateAccessToken(userID, username, isAdmin, testSecret, testExpiration)
	assert.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, isAdmin, claims.IsAdmin)
}

// TestValidateToken_InvalidSecret 测试使用错误的secret验证token
func TestValidateToken_InvalidSecret(t *testing.T) {
	token, err := GenerateAccessToken(1, "testuser", false, testSecret, testExpiration)
	assert.NoError(t, err)

	wrongSecret := "wrong-secret-key"
	claims, err := ValidateToken(token, wrongSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

// TestValidateToken_ExpiredToken 测试验证过期token
func TestValidateToken_ExpiredToken(t *testing.T) {
	// 生成一个已过期的token（过期时间为过去）
	userID := uint(1)
	username := "testuser"
	isAdmin := false

	// 使用很短的过期时间，然后等待过期
	shortExpiration := 1 * time.Millisecond
	token, err := GenerateAccessToken(userID, username, isAdmin, testSecret, shortExpiration)
	assert.NoError(t, err)

	// 等待token过期
	time.Sleep(10 * time.Millisecond)

	claims, err := ValidateToken(token, testSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrExpiredToken, err)
}

// TestValidateToken_InvalidFormat 测试无效格式的token
func TestValidateToken_InvalidFormat(t *testing.T) {
	invalidToken := "not-a-valid-jwt-token"

	claims, err := ValidateToken(invalidToken, testSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

// TestValidateToken_EmptyToken 测试空token
func TestValidateToken_EmptyToken(t *testing.T) {
	claims, err := ValidateToken("", testSecret)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrInvalidToken, err)
}

// TestGenerateRefreshToken_Success 测试生成Refresh Token成功
func TestGenerateRefreshToken_Success(t *testing.T) {
	token, err := GenerateRefreshToken()

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 20) // Base64编码的32字节随机数应该比较长
}

// TestGenerateRefreshToken_Uniqueness 测试每次生成的Refresh Token都不同
func TestGenerateRefreshToken_Uniqueness(t *testing.T) {
	token1, err1 := GenerateRefreshToken()
	token2, err2 := GenerateRefreshToken()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, token1, token2) // 每次生成的token应该不同
}

// TestGenerateRefreshToken_MultipleCalls 测试多次生成token的唯一性
func TestGenerateRefreshToken_MultipleCalls(t *testing.T) {
	tokens := make(map[string]bool)

	// 生成100个token，确保都是唯一的
	for i := 0; i < 100; i++ {
		token, err := GenerateRefreshToken()
		assert.NoError(t, err)

		// 检查是否重复
		assert.False(t, tokens[token], "生成的token应该是唯一的")
		tokens[token] = true
	}
}

// TestValidateToken_ClaimsContent 测试验证token中的claims内容
func TestValidateToken_ClaimsContent(t *testing.T) {
	userID := uint(123)
	username := "testuser123"
	isAdmin := true

	token, err := GenerateAccessToken(userID, username, isAdmin, testSecret, testExpiration)
	assert.NoError(t, err)

	claims, err := ValidateToken(token, testSecret)
	assert.NoError(t, err)

	// 验证所有claims字段
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.True(t, claims.IsAdmin)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
}

// TestGenerateToken_Compatibility 测试兼容性函数
func TestGenerateToken_Compatibility(t *testing.T) {
	userID := uint(1)
	username := "testuser"
	isAdmin := false

	// 使用兼容函数生成token
	token, err := GenerateToken(userID, username, isAdmin, testSecret, testExpiration)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证token可以正常解析
	claims, err := ValidateToken(token, testSecret)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
}

// TestValidateToken_DifferentExpiration 测试不同过期时间的token
func TestValidateToken_DifferentExpiration(t *testing.T) {
	testCases := []struct {
		name       string
		expiration time.Duration
	}{
		{"1分钟", 1 * time.Minute},
		{"15分钟", 15 * time.Minute},
		{"1小时", 1 * time.Hour},
		{"24小时", 24 * time.Hour},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := GenerateAccessToken(1, "testuser", false, testSecret, tc.expiration)
			assert.NoError(t, err)

			claims, err := ValidateToken(token, testSecret)
			assert.NoError(t, err)
			assert.NotNil(t, claims)

			// 验证过期时间设置正确
			expectedExpiry := time.Now().Add(tc.expiration)
			actualExpiry := claims.ExpiresAt.Time

			// 允许1秒的误差
			diff := actualExpiry.Sub(expectedExpiry)
			assert.True(t, diff < time.Second && diff > -time.Second,
				"过期时间应该在预期范围内")
		})
	}
}
