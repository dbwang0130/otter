package calendar

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCalendarItemNotFound = errors.New("日历项不存在")
	ErrValarmNotFound       = errors.New("提醒不存在")
	ErrInvalidInput         = errors.New("输入参数无效")
	ErrInvalidType          = errors.New("无效的日历项类型")
	ErrInvalidAction        = errors.New("无效的提醒动作类型")
	ErrInvalidSearchField   = errors.New("无效的搜索字段")
	ErrForbidden            = errors.New("无权访问该日历项")
)

// SearchableField 可搜索的字段
type SearchableField string

const (
	SearchFieldSummary     SearchableField = "summary"
	SearchFieldDescription SearchableField = "description"
	SearchFieldLocation    SearchableField = "location"
	SearchFieldOrganizer   SearchableField = "organizer"
	SearchFieldComment     SearchableField = "comment"
	SearchFieldContact     SearchableField = "contact"
	SearchFieldCategories  SearchableField = "categories"
	SearchFieldResources   SearchableField = "resources"
)

// IsValid 验证字段是否有效
func (f SearchableField) IsValid() bool {
	switch f {
	case SearchFieldSummary, SearchFieldDescription, SearchFieldLocation,
		SearchFieldOrganizer, SearchFieldComment, SearchFieldContact,
		SearchFieldCategories, SearchFieldResources:
		return true
	default:
		return false
	}
}

// Service 日历服务接口
type Service interface {
	// CalendarItem 相关方法
	CreateCalendarItem(userID *uint, req *CreateCalendarItemRequest) (*CreateCalendarItemResponse, error)
	GetCalendarItemByID(userID *uint, id uint) (*CalendarItem, error)
	GetCalendarItemByUID(userID *uint, uid string) (*CalendarItem, error)
	UpdateCalendarItem(userID *uint, id uint, req *UpdateCalendarItemRequest) (*CalendarItem, error)
	DeleteCalendarItem(userID *uint, id uint) error
	ListCalendarItems(userID *uint, req *ListCalendarItemsRequest) (*CalendarItemListResponse, error)
	SearchCalendarItems(userID *uint, req *SearchCalendarItemsRequest) ([]*CalendarItem, error)

	// Valarm 相关方法
	CreateValarm(calendarItemID uint, req *CreateValarmRequest) (*Valarm, error)
	GetValarmByID(id uint) (*Valarm, error)
	GetValarmsByCalendarItemID(calendarItemID uint) ([]*Valarm, error)
	UpdateValarm(id uint, req *UpdateValarmRequest) (*Valarm, error)
	DeleteValarm(id uint) error
}

// CreateCalendarItemRequest 创建日历项请求
// 根据 iCalendar 标准 (RFC 5545):
// - VEVENT: DTSTART 必需，DTEND 或 DURATION 至少一个（但不能同时存在）
// - VTODO: DTSTART 或 DUE 至少一个
// - VJOURNAL: DTSTART 必需
// - VFREEBUSY: DTSTART 和 DTEND 都必需
type CreateCalendarItemRequest struct {
	Type            CalendarItemType `json:"type" binding:"required,oneof=VEVENT VTODO VJOURNAL VFREEBUSY"`
	Summary         *string          `json:"summary"`
	Description     *string          `json:"description"`
	Location        *string          `json:"location"`
	Organizer       *string          `json:"organizer"`
	DtStart         *time.Time       `json:"dtstart"`  // 根据类型可能必需
	DtEnd           *time.Time       `json:"dtend"`    // VFREEBUSY 必需，VEVENT 与 DURATION 二选一
	Due             *time.Time       `json:"due"`      // VTODO 可选（与 DTSTART 二选一）
	Duration        *string          `json:"duration"` // VEVENT 可选（与 DTEND 二选一）
	Status          *string          `json:"status"`
	Priority        *int             `json:"priority" binding:"omitempty,gte=0,lte=9"`
	PercentComplete *int             `json:"percent_complete" binding:"omitempty,gte=0,lte=100"`
	RRule           *string          `json:"rrule"`
	ExDate          []string         `json:"exdate"`
	RDate           []string         `json:"rdate"`
	Categories      []string         `json:"categories"`
	Comment         *string          `json:"comment"`
	Contact         *string          `json:"contact"`
	RelatedTo       *string          `json:"related_to"`
	Resources       []string         `json:"resources"`
	URL             *string          `json:"url"`
	Class           *string          `json:"class"`
	RawIcal         *string          `json:"raw_ical"`
	Sequence        *int             `json:"sequence"`
}

type CreateCalendarItemResponse struct {
	ID   uint             `json:"id"`
	UID  string           `json:"uid"`
	Type CalendarItemType `json:"type"`
}

