package router

import (
	"github.com/galilio/otter/internal/admin"
	"github.com/galilio/otter/internal/common/middleware"
	"github.com/gin-gonic/gin"
)

// setupAdminRoutes 设置管理端路由
func setupAdminRoutes(api *gin.RouterGroup, opts *Options) {
	adminAPI := api.Group("/admin")
	adminAPI.Use(middleware.AuthRequired(opts.JWTConfig))
	adminAPI.Use(middleware.AdminRequired())
	{
		adminHandler := admin.NewHandler(opts.UserService)
		users := adminAPI.Group("/users")
		{
			// POST /api/v1/admin/users - 创建用户（管理员）
			// GET /api/v1/admin/users - 获取用户列表（管理员）
			// GET /api/v1/admin/users/:id - 获取指定用户（管理员）
			// PUT /api/v1/admin/users/:id - 更新指定用户（管理员）
			// DELETE /api/v1/admin/users/:id - 删除指定用户（管理员）
			users.POST("", adminHandler.CreateUser)
			users.GET("", adminHandler.ListUsers)
			users.GET("/:id", adminHandler.GetUser)
			users.PUT("/:id", adminHandler.UpdateUser)
			users.DELETE("/:id", adminHandler.DeleteUser)
		}
	}
}
