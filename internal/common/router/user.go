package router

import (
	"github.com/galilio/otter/internal/common/middleware"
	"github.com/galilio/otter/internal/user"
	"github.com/gin-gonic/gin"
)

// setupUserRoutes 设置用户端路由
func setupUserRoutes(api *gin.RouterGroup, opts *Options) {
	userAPI := api.Group("/users")
	userAPI.Use(middleware.AuthRequired(opts.JWTConfig))
	{
		userHandler := user.NewHandler(opts.UserService)
		// GET /api/v1/users/me - 获取当前用户信息
		// PUT /api/v1/users/me - 更新当前用户信息
		// DELETE /api/v1/users/me - 删除当前用户
		userAPI.GET("/me", userHandler.GetCurrentUser)
		userAPI.PUT("/me", userHandler.UpdateCurrentUser)
		userAPI.DELETE("/me", userHandler.DeleteCurrentUser)
	}
}