// UpdateCalendarItemRequest 更新日历项请求
type UpdateCalendarItemRequest struct {
	Summary         *string    `json:"summary,omitempty"`
	Description     *string    `json:"description,omitempty"`
	Location        *string    `json:"location,omitempty"`
	Organizer       *string    `json:"organizer,omitempty"`
	DtStart         *time.Time `json:"dtstart,omitempty"`
	DtEnd           *time.Time `json:"dtend,omitempty"`
	Due             *time.Time `json:"due,omitempty"`
	Completed       *time.Time `json:"completed,omitempty"`
	Duration        *string    `json:"duration,omitempty"`
	Status          *string    `json:"status,omitempty"`
	Priority        *int       `json:"priority" binding:"omitempty,gte=0,lte=9"`
	PercentComplete *int       `json:"percent_complete" binding:"omitempty,gte=0,lte=100"`
	RRule           *string    `json:"rrule,omitempty"`
	ExDate          []string   `json:"exdate"`
	RDate           []string   `json:"rdate,omitempty"`
	Categories      []string   `json:"categories,omitempty"`
	Comment         *string    `json:"comment,omitempty"`
	Contact         *string    `json:"contact,omitempty"`
	RelatedTo       *string    `json:"related_to,omitempty"`
	Resources       []string   `json:"resources,omitempty"`
	URL             *string    `json:"url,omitempty"`
	Class           *string    `json:"class,omitempty"`
	RawIcal         *string    `json:"raw_ical,omitempty"`
	Sequence        *int       `json:"sequence,omitempty"`
}

// ListCalendarItemsRequest 列出日历项请求
type ListCalendarItemsRequest struct {
	Page      int               `form:"page" binding:"omitempty,min=1"`
	PageSize  int               `form:"page_size" binding:"omitempty,min=1,max=100"`
	StartTime *time.Time        `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime   *time.Time        `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
	Type      *CalendarItemType `form:"type"`
}

