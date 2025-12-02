package router

import (
	"github.com/galilio/otter/internal/common/middleware"
	"github.com/gin-gonic/gin"
)

// setupAgentRoutes 设置代理路由
func setupAgentRoutes(api *gin.RouterGroup, opts *Options) {
	agentAPI := api.Group("/agents")
	agentAPI.Use(middleware.AuthRequired(opts.JWTConfig))
	{
		// 代理相关路由将在这里添加
	}
}
