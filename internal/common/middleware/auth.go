package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/galilio/otter/internal/auth"
	"github.com/galilio/otter/internal/common/config"
	"github.com/gin-gonic/gin"
)

// AuthRequired JWT认证中间件
func AuthRequired(jwtConfig *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证token"})
			c.Abort()
			return
		}

		// 提取Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误，应为: Bearer <token>"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := auth.ValidateToken(token, jwtConfig.Secret)
		if err != nil {
			if err == auth.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "token已过期"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			}
			c.Abort()
			return
		}

		// 将用户信息存储到context中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)

		c.Next()
	}
}

// AdminRequired 管理员权限中间件
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		if !isAdmin.(bool) {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetCurrentUserID 从context中获取当前用户ID（辅助函数）
func GetCurrentUserID(c *gin.Context) (uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, gin.Error{Err: gin.Error{Err: nil}, Meta: "用户未认证"}
	}

	switch v := userID.(type) {
	case uint:
		return v, nil
	case uint64:
		return uint(v), nil
	case int:
		return uint(v), nil
	case int64:
		return uint(v), nil
	default:
		return 0, gin.Error{Err: gin.Error{Err: nil}, Meta: "无效的用户ID类型"}
	}
}

// IsAdmin 检查当前用户是否为管理员（辅助函数）
func IsAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get("is_admin")
	if !exists {
		return false
	}
	admin, ok := isAdmin.(bool)
	return ok && admin
}

// GetUserIDFromContext 从上下文中获取用户ID，返回指针类型
func GetUserIDFromContext(c *gin.Context) (*uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, errors.New("用户未认证")
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
		return nil, nil
	}

	return &uid, nil
}