// CalendarItemListResponse 日历项列表响应
type CalendarItemListResponse struct {
	Items      []*CalendarItem `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// TimeRange 时间范围
type TimeRange struct {
	Start *time.Time `json:"start"`
	End   *time.Time `json:"end"`
}

// SearchCalendarItemsRequest 搜索日历项请求
type SearchCalendarItemsRequest struct {
	// 搜索关键字：在所有可搜索字段中搜索（summary, description, location, organizer, comment, contact, categories, resources）
	Q *string `json:"q,omitempty"`

	// 开始时间范围过滤
	DtStart *TimeRange `json:"dtstart,omitempty"`
	// 结束时间范围过滤
	DtEnd *TimeRange `json:"dtend,omitempty"`
	// 截止时间范围过滤
	Due *TimeRange `json:"due,omitempty"`
	// 完成时间范围过滤
	Completed *TimeRange `json:"completed,omitempty"`

	// 返回结果数量限制，默认20，最大100
	Limit *int `json:"limit,omitempty"`
}

// CreateValarmRequest 创建提醒请求
type CreateValarmRequest struct {
	Action      ValarmAction           `json:"action" binding:"required,oneof=DISPLAY AUDIO EMAIL"`
	Trigger     string                 `json:"trigger" binding:"required"`
	Description *string                `json:"description"`
	Summary     *string                `json:"summary"`
	Attendee    *string                `json:"attendee"`
	Duration    *string                `json:"duration"`
	RepeatCount *int                   `json:"repeat_count"`
	XProperty   map[string]interface{} `json:"x_property"`
}

// UpdateValarmRequest 更新提醒请求
type UpdateValarmRequest struct {
	Action      *ValarmAction          `json:"action" binding:"omitempty,oneof=DISPLAY AUDIO EMAIL"`
	Trigger     *string                `json:"trigger"`
	Description *string                `json:"description"`
	Summary     *string                `json:"summary"`
	Attendee    *string                `json:"attendee"`
	Duration    *string                `json:"duration"`
	RepeatCount *int                   `json:"repeat_count"`
	XProperty   map[string]interface{} `json:"x_property"`
}

type service struct {
	repo Repository
}

// NewService 创建新的服务实例
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateCalendarItem 创建日历项
func (s *service) CreateCalendarItem(userID *uint, req *CreateCalendarItemRequest) (*CreateCalendarItemResponse, error) {
	// 验证类型
	if !isValidCalendarItemType(req.Type) {
		slog.Error("无效的日历项类型", "type", req.Type)
		return nil, ErrInvalidType
	}

	// 根据 iCalendar 标准验证必需字段
	if err := validateCalendarItemRequest(req); err != nil {
		slog.Error("创建日历校验未通过", "error", err)
		return nil, err
	}

	// 生成 UID
	uid := uuid.New().String()

	// 构建日历项
	item := &CalendarItem{
		UID:             uid,
		Type:            req.Type,
		Summary:         req.Summary,
		Description:     req.Description,
		Location:        req.Location,
		Organizer:       req.Organizer,
		DtEnd:           req.DtEnd,
		Due:             req.Due,
		Duration:        req.Duration,
		Status:          req.Status,
		Priority:        req.Priority,
		PercentComplete: req.PercentComplete,
		RRule:           req.RRule,
		ExDate:          StringArray(req.ExDate),
		RDate:           StringArray(req.RDate),
		Categories:      StringArray(req.Categories),
		Comment:         req.Comment,
		Contact:         req.Contact,
		RelatedTo:       req.RelatedTo,
		Resources:       StringArray(req.Resources),
		URL:             req.URL,
		Class:           req.Class,
		RawIcal:         req.RawIcal,
		UserID:          userID,
	}

	// 设置 DtStart（已验证不为空）
	if req.DtStart != nil {
		item.DtStart = *req.DtStart
	}

	now := time.Now()
	item.LastModified = &now

	if err := s.repo.CreateCalendarItem(item); err != nil {
		return nil, fmt.Errorf("创建日历项失败: %w", err)
	}

	return &CreateCalendarItemResponse{
		ID:   item.ID,
		UID:  item.UID,
		Type: item.Type,
	}, nil
}

// GetCalendarItemByID 根据ID获取日历项
func (s *service) GetCalendarItemByID(userID *uint, id uint) (*CalendarItem, error) {
	item, err := s.repo.GetCalendarItemByID(userID, id)
	if err != nil {
		return nil, ErrCalendarItemNotFound
	}
	return item, nil
}

// GetCalendarItemByUID 根据UID获取日历项
func (s *service) GetCalendarItemByUID(userID *uint, uid string) (*CalendarItem, error) {
	item, err := s.repo.GetCalendarItemByUID(userID, uid)
	if err != nil {
		return nil, ErrCalendarItemNotFound
	}
	return item, nil
}

// UpdateCalendarItem 更新日历项
func (s *service) UpdateCalendarItem(userID *uint, id uint, req *UpdateCalendarItemRequest) (*CalendarItem, error) {
	// 先获取现有项（带用户ID过滤）
	item, err := s.repo.GetCalendarItemByID(userID, id)
	if err != nil {
		return nil, ErrCalendarItemNotFound
	}

	// 更新字段
	if req.Summary != nil {
		item.Summary = req.Summary
	}
	if req.Description != nil {
		item.Description = req.Description
	}
	if req.Location != nil {
		item.Location = req.Location
	}
	if req.Organizer != nil {
		item.Organizer = req.Organizer
	}
	if req.DtStart != nil {
		item.DtStart = *req.DtStart
	}
	if req.DtEnd != nil {
		item.DtEnd = req.DtEnd
	}
	if req.Due != nil {
		item.Due = req.Due
	}
	if req.Completed != nil {
		item.Completed = req.Completed
	}
	if req.Duration != nil {
		item.Duration = req.Duration
	}
	if req.Status != nil {
		item.Status = req.Status
	}
	if req.Priority != nil {
		item.Priority = req.Priority
	}
	if req.PercentComplete != nil {
		item.PercentComplete = req.PercentComplete
	}
	if req.RRule != nil {
		item.RRule = req.RRule
	}
	if req.ExDate != nil {
		item.ExDate = StringArray(req.ExDate)
	}
	if req.RDate != nil {
		item.RDate = StringArray(req.RDate)
	}
	if req.Categories != nil {
		item.Categories = StringArray(req.Categories)
	}
	if req.Comment != nil {
		item.Comment = req.Comment
	}
	if req.Contact != nil {
		item.Contact = req.Contact
	}
	if req.RelatedTo != nil {
		item.RelatedTo = req.RelatedTo
	}
	if req.Resources != nil {
		item.Resources = StringArray(req.Resources)
	}
	if req.URL != nil {
		item.URL = req.URL
	}
	if req.Class != nil {
		item.Class = req.Class
	}
	if req.RawIcal != nil {
		item.RawIcal = req.RawIcal
	}

	now := time.Now()
	item.LastModified = &now

	if req.Sequence != nil {
		item.Sequence = req.Sequence
	} else {
		// 自动增加序号
		if item.Sequence == nil {
			seq := 0
			item.Sequence = &seq
		} else {
			*item.Sequence++
		}
	}

	if err := s.repo.UpdateCalendarItem(userID, item); err != nil {
		return nil, fmt.Errorf("更新日历项失败: %w", err)
	}

	// 重新获取更新后的项
	updatedItem, err := s.repo.GetCalendarItemByID(userID, id)
	if err != nil {
		return nil, fmt.Errorf("获取更新后的日历项失败: %w", err)
	}

	return updatedItem, nil
}

// DeleteCalendarItem 删除日历项
func (s *service) DeleteCalendarItem(userID *uint, id uint) error {
	if err := s.repo.DeleteCalendarItem(userID, id); err != nil {
		return ErrCalendarItemNotFound
	}
	return nil
}

// ListCalendarItems 列出日历项
func (s *service) ListCalendarItems(userID *uint, req *ListCalendarItemsRequest) (*CalendarItemListResponse, error) {
	// 设置默认分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	items, total, err := s.repo.ListCalendarItems(userID, req.StartTime, req.EndTime, req.Type, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取日历项列表失败: %w", err)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &CalendarItemListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// SearchCalendarItems 搜索日历项（按匹配程度排序）
func (s *service) SearchCalendarItems(userID *uint, req *SearchCalendarItemsRequest) ([]*CalendarItem, error) {
	// 准备搜索关键字
	var q string
	if req.Q != nil {
		q = strings.TrimSpace(*req.Q)
	}

	// 收集时间范围
	timeRanges := make(map[string]TimeRange)
	if req.DtStart != nil {
		timeRanges["dtstart"] = *req.DtStart
	}
	if req.DtEnd != nil {
		timeRanges["dtend"] = *req.DtEnd
	}
	if req.Due != nil {
		timeRanges["due"] = *req.Due
	}
	if req.Completed != nil {
		timeRanges["completed"] = *req.Completed
	}

	// 验证：至少需要指定搜索关键字或时间范围
	if q == "" && len(timeRanges) == 0 {
		return nil, fmt.Errorf("%w: 至少需要指定搜索关键字(q)或时间范围", ErrInvalidInput)
	}

	// 验证并设置返回数量限制
	limit := 20 // 默认值
	if req.Limit != nil {
		if *req.Limit < 1 {
			return nil, fmt.Errorf("%w: 返回数量限制必须大于0", ErrInvalidInput)
		}
		if *req.Limit > 100 {
			limit = 100 // 最大限制
		} else {
			limit = *req.Limit
		}
	}

	// 调用Repository层进行搜索
	items, err := s.repo.SearchCalendarItems(userID, q, timeRanges, limit)
	if err != nil {
		return nil, fmt.Errorf("搜索日历项失败: %w", err)
	}

	return items, nil
}

// CreateValarm 创建提醒
func (s *service) CreateValarm(calendarItemID uint, req *CreateValarmRequest) (*Valarm, error) {
	// 验证日历项是否存在（不验证用户ID，因为创建提醒时可能不需要用户验证）
	_, err := s.repo.GetCalendarItemByID(nil, calendarItemID)
	if err != nil {
		return nil, ErrCalendarItemNotFound
	}

	// 验证动作类型
	if !isValidValarmAction(req.Action) {
		return nil, ErrInvalidAction
	}

	// DISPLAY 类型需要 description
	if req.Action == ValarmActionDisplay && (req.Description == nil || *req.Description == "") {
		return nil, fmt.Errorf("%w: DISPLAY 类型需要 description", ErrInvalidInput)
	}

	alarm := &Valarm{
		CalendarItemID: calendarItemID,
		Action:         req.Action,
		Trigger:        req.Trigger,
		Description:    req.Description,
		Summary:        req.Summary,
		Attendee:       req.Attendee,
		Duration:       req.Duration,
		RepeatCount:    req.RepeatCount,
	}

	if req.XProperty != nil {
		alarm.XProperty = JSONB(req.XProperty)
	}

	if err := s.repo.CreateValarm(alarm); err != nil {
		return nil, fmt.Errorf("创建提醒失败: %w", err)
	}

	return alarm, nil
}

// GetValarmByID 根据ID获取提醒
func (s *service) GetValarmByID(id uint) (*Valarm, error) {
	alarm, err := s.repo.GetValarmByID(id)
	if err != nil {
		return nil, ErrValarmNotFound
	}
	return alarm, nil
}

// GetValarmsByCalendarItemID 根据日历项ID获取所有提醒
func (s *service) GetValarmsByCalendarItemID(calendarItemID uint) ([]*Valarm, error) {
	// 验证日历项是否存在（不验证用户ID，因为获取提醒时可能不需要用户验证）
	_, err := s.repo.GetCalendarItemByID(nil, calendarItemID)
	if err != nil {
		return nil, ErrCalendarItemNotFound
	}

	alarms, err := s.repo.GetValarmsByCalendarItemID(calendarItemID)
	if err != nil {
		return nil, fmt.Errorf("获取提醒列表失败: %w", err)
	}

	return alarms, nil
}

// UpdateValarm 更新提醒
func (s *service) UpdateValarm(id uint, req *UpdateValarmRequest) (*Valarm, error) {
	alarm, err := s.repo.GetValarmByID(id)
	if err != nil {
		return nil, ErrValarmNotFound
	}

	if req.Action != nil {
		if !isValidValarmAction(*req.Action) {
			return nil, ErrInvalidAction
		}
		alarm.Action = *req.Action
	}
	if req.Trigger != nil {
		alarm.Trigger = *req.Trigger
	}
	if req.Description != nil {
		alarm.Description = req.Description
	}
	if req.Summary != nil {
		alarm.Summary = req.Summary
	}
	if req.Attendee != nil {
		alarm.Attendee = req.Attendee
	}
	if req.Duration != nil {
		alarm.Duration = req.Duration
	}
	if req.RepeatCount != nil {
		alarm.RepeatCount = req.RepeatCount
	}
	if req.XProperty != nil {
		alarm.XProperty = JSONB(req.XProperty)
	}

	if err := s.repo.UpdateValarm(alarm); err != nil {
		return nil, fmt.Errorf("更新提醒失败: %w", err)
	}

	return alarm, nil
}

// DeleteValarm 删除提醒
func (s *service) DeleteValarm(id uint) error {
	_, err := s.repo.GetValarmByID(id)
	if err != nil {
		return ErrValarmNotFound
	}

	if err := s.repo.DeleteValarm(id); err != nil {
		return fmt.Errorf("删除提醒失败: %w", err)
	}

	return nil
}

// isValidCalendarItemType 验证日历项类型
func isValidCalendarItemType(t CalendarItemType) bool {
	return t == CalendarItemTypeEvent || t == CalendarItemTypeTodo ||
		t == CalendarItemTypeJournal || t == CalendarItemTypeFreeBusy
}

// isValidValarmAction 验证提醒动作类型
func isValidValarmAction(a ValarmAction) bool {
	return a == ValarmActionDisplay || a == ValarmActionAudio || a == ValarmActionEmail
}

// validateCalendarItemRequest 根据 iCalendar 标准验证日历项请求
// RFC 5545 要求：
// - VEVENT: DTSTART 必需，DTEND 或 DURATION 至少一个（但不能同时存在）
// - VTODO: DTSTART 或 DUE 至少一个
// - VJOURNAL: DTSTART 必需
// - VFREEBUSY: DTSTART 和 DTEND 都必需
func validateCalendarItemRequest(req *CreateCalendarItemRequest) error {
	switch req.Type {
	case CalendarItemTypeEvent:
		// VEVENT: DTSTART 必需
		if req.DtStart == nil {
			return fmt.Errorf("%w: VEVENT 类型需要 dtstart", ErrInvalidInput)
		}
		// DTEND 或 DURATION 至少一个，但不能同时存在
		hasDtEnd := req.DtEnd != nil
		hasDuration := req.Duration != nil && *req.Duration != ""
		if !hasDtEnd && !hasDuration {
			return fmt.Errorf("%w: VEVENT 类型需要 dtend 或 duration 至少一个", ErrInvalidInput)
		}
		if hasDtEnd && hasDuration {
			return fmt.Errorf("%w: VEVENT 类型不能同时指定 dtend 和 duration", ErrInvalidInput)
		}

	case CalendarItemTypeTodo:
		// VTODO: DTSTART 或 DUE 至少一个
		if req.DtStart == nil && req.Due == nil {
			return fmt.Errorf("%w: VTODO 类型需要 dtstart 或 due 至少一个", ErrInvalidInput)
		}

	case CalendarItemTypeJournal:
		// VJOURNAL: DTSTART 必需
		if req.DtStart == nil {
			return fmt.Errorf("%w: VJOURNAL 类型需要 dtstart", ErrInvalidInput)
		}

	case CalendarItemTypeFreeBusy:
		// VFREEBUSY: DTSTART 和 DTEND 都必需
		if req.DtStart == nil {
			return fmt.Errorf("%w: VFREEBUSY 类型需要 dtstart", ErrInvalidInput)
		}
		if req.DtEnd == nil {
			return fmt.Errorf("%w: VFREEBUSY 类型需要 dtend", ErrInvalidInput)
		}
	}

	return nil
}
