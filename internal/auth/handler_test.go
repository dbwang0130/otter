package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/galilio/otter/internal/common/config"
	"github.com/galilio/otter/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService 模拟用户服务
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(req *user.CreateUserRequest) (*user.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uint) (*user.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(id uint, req *user.UpdateUserRequest) (*user.User, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) ListUsers(page, pageSize int) (*user.UserListResponse, error) {
	args := m.Called(page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserListResponse), args.Error(1)
}

func (m *MockUserService) Login(req *user.LoginRequest) (*user.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

// MockRefreshTokenRepository 模拟刷新token仓库
type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(token *RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) GetByToken(token string) (*RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) RevokeByToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) RevokeByUserID(userID uint) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpired() error {
	args := m.Called()
	return args.Error(0)
}

// setupTestHandler 创建测试用的handler
func setupTestHandler() (*Handler, *MockUserService, *MockRefreshTokenRepository, *config.JWTConfig) {
	gin.SetMode(gin.TestMode)

	mockUserService := new(MockUserService)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)

	jwtConfig := &config.JWTConfig{
		Secret:            "test-secret-key",
		Expiration:        15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}

	handler := NewHandler(mockUserService, mockRefreshTokenRepo, jwtConfig)

	return handler, mockUserService, mockRefreshTokenRepo, jwtConfig
}

// TestHandler_Login_Success 测试登录成功
func TestHandler_Login_Success(t *testing.T) {
	handler, mockUserService, mockRefreshTokenRepo, _ := setupTestHandler()

	// 准备测试数据
	testUser := &user.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		IsAdmin:   false,
		Status:    "active",
	}

	loginReq := &user.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	// 设置Mock期望
	mockUserService.On("Login", loginReq).Return(testUser, nil)
	mockRefreshTokenRepo.On("Create", mock.AnythingOfType("*auth.RefreshToken")).Return(nil)

	// 创建HTTP请求
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// 执行
	handler.Login(c)

	// 验证
	assert.Equal(t, http.StatusOK, w.Code)

	var response LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, testUser.Username, response.User.Username)
	assert.Equal(t, testUser.Email, response.User.Email)

	mockUserService.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestHandler_Login_InvalidRequest 测试无效请求
func TestHandler_Login_InvalidRequest(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	// 无效的JSON
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestHandler_Login_UserNotFound 测试用户不存在
func TestHandler_Login_UserNotFound(t *testing.T) {
	handler, mockUserService, _, _ := setupTestHandler()

	loginReq := &user.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}

	mockUserService.On("Login", loginReq).Return(nil, user.ErrUserNotFound)

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response.Error, "用户名或密码错误")

	mockUserService.AssertExpectations(t)
}

// TestHandler_Login_InvalidPassword 测试密码错误
func TestHandler_Login_InvalidPassword(t *testing.T) {
	handler, mockUserService, _, _ := setupTestHandler()

	loginReq := &user.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}

	mockUserService.On("Login", loginReq).Return(nil, user.ErrInvalidPassword)

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response.Error, "用户名或密码错误")

	mockUserService.AssertExpectations(t)
}

