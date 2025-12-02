package calendar

import (
	"net/http"
	"strconv"
	"strings"

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
// GET /api/v1/calendar/items/search?fields=summary&fields=location&keyword=测试
// 或者 GET /api/v1/calendar/items/search?fields=summary,location&keyword=测试
func (h *Handler) SearchCalendarItems(c *gin.Context) {
	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "未认证"})
		return
	}

	var req SearchCalendarItemsRequest
	
	// 处理 fields 参数：支持两种格式
	// 1. fields=summary&fields=location (多个同名参数)
	// 2. fields=summary,location (逗号分隔)
	fieldsParam := c.QueryArray("fields")
	if len(fieldsParam) == 0 {
		// 尝试逗号分隔的格式
		fieldsStr := c.Query("fields")
		if fieldsStr != "" {
			fieldsParam = []string{fieldsStr}
		}
	}
	
	// 展开逗号分隔的字段
	var fields []string
	for _, f := range fieldsParam {
		if f != "" {
			// 如果包含逗号，则分割
			if strings.Contains(f, ",") {
				parts := strings.Split(f, ",")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if part != "" {
						fields = append(fields, part)
					}
				}
			} else {
				fields = append(fields, strings.TrimSpace(f))
			}
		}
	}
	
	req.Fields = fields
	req.Keyword = strings.TrimSpace(c.Query("keyword"))
	
	// 手动验证
	if len(req.Fields) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "fields 参数不能为空"})
		return
	}
	if req.Keyword == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "keyword 参数不能为空"})
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

