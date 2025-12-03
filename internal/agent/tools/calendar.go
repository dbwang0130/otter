package tools

import (
	"log/slog"
	"time"

	"github.com/galilio/otter/internal/calendar"
	"github.com/galilio/otter/internal/common/utils"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// calendarTools 日历工具结构体，包含 calendar Service
type calendarTools struct {
	service calendar.Service
}

// CreateCalendarItemRequest 创建日历项请求（统一结构）
// 根据 iCalendar 标准 (RFC 5545):
// - VEVENT: type=VEVENT, dtstart 必需，dtend 或 duration 至少一个（但不能同时存在）
// - VTODO: type=VTODO, dtstart 或 due 至少一个
// - VJOURNAL: type=VJOURNAL, dtstart 必需
// - VFREEBUSY: type=VFREEBUSY, dtstart 和 dtend 都必需
type CreateCalendarItemRequest struct {
	Type            string   `json:"type" binding:"required,oneof=VEVENT VTODO VJOURNAL VFREEBUSY"` // 日历项类型
	DtStart         *string  `json:"dtstart,omitempty"`                                             // 开始时间，日期时间字符串
	DtEnd           *string  `json:"dtend,omitempty"`                                               // 结束时间，日期时间字符串
	Due             *string  `json:"due,omitempty"`                                                 // 截止时间（VTODO），日期时间字符串
	Duration        *string  `json:"duration,omitempty"`                                            // 持续时间（VEVENT），与 dtend 二选一
	Summary         *string  `json:"summary,omitempty"`                                             // 标题
	Description     *string  `json:"description,omitempty"`                                         // 描述
	Location        *string  `json:"location,omitempty"`                                            // 地点
	Organizer       *string  `json:"organizer,omitempty"`                                           // 组织者
	Status          *string  `json:"status,omitempty"`                                              // 状态
	Priority        *int     `json:"priority,omitempty"`                                            // 优先级 (0-9)，仅 VTODO
	PercentComplete *int     `json:"percent_complete,omitempty"`                                    // 完成百分比 (0-100)，仅 VTODO
	RRule           *string  `json:"rrule,omitempty"`                                               // 重复规则
	ExDate          []string `json:"exdate,omitempty"`                                              // 排除日期
	RDate           []string `json:"rdate,omitempty"`                                               // 重复日期
	Categories      []string `json:"categories,omitempty"`                                          // 分类
	Comment         *string  `json:"comment,omitempty"`                                             // 备注
	Contact         *string  `json:"contact,omitempty"`                                             // 联系人
	RelatedTo       *string  `json:"related_to,omitempty"`                                          // 关联项
	Resources       []string `json:"resources,omitempty"`                                           // 资源
	URL             *string  `json:"url,omitempty"`                                                 // URL
	Class           *string  `json:"class,omitempty"`                                               // 分类（PUBLIC/PRIVATE/CONFIDENTIAL）
}

// CalendarItemResponse 日历项响应（用于列表展示）
type CalendarItemResponse struct {
	ID              uint                      `json:"id"`                         // 日历项ID
	Type            calendar.CalendarItemType `json:"type,omitempty"`             // 类型（VEVENT/VTODO/VJOURNAL/VFREEBUSY）
	Summary         *string                   `json:"summary,omitempty"`          // 标题
	Description     *string                   `json:"description,omitempty"`      // 描述
	Location        *string                   `json:"location,omitempty"`         // 地点
	Organizer       *string                   `json:"organizer,omitempty"`        // 组织者
	DtStart         time.Time                 `json:"dtstart,omitempty"`          // 开始时间
	DtEnd           *time.Time                `json:"dtend,omitempty"`            // 结束时间（VEVENT/VFREEBUSY）
	Due             *time.Time                `json:"due,omitempty"`              // 截止时间（VTODO）
	Completed       *time.Time                `json:"completed,omitempty"`        // 完成时间（VTODO）
	Status          *string                   `json:"status,omitempty"`           // 状态
	Priority        *int                      `json:"priority,omitempty"`         // 优先级 (0-9)，仅 VTODO
	PercentComplete *int                      `json:"percent_complete,omitempty"` // 完成百分比 (0-100)，仅 VTODO
	Categories      []string                  `json:"categories,omitempty"`       // 分类
}

type SearchCalendarItemsResponse struct {
	Items []*CalendarItemResponse `json:"items"`
}

// NewCalendarTools 创建日历工具列表，需要传入 calendar Service
func NewCalendarTools(service calendar.Service) ([]tool.Tool, error) {
	tools := []tool.Tool{}

	ct := calendarTools{service: service}

	// 创建 create_calendar_item 工具（统一工具，支持 VEVENT、VTODO、VJOURNAL、VFREEBUSY）
	createItemTool, err := functiontool.New(functiontool.Config{
		Name:         "create_calendar_item",
		Description:  "Create a calendar item. Supports VEVENT (requires dtstart, and either dtend or duration), VTODO (requires dtstart or due), VJOURNAL (requires dtstart), VFREEBUSY (requires both dtstart and dtend).",
		InputSchema:  utils.SchemaFromStruct(CreateCalendarItemRequest{}),
		OutputSchema: utils.SchemaFromStruct(calendar.CreateCalendarItemResponse{}),
	}, ct.CreateCalendarItem)
	if err != nil {
		slog.Error("Failed to create create_calendar_item tool", "error", err)
		return nil, err
	}
	tools = append(tools, createItemTool)

	// 创建 get_calendar_item 工具
	getItemTool, err := functiontool.New(functiontool.Config{
		Name:         "get_calendar_item",
		Description:  "Get a calendar item by ID or UID. Requires either id or uid.",
		InputSchema:  utils.SchemaFromStruct(GetCalendarItemRequest{}),
		OutputSchema: utils.SchemaFromStruct(calendar.CalendarItem{}),
	}, ct.GetCalendarItem)
	if err != nil {
		slog.Error("Failed to create get_calendar_item tool", "error", err)
		return nil, err
	}
	tools = append(tools, getItemTool)

	// 创建 update_calendar_item 工具
	updateItemTool, err := functiontool.New(functiontool.Config{
		Name:         "update_calendar_item",
		Description:  "Update a calendar item. Requires id and optional fields to update.",
		InputSchema:  utils.SchemaFromStruct(UpdateCalendarItemRequest{}),
		OutputSchema: utils.SchemaFromStruct(calendar.CalendarItem{}),
	}, ct.UpdateCalendarItem)
	if err != nil {
		slog.Error("Failed to create update_calendar_item tool", "error", err)
		return nil, err
	}
	tools = append(tools, updateItemTool)

	// 创建 delete_calendar_item 工具
	deleteItemTool, err := functiontool.New(functiontool.Config{
		Name:         "delete_calendar_item",
		Description:  "Delete a calendar item by ID.",
		InputSchema:  utils.SchemaFromStruct(DeleteCalendarItemRequest{}),
		OutputSchema: utils.SchemaFromStruct(struct{}{}),
	}, ct.DeleteCalendarItem)
	if err != nil {
		slog.Error("Failed to create delete_calendar_item tool", "error", err)
		return nil, err
	}
	tools = append(tools, deleteItemTool)

	// 创建 search_calendar_items 工具
	searchItemsTool, err := functiontool.New(functiontool.Config{
		Name:        "search_calendar_items",
		Description: "Search calendar items by keyword (q) and/or time ranges. The keyword will be searched across all searchable fields (summary, description, location, organizer, comment, contact, categories, resources). At least one search criteria (q or time_ranges) is required.",
		InputSchema: utils.SchemaFromStruct(SearchCalendarItemsRequest{}),
	}, ct.SearchCalendarItems)
	if err != nil {
		slog.Error("Failed to create search_calendar_items tool", "error", err)
		return nil, err
	}
	tools = append(tools, searchItemsTool)

	return tools, nil
}

// CreateCalendarItem 创建日历项（统一函数，支持 VEVENT、VTODO、VJOURNAL、VFREEBUSY）
func (ct *calendarTools) CreateCalendarItem(ctx tool.Context, input CreateCalendarItemRequest) (*calendar.CreateCalendarItemResponse, error) {
	var userID uint = 2 // TODO: 从 context 获取用户ID
	slog.Debug("CreateCalendarItem", "input", input)

	// 转换类型字符串为 CalendarItemType
	itemType := calendar.CalendarItemType(input.Type)
	if !isValidCalendarItemType(itemType) {
		return nil, calendar.ErrInvalidType
	}

	// 解析时间字符串
	var dtStart *time.Time
	if input.DtStart != nil {
		parsed, err := utils.ParseDateTime(*input.DtStart)
		if err != nil {
			return nil, err
		}
		dtStart = &parsed
	}

	var dtEnd *time.Time
	if input.DtEnd != nil {
		parsed, err := utils.ParseDateTime(*input.DtEnd)
		if err != nil {
			return nil, err
		}
		dtEnd = &parsed
	}

	var due *time.Time
	if input.Due != nil {
		parsed, err := utils.ParseDateTime(*input.Due)
		if err != nil {
			return nil, err
		}
		due = &parsed
	}

	req := &calendar.CreateCalendarItemRequest{
		Type:            itemType,
		DtStart:         dtStart,
		DtEnd:           dtEnd,
		Due:             due,
		Duration:        input.Duration,
		Summary:         input.Summary,
		Description:     input.Description,
		Location:        input.Location,
		Organizer:       input.Organizer,
		Status:          input.Status,
		Priority:        input.Priority,
		PercentComplete: input.PercentComplete,
		RRule:           input.RRule,
		ExDate:          input.ExDate,
		RDate:           input.RDate,
		Categories:      input.Categories,
		Comment:         input.Comment,
		Contact:         input.Contact,
		RelatedTo:       input.RelatedTo,
		Resources:       input.Resources,
		URL:             input.URL,
		Class:           input.Class,
	}

	resp, err := ct.service.CreateCalendarItem(&userID, req)
	if err != nil {
		slog.Error("Failed to create calendar item", "type", input.Type, "error", err)
		return nil, err
	}
	slog.Debug("Created calendar item", "type", input.Type, "response", resp)

	return resp, nil
}

// isValidCalendarItemType 验证日历项类型
func isValidCalendarItemType(t calendar.CalendarItemType) bool {
	return t == calendar.CalendarItemTypeEvent || t == calendar.CalendarItemTypeTodo ||
		t == calendar.CalendarItemTypeJournal || t == calendar.CalendarItemTypeFreeBusy
}

// GetCalendarItemRequest 获取日历项请求
type GetCalendarItemRequest struct {
	ID  *uint   `json:"id,omitempty"`
	UID *string `json:"uid,omitempty"`
}

// GetCalendarItem 获取日历项
func (ct *calendarTools) GetCalendarItem(ctx tool.Context, input GetCalendarItemRequest) (*calendar.CalendarItem, error) {
	// TODO: 从 context 获取用户ID，目前暂时使用硬编码值（需要根据实际的 tool.Context API 调整）
	var userID uint = 2

	if input.ID != nil {
		return ct.service.GetCalendarItemByID(&userID, *input.ID)
	}
	if input.UID != nil {
		return ct.service.GetCalendarItemByUID(&userID, *input.UID)
	}
	return nil, calendar.ErrInvalidInput
}

// UpdateCalendarItemRequest 更新日历项请求（包含ID）
type UpdateCalendarItemRequest struct {
	ID uint `json:"id" binding:"required"`
	calendar.UpdateCalendarItemRequest
}

// UpdateCalendarItem 更新日历项
func (ct *calendarTools) UpdateCalendarItem(ctx tool.Context, input UpdateCalendarItemRequest) (*calendar.CalendarItem, error) {
	// TODO: 从 context 获取用户ID，目前暂时使用硬编码值（需要根据实际的 tool.Context API 调整）
	var userID uint = 2

	return ct.service.UpdateCalendarItem(&userID, input.ID, &input.UpdateCalendarItemRequest)
}

// DeleteCalendarItemRequest 删除日历项请求
type DeleteCalendarItemRequest struct {
	ID uint `json:"id" binding:"required"`
}

// DeleteCalendarItem 删除日历项
func (ct *calendarTools) DeleteCalendarItem(ctx tool.Context, input DeleteCalendarItemRequest) (*struct{}, error) {
	// TODO: 从 context 获取用户ID，目前暂时使用 nil（需要根据实际的 tool.Context API 调整）
	var userID uint = 2

	err := ct.service.DeleteCalendarItem(&userID, input.ID)
	if err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}

// SearchCalendarItemsRequest 搜索日历项请求（工具友好版本）
type SearchCalendarItemsRequest struct {
	// 搜索关键字：在所有可搜索字段中搜索（summary, description, location, organizer, comment, contact, categories, resources）
	Q *string `json:"q,omitempty"`

	// 开始时间范围过滤（dtstart 字段的时间范围）
	DtStart *TimeRangeInput `json:"dtstart,omitempty"`

	// 返回结果数量限制，默认20，最大100
	Limit *int `json:"limit,omitempty"`
}

// TimeRangeInput 时间范围输入
type TimeRangeInput struct {
	Start *string `json:"start,omitempty"` // RFC3339 格式的时间字符串，dtstart >= start
	End   *string `json:"end,omitempty"`   // RFC3339 格式的时间字符串，dtstart <= end
}

// SearchCalendarItems 搜索日历项
func (ct *calendarTools) SearchCalendarItems(ctx tool.Context, input SearchCalendarItemsRequest) (*SearchCalendarItemsResponse, error) {
	// TODO: 从 context 获取用户ID，目前暂时使用 nil（需要根据实际的 tool.Context API 调整）
	var userID uint = 2

	// 转换时间范围
	req := calendar.SearchCalendarItemsRequest{
		Q:     input.Q,
		Limit: input.Limit,
	}

	// 转换 dtstart 时间范围
	if input.DtStart != nil {
		dtStart, err := convertTimeRangeInput(input.DtStart)
		if err != nil {
			return nil, err
		}
		if dtStart != nil {
			req.DtStart = dtStart
		}
	}

	// 调用服务层搜索
	items, err := ct.service.SearchCalendarItems(&userID, &req)
	if err != nil {
		slog.Error("Failed to search calendar items", "error", err)
		return nil, err
	}

	// 转换为响应列表
	summaries := make([]*CalendarItemResponse, 0, len(items))
	for _, item := range items {
		slog.Debug("SearchCalendarItems", "item-summary", *item.Summary)
		summaries = append(summaries, &CalendarItemResponse{
			ID:              item.ID,
			Type:            item.Type,
			Summary:         item.Summary,
			Description:     item.Description,
			Location:        item.Location,
			Organizer:       item.Organizer,
			DtStart:         item.DtStart,
			DtEnd:           item.DtEnd,
			Due:             item.Due,
			Completed:       item.Completed,
			Status:          item.Status,
			Priority:        item.Priority,
			PercentComplete: item.PercentComplete,
			// 不处理 Categories 字段
		})
	}

	return &SearchCalendarItemsResponse{Items: summaries}, nil
}

// convertTimeRangeInput 转换时间范围输入
func convertTimeRangeInput(tr *TimeRangeInput) (*calendar.TimeRange, error) {
	var start, end *time.Time
	if tr.Start != nil {
		parsed, err := utils.ParseDateTime(*tr.Start)
		if err != nil {
			return nil, err
		}
		start = &parsed
	}
	if tr.End != nil {
		parsed, err := utils.ParseDateTime(*tr.End)
		if err != nil {
			return nil, err
		}
		end = &parsed
	}
	if start == nil && end == nil {
		return nil, nil
	}
	return &calendar.TimeRange{
		Start: start,
		End:   end,
	}, nil
}