// TestHandler_Refresh_Success 测试刷新token成功
func TestHandler_Refresh_Success(t *testing.T) {
	handler, mockUserService, mockRefreshTokenRepo, _ := setupTestHandler()

	testUser := &user.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		IsAdmin:  false,
		Status:   "active",
	}

	refreshToken := "valid-refresh-token"
	refreshTokenModel := &RefreshToken{
		ID:        1,
		Token:     refreshToken,
		UserID:    1,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IsRevoked: false,
	}

	mockRefreshTokenRepo.On("GetByToken", refreshToken).Return(refreshTokenModel, nil)
	mockUserService.On("GetUserByID", uint(1)).Return(testUser, nil)

	refreshReq := RefreshRequest{RefreshToken: refreshToken}
	body, _ := json.Marshal(refreshReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Refresh(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response RefreshResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)

	mockRefreshTokenRepo.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

// TestHandler_Refresh_InvalidToken 测试无效的refresh token
func TestHandler_Refresh_InvalidToken(t *testing.T) {
	handler, _, mockRefreshTokenRepo, _ := setupTestHandler()

	refreshToken := "invalid-token"
	mockRefreshTokenRepo.On("GetByToken", refreshToken).Return(nil, errors.New("not found"))

	refreshReq := RefreshRequest{RefreshToken: refreshToken}
	body, _ := json.Marshal(refreshReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Refresh(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response.Error, "无效的refresh token")

	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestHandler_Refresh_ExpiredToken 测试过期的refresh token
func TestHandler_Refresh_ExpiredToken(t *testing.T) {
	handler, _, mockRefreshTokenRepo, _ := setupTestHandler()

	refreshToken := "expired-token"
	expiredTokenModel := &RefreshToken{
		ID:        1,
		Token:     refreshToken,
		UserID:    1,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // 已过期
		IsRevoked: false,
	}

	mockRefreshTokenRepo.On("GetByToken", refreshToken).Return(expiredTokenModel, nil)

	refreshReq := RefreshRequest{RefreshToken: refreshToken}
	body, _ := json.Marshal(refreshReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Refresh(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response.Error, "refresh token已过期")

	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestHandler_Logout_Success 测试登出成功
func TestHandler_Logout_Success(t *testing.T) {
	handler, _, mockRefreshTokenRepo, _ := setupTestHandler()

	refreshToken := "valid-refresh-token"
	mockRefreshTokenRepo.On("RevokeByToken", refreshToken).Return(nil)

	logoutReq := RefreshRequest{RefreshToken: refreshToken}
	body, _ := json.Marshal(logoutReq)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Logout(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "登出成功", response["message"])

	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestHandler_Logout_InvalidRequest 测试登出无效请求
func TestHandler_Logout_InvalidRequest(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.Logout(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestHandler_LogoutAll_Success 测试撤销所有token成功
func TestHandler_LogoutAll_Success(t *testing.T) {
	handler, _, mockRefreshTokenRepo, _ := setupTestHandler()

	userID := uint(1)
	mockRefreshTokenRepo.On("RevokeByUserID", userID).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout-all", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.LogoutAll(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "已撤销所有token", response["message"])

	mockRefreshTokenRepo.AssertExpectations(t)
}

// TestHandler_LogoutAll_Unauthorized 测试未认证用户
func TestHandler_LogoutAll_Unauthorized(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout-all", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// 不设置user_id

	handler.LogoutAll(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response.Error, "未认证")
}

// TestHandler_LogoutAll_InvalidUserID 测试无效的用户ID类型
func TestHandler_LogoutAll_InvalidUserID(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout-all", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "invalid-type") // 无效类型

	handler.LogoutAll(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response.Error, "无效的用户ID类型")
}

// TestHandler_LogoutAll_DifferentUserIDTypes 测试不同用户ID类型
func TestHandler_LogoutAll_DifferentUserIDTypes(t *testing.T) {
	handler, _, mockRefreshTokenRepo, _ := setupTestHandler()

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
			mockRefreshTokenRepo.On("RevokeByUserID", tc.expect).Return(nil)

			req := httptest.NewRequest(http.MethodPost, "/auth/logout-all", nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", tc.userID)

			handler.LogoutAll(c)

			assert.Equal(t, http.StatusOK, w.Code)
			mockRefreshTokenRepo.AssertExpectations(t)
		})
	}
}

func TestHandler_LogoutAll_RevokeByUserIDError(t *testing.T) {
	handler, _, mockRefreshTokenRepo, _ := setupTestHandler()
	userID := uint(1)
	mockRefreshTokenRepo.On("RevokeByUserID", userID).Return(errors.New("revoke by user id error"))

	req := httptest.NewRequest(http.MethodPost, "/auth/logout-all", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.LogoutAll(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response.Error, "撤销所有token失败")
	mockRefreshTokenRepo.AssertExpectations(t)
}
