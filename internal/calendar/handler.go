package calendar

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/galilio/otter/internal/common/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// SearchCalendarItems 搜索日历项
// GET /api/v1/calendar/items/search?summary=会议&location=北京
// GET /api/v1/calendar/items/search?summary=会议&dtstart=2024-12-01T00:00:00Z,2024-12-31T23:59:59Z
// GET /api/v1/calendar/items/search?dtstart=2024-12-01T00:00:00Z,2024-12-31T23:59:59Z (仅时间范围)
// GET /api/v1/calendar/items/search?summary=会议&dtstart=2024-12-01T00:00:00Z, (只有开始时间)
// GET /api/v1/calendar/items/search?summary=会议&dtstart=,2024-12-31T23:59:59Z (只有结束时间)
func (h *Handler) SearchCalendarItems(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未认证"})
		return
	}

	// 解析搜索关键字 q 参数
	var q *string
	if qStr := c.Query("q"); qStr != "" {
		q = &qStr
	}

	// 解析时间范围参数：支持多个时间字段的范围搜索
	// 格式：dtstart=start,end 或 dtstart=start, 或 dtstart=,end
	parseTimeRange := func(field string) *TimeRange {
		timeRangeStr := c.Query(field)
		if timeRangeStr == "" {
			return nil
		}

		// 解析 start,end 格式
		parts := strings.Split(timeRangeStr, ",")
		var start, end *time.Time

		// 解析开始时间
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[0])); err == nil {
				start = &parsed
			}
		}

		// 解析结束时间
		if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
			if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[1])); err == nil {
				end = &parsed
			}
		}

		// 如果至少有一个时间值，则返回时间范围
		if start != nil || end != nil {
			return &TimeRange{
				Start: start,
				End:   end,
			}
		}

		return nil
	}

	req := SearchCalendarItemsRequest{
		Q:         q,
		DtStart:   parseTimeRange("dtstart"),
		DtEnd:     parseTimeRange("dtend"),
		Due:       parseTimeRange("due"),
		Completed: parseTimeRange("completed"),
	}

	// 验证：至少需要指定搜索关键字或时间范围
	if req.Q == nil && req.DtStart == nil && req.DtEnd == nil && req.Due == nil && req.Completed == nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "至少需要指定搜索关键字(q)或时间范围"})
		return
	}

	items, err := h.service.SearchCalendarItems(userID, &req)
	if err != nil {
		if err == ErrInvalidSearchField || err == ErrInvalidInput {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"count": len(items),
	})
}

// CreateCalendarItem 创建日历项
// POST /api/v1/calendar/items
func (h *Handler) CreateCalendarItem(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未认证"})
		return
	}

	var req CreateCalendarItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := h.service.CreateCalendarItem(userID, &req)
	if err != nil {
		if err == ErrInvalidType || err == ErrInvalidInput {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// GetCalendarItem 根据ID获取日历项
// GET /api/v1/calendar/items/:id
func (h *Handler) GetCalendarItem(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未认证"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的ID"})
		return
	}

	item, err := h.service.GetCalendarItemByID(userID, uint(id))
	if err != nil {
		if err == ErrCalendarItemNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		if err == ErrForbidden {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// GetCalendarItemByUID 根据UID获取日历项
// GET /api/v1/calendar/items/uid/:uid
func (h *Handler) GetCalendarItemByUID(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未认证"})
		return
	}

	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "UID不能为空"})
		return
	}

	item, err := h.service.GetCalendarItemByUID(userID, uid)
	if err != nil {
		if err == ErrCalendarItemNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		if err == ErrForbidden {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// UpdateCalendarItem 更新日历项
// PUT /api/v1/calendar/items/:id
func (h *Handler) UpdateCalendarItem(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未认证"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的ID"})
		return
	}

	var req UpdateCalendarItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := h.service.UpdateCalendarItem(userID, uint(id), &req)
	if err != nil {
		if err == ErrCalendarItemNotFound || err == ErrInvalidInput {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		if err == ErrForbidden {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteCalendarItem 删除日历项
// DELETE /api/v1/calendar/items/:id
func (h *Handler) DeleteCalendarItem(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未认证"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的ID"})
		return
	}

	err = h.service.DeleteCalendarItem(userID, uint(id))
	if err != nil {
		if err == ErrCalendarItemNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		if err == ErrForbidden {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "日历项删除成功"})
}

// ListCalendarItems 列出日历项
// GET /api/v1/calendar/items
func (h *Handler) ListCalendarItems(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未认证"})
		return
	}

	var req ListCalendarItemsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.ListCalendarItems(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
