package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/galilio/otter/internal/auth"
	"github.com/galilio/otter/internal/common/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const (
	testSecret     = "test-secret-key-for-middleware"
	testExpiration = 15 * time.Minute
)

// setupTestJWTConfig 创建测试用的JWT配置
func setupTestJWTConfig() *config.JWTConfig {
	return &config.JWTConfig{
		Secret:            testSecret,
		Expiration:        testExpiration,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
}

// TestAuthRequired_NoToken 测试未提供token
func TestAuthRequired_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtConfig := setupTestJWTConfig()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	AuthRequired(jwtConfig)(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "未提供认证token")
}

// TestAuthRequired_InvalidFormat 测试认证格式错误
func TestAuthRequired_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtConfig := setupTestJWTConfig()

	testCases := []struct {
		name        string
		authHeader  string
		expectError string
	}{
		{"无Bearer前缀", "invalid-token", "认证格式错误"},
		{"空token", "Bearer ", "无效的token"},     // 空token会被传给ValidateToken，返回"无效的token"
		{"多个空格", "Bearer  token", "无效的token"}, // 带空格的token会被传给ValidateToken，返回"无效的token"
		{"错误前缀", "Basic token123", "认证格式错误"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tc.authHeader)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			AuthRequired(jwtConfig)(c)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectError)
		})
	}
}

// TestAuthRequired_InvalidToken 测试无效token
func TestAuthRequired_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtConfig := setupTestJWTConfig()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	AuthRequired(jwtConfig)(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}

// TestAuthRequired_ExpiredToken 测试过期token
func TestAuthRequired_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtConfig := setupTestJWTConfig()

	// 生成一个已过期的token
	shortExpiration := 1 * time.Millisecond
	token, err := auth.GenerateAccessToken(1, "testuser", false, testSecret, shortExpiration)
	assert.NoError(t, err)

	// 等待token过期
	time.Sleep(10 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	AuthRequired(jwtConfig)(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "token已过期")
}

// TestAuthRequired_Success 测试认证成功
func TestAuthRequired_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtConfig := setupTestJWTConfig()

	// 生成有效token
	userID := uint(1)
	username := "testuser"
	isAdmin := false
	token, err := auth.GenerateAccessToken(userID, username, isAdmin, testSecret, testExpiration)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 创建一个handler来验证context中的值
	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		// 验证context中的值
		uid, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, userID, uid)

		uname, exists := c.Get("username")
		assert.True(t, exists)
		assert.Equal(t, username, uname)

		admin, exists := c.Get("is_admin")
		assert.True(t, exists)
		assert.Equal(t, isAdmin, admin)
	}

	// 执行中间件和handler
	AuthRequired(jwtConfig)(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthRequired_WithAdmin 测试管理员token
func TestAuthRequired_WithAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtConfig := setupTestJWTConfig()

	// 生成管理员token
	token, err := auth.GenerateAccessToken(1, "admin", true, testSecret, testExpiration)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
		isAdmin, exists := c.Get("is_admin")
		assert.True(t, exists)
		assert.True(t, isAdmin.(bool))
	}

	AuthRequired(jwtConfig)(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.True(t, handlerCalled)
}

// TestAuthRequired_WrongSecret 测试错误的secret
func TestAuthRequired_WrongSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 使用一个secret生成token
	token, err := auth.GenerateAccessToken(1, "testuser", false, "wrong-secret", testExpiration)
	assert.NoError(t, err)

	// 使用另一个secret验证
	jwtConfig := setupTestJWTConfig()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	AuthRequired(jwtConfig)(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "无效的token")
}

// TestAdminRequired_NotAuthenticated 测试未认证用户
func TestAdminRequired_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// 不设置is_admin

	AdminRequired()(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "未认证")
}

// TestAdminRequired_NotAdmin 测试非管理员用户
func TestAdminRequired_NotAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("is_admin", false) // 设置为非管理员

	AdminRequired()(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "需要管理员权限")
}

// TestAdminRequired_Success 测试管理员权限验证成功
func TestAdminRequired_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("is_admin", true) // 设置为管理员

	handlerCalled := false
	testHandler := func(c *gin.Context) {
		handlerCalled = true
	}

	AdminRequired()(c)
	if !c.IsAborted() {
		testHandler(c)
	}

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestGetCurrentUserID_Success 测试获取用户ID成功
func TestGetCurrentUserID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name   string
		userID interface{}
		expect uint
	}{
		{"uint", uint(1), 1},
		{"uint64", uint64(2), 2},
		{"int", 3, 3},
		{"int64", int64(4), 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", tc.userID)

			uid, err := GetCurrentUserID(c)

			assert.NoError(t, err)
			assert.Equal(t, tc.expect, uid)
		})
	}
}

// TestGetCurrentUserID_NotExists 测试用户ID不存在
func TestGetCurrentUserID_NotExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// 不设置user_id

	uid, err := GetCurrentUserID(c)

	assert.Error(t, err)
	assert.Equal(t, uint(0), uid)
}

// TestGetCurrentUserID_InvalidType 测试无效的用户ID类型
func TestGetCurrentUserID_InvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "invalid-type") // 无效类型

	uid, err := GetCurrentUserID(c)

	assert.Error(t, err)
	assert.Equal(t, uint(0), uid)
}

// TestIsAdmin_True 测试是管理员
func TestIsAdmin_True(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("is_admin", true)

	result := IsAdmin(c)

	assert.True(t, result)
}

// TestIsAdmin_False 测试不是管理员
func TestIsAdmin_False(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("is_admin", false)

	result := IsAdmin(c)

	assert.False(t, result)
}

// TestIsAdmin_NotExists 测试is_admin不存在
func TestIsAdmin_NotExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// 不设置is_admin

	result := IsAdmin(c)

	assert.False(t, result)
}

// TestIsAdmin_InvalidType 测试is_admin类型错误
func TestIsAdmin_InvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("is_admin", "true") // 字符串类型，不是bool

	result := IsAdmin(c)

	assert.False(t, result)
}
