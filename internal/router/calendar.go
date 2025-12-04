package router

import (
	"github.com/galilio/otter/internal/calendar"
	"github.com/galilio/otter/internal/common/middleware"
	"github.com/gin-gonic/gin"
)

// setupCalendarRoutes 设置日历相关路由
func setupCalendarRoutes(api *gin.RouterGroup, opts *Options) {
	calendarHandler := calendar.NewHandler(opts.CalendarService)
	items := api.Group("/calendar/items")
	items.Use(middleware.AuthRequired(opts.JWTConfig))
	
	// POST /api/v1/calendar/items - 创建日历项
	items.POST("", calendarHandler.CreateCalendarItem)
	// GET /api/v1/calendar/items - 列出日历项
	items.GET("", calendarHandler.ListCalendarItems)
	// GET /api/v1/calendar/items/search - 搜索日历项
	items.GET("/search", calendarHandler.SearchCalendarItems)
	// GET /api/v1/calendar/items/uid/:uid - 根据UID获取日历项
	items.GET("/uid/:uid", calendarHandler.GetCalendarItemByUID)
	// GET /api/v1/calendar/items/:id - 根据ID获取日历项
	items.GET("/:id", calendarHandler.GetCalendarItem)
	// PUT /api/v1/calendar/items/:id - 更新日历项
	items.PUT("/:id", calendarHandler.UpdateCalendarItem)
	// DELETE /api/v1/calendar/items/:id - 删除日历项
	items.DELETE("/:id", calendarHandler.DeleteCalendarItem)
}
