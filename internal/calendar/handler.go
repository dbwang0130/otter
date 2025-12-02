package calendar

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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

// getUserIDFromContext 从上下文中获取用户ID
func (h *Handler) getUserIDFromContext(c *gin.Context) (*uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, nil
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

// SearchCalendarItems 搜索日历项
// GET /api/v1/calendar/items/search?summary=会议&location=北京
// GET /api/v1/calendar/items/search?summary=会议&dtstart=2024-12-01T00:00:00Z,2024-12-31T23:59:59Z
// GET /api/v1/calendar/items/search?dtstart=2024-12-01T00:00:00Z,2024-12-31T23:59:59Z (仅时间范围)
// GET /api/v1/calendar/items/search?summary=会议&dtstart=2024-12-01T00:00:00Z, (只有开始时间)
// GET /api/v1/calendar/items/search?summary=会议&dtstart=,2024-12-31T23:59:59Z (只有结束时间)
func (h *Handler) SearchCalendarItems(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未认证"})
		return
	}

	// 从 Query 参数解析，格式：summary=会议&location=北京
	// 只接受可搜索的字段作为参数名
	fieldKeywords := make(map[string]string)

	// 支持的时间字段（需要跳过，不作为搜索字段）
	timeFields := []string{"dtstart", "dtend", "due", "completed"}
	timeFieldMap := make(map[string]bool)
	for _, field := range timeFields {
		timeFieldMap[field] = true
	}

	for key, values := range c.Request.URL.Query() {
		// 跳过时间范围参数
		if timeFieldMap[key] {
			continue
		}
		// 检查是否是有效的搜索字段
		field := SearchableField(key)
		if field.IsValid() && len(values) > 0 && values[0] != "" {
			fieldKeywords[key] = strings.TrimSpace(values[0])
		}
	}

	// 解析时间范围参数：支持多个时间字段的范围搜索
	// 格式：dtstart=start,end 或 dtstart=start, 或 dtstart=,end
	timeRanges := make(map[string]TimeRange)

	for _, field := range timeFields {
		timeRangeStr := c.Query(field)
		if timeRangeStr == "" {
			continue
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

		// 如果至少有一个时间值，则添加到时间范围
		if start != nil || end != nil {
			timeRanges[field] = TimeRange{
				Start: start,
				End:   end,
			}
		}
	}

	// 验证：至少需要指定一个搜索字段或时间范围
	if len(fieldKeywords) == 0 && len(timeRanges) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "至少需要指定一个搜索字段或时间范围"})
		return
	}

	req := SearchCalendarItemsRequest{
		FieldKeywords: fieldKeywords,
		TimeRanges:    timeRanges,
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
	userID, err := h.getUserIDFromContext(c)
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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的ID"})
		return
	}

	item, err := h.service.GetCalendarItemByID(uint(id))
	if err != nil {
		if err == ErrCalendarItemNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
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
	uid := c.Param("uid")
	if uid == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "UID不能为空"})
		return
	}

	item, err := h.service.GetCalendarItemByUID(uid)
	if err != nil {
		if err == ErrCalendarItemNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
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

	item, err := h.service.UpdateCalendarItem(uint(id), &req)
	if err != nil {
		if err == ErrCalendarItemNotFound || err == ErrInvalidInput {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "无效的ID"})
		return
	}

	err = h.service.DeleteCalendarItem(uint(id))
	if err != nil {
		if err == ErrCalendarItemNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
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
	userID, err := h.getUserIDFromContext(c)
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
