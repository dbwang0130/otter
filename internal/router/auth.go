package router

import (
	"github.com/galilio/otter/internal/auth"
	"github.com/galilio/otter/internal/common/middleware"
	"github.com/gin-gonic/gin"
)

// setupAuthRoutes 设置认证相关路由
func setupAuthRoutes(api *gin.RouterGroup, opts *Options) {
	authHandler := auth.NewHandler(opts.UserService, opts.RefreshTokenRepo, opts.JWTConfig)

	// 认证相关路由（无需认证）
	authGroup := api.Group("/auth")
	{
		// POST /api/v1/auth/login - 用户登录
		authGroup.POST("/login", authHandler.Login)
		// POST /api/v1/auth/refresh - 刷新Access Token
		authGroup.POST("/refresh", authHandler.Refresh)
		// POST /api/v1/auth/logout - 用户登出（撤销refresh token）
		authGroup.POST("/logout", authHandler.Logout)
	}

	// 需要认证的认证相关路由
	authProtectedGroup := api.Group("/auth")
	authProtectedGroup.Use(middleware.AuthRequired(opts.JWTConfig))
	{
		// POST /api/v1/auth/logout-all - 撤销所有refresh token（需要认证）
		authProtectedGroup.POST("/logout-all", authHandler.LogoutAll)
	}
}
