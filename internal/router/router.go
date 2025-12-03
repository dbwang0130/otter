package router

import (
	"github.com/galilio/otter/internal/auth"
	"github.com/galilio/otter/internal/calendar"
	"github.com/galilio/otter/internal/common/config"
	"github.com/galilio/otter/internal/common/middleware"
	"github.com/galilio/otter/internal/user"
	"github.com/gin-gonic/gin"
	"google.golang.org/adk/session"
)

// Options 路由选项
type Options struct {
	SessionService   session.Service
	UserService      user.Service
	CalendarService  calendar.Service
	RefreshTokenRepo auth.RefreshTokenRepository
	JWTConfig        *config.JWTConfig
}

// Option 路由选项函数
type Option func(*Options)

// WithUserService 设置用户服务
func WithUserService(userService user.Service) Option {
	return func(opts *Options) {
		opts.UserService = userService
	}
}

// WithRefreshTokenRepo 设置刷新令牌仓库
func WithRefreshTokenRepo(repo auth.RefreshTokenRepository) Option {
	return func(opts *Options) {
		opts.RefreshTokenRepo = repo
	}
}

// WithJWTConfig 设置JWT配置
func WithJWTConfig(jwtConfig *config.JWTConfig) Option {
	return func(opts *Options) {
		opts.JWTConfig = jwtConfig
	}
}

// WithCalendarService 设置日历服务
func WithCalendarService(calendarService calendar.Service) Option {
	return func(opts *Options) {
		opts.CalendarService = calendarService
	}
}

// NewRouter 使用选项创建路由
func NewRouter(opts ...Option) *gin.Engine {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	router := gin.New()

	// 全局中间件（注意顺序：TraceID 应该在 Logger 之前）
	router.Use(middleware.TraceID())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// 健康检查（无需认证）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API路由组
	api := router.Group("/api/v1")
	{
		// 设置各功能模块的路由
		setupAuthRoutes(api, options)
		setupUserRoutes(api, options)
		setupAdminRoutes(api, options)
		setupCalendarRoutes(api, options)

	}

	return router
}
