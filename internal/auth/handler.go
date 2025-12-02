package auth

import (
	"net/http"
	"time"

	"github.com/galilio/otter/internal/common/config"
	"github.com/galilio/otter/internal/user"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	userService      user.Service
	refreshTokenRepo RefreshTokenRepository
	jwtConfig        *config.JWTConfig
}

func NewHandler(userService user.Service, refreshTokenRepo RefreshTokenRepository, jwtConfig *config.JWTConfig) *Handler {
	return &Handler{
		userService:      userService,
		refreshTokenRepo: refreshTokenRepo,
		jwtConfig:        jwtConfig,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type LoginResponse struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	User                  *UserInfo `json:"user"`
}

type RefreshResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsAdmin   bool   `json:"is_admin"`
}

// Login 用户登录
func (h *Handler) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	u, err := h.userService.Login(&req)
	if err != nil {
		if err == user.ErrUserNotFound {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "用户名或密码错误"})
			return
		}
		if err == user.ErrInvalidPassword {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "用户名或密码错误"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// 生成Access Token
	accessExpiration := h.jwtConfig.Expiration
	accessToken, err := GenerateAccessToken(u.ID, u.Username, u.IsAdmin, h.jwtConfig.Secret, accessExpiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "生成access token失败"})
		return
	}

	// 生成Refresh Token
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "生成refresh token失败"})
		return
	}

	// 保存Refresh Token到数据库
	refreshExpiration := h.jwtConfig.RefreshExpiration
	refreshTokenModel := &RefreshToken{
		Token:     refreshToken,
		UserID:    u.ID,
		ExpiresAt: time.Now().Add(refreshExpiration),
		IsRevoked: false,
	}

	if err := h.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "保存refresh token失败"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  time.Now().Add(accessExpiration),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: time.Now().Add(refreshExpiration),
		User: &UserInfo{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			IsAdmin:   u.IsAdmin,
		},
	})
}

// Refresh 刷新Access Token
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// 验证Refresh Token
	refreshTokenModel, err := h.refreshTokenRepo.GetByToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "无效的refresh token"})
		return
	}

	// 检查是否过期
	if time.Now().After(refreshTokenModel.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "refresh token已过期"})
		return
	}

	// 获取用户信息
	u, err := h.userService.GetUserByID(refreshTokenModel.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "用户不存在"})
		return
	}

	// 生成新的Access Token
	accessExpiration := h.jwtConfig.Expiration
	accessToken, err := GenerateAccessToken(u.ID, u.Username, u.IsAdmin, h.jwtConfig.Secret, accessExpiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "生成access token失败"})
		return
	}

	c.JSON(http.StatusOK, RefreshResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: time.Now().Add(accessExpiration),
	})
}

// Logout 用户登出（撤销refresh token）
func (h *Handler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// 撤销refresh token
	if err := h.refreshTokenRepo.RevokeByToken(req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "撤销token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// LogoutAll 撤销用户的所有refresh token（用于安全场景，如密码修改）
func (h *Handler) LogoutAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未认证"})
		return
	}

	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case uint64:
		uid = uint(v)
	case int:
		uid = uint(v)
	case int64:
		uid = uint(v)
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的用户ID类型"})
		return
	}

	// 撤销该用户的所有refresh token
	if err := h.refreshTokenRepo.RevokeByUserID(uid); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "撤销所有token失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已撤销所有token"})
}
