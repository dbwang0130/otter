package calendar

import (
	"errors"
	"fmt"
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
	CreateCalendarItem(userID *uint, req *CreateCalendarItemRequest) (*CalendarItem, error)
	GetCalendarItemByID(id uint) (*CalendarItem, error)
	GetCalendarItemByUID(uid string) (*CalendarItem, error)
	UpdateCalendarItem(id uint, req *UpdateCalendarItemRequest) (*CalendarItem, error)
	DeleteCalendarItem(id uint) error
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
type CreateCalendarItemRequest struct {
	Type            CalendarItemType `json:"type" binding:"required,oneof=VEVENT VTODO VJOURNAL VFREEBUSY"`
	Summary         *string          `json:"summary"`
	Description     *string          `json:"description"`
	Location        *string          `json:"location"`
	Organizer       *string          `json:"organizer"`
	DtStart         time.Time        `json:"dtstart" binding:"required"`
	DtEnd           *time.Time       `json:"dtend"`
	Due             *time.Time       `json:"due"`
	Duration        *string          `json:"duration"`
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

// UpdateCalendarItemRequest 更新日历项请求
type UpdateCalendarItemRequest struct {
	Summary         *string    `json:"summary"`
	Description     *string    `json:"description"`
	Location        *string    `json:"location"`
	Organizer       *string    `json:"organizer"`
	DtStart         *time.Time `json:"dtstart"`
	DtEnd           *time.Time `json:"dtend"`
	Due             *time.Time `json:"due"`
	Completed       *time.Time `json:"completed"`
	Duration        *string    `json:"duration"`
	Status          *string    `json:"status"`
	Priority        *int       `json:"priority" binding:"omitempty,gte=0,lte=9"`
	PercentComplete *int       `json:"percent_complete" binding:"omitempty,gte=0,lte=100"`
	RRule           *string    `json:"rrule"`
	ExDate          []string   `json:"exdate"`
	RDate           []string   `json:"rdate"`
	Categories      []string   `json:"categories"`
	Comment         *string    `json:"comment"`
	Contact         *string    `json:"contact"`
	RelatedTo       *string    `json:"related_to"`
	Resources       []string   `json:"resources"`
	URL             *string    `json:"url"`
	Class           *string    `json:"class"`
	RawIcal         *string    `json:"raw_ical"`
	Sequence        *int       `json:"sequence"`
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

// SearchCalendarItemsRequest 搜索日历项请求
type SearchCalendarItemsRequest struct {
	Fields  []string `form:"fields" binding:"required,min=1,dive,oneof=summary description location organizer comment contact categories resources"`
	Keyword string   `form:"keyword" binding:"required,min=1"`
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
func (s *service) CreateCalendarItem(userID *uint, req *CreateCalendarItemRequest) (*CalendarItem, error) {
	// 验证类型
	if !isValidCalendarItemType(req.Type) {
		return nil, ErrInvalidType
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
		DtStart:         req.DtStart,
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

	now := time.Now()
	item.LastModified = &now

	if err := s.repo.CreateCalendarItem(item); err != nil {
		return nil, fmt.Errorf("创建日历项失败: %w", err)
	}

	return item, nil
}

// GetCalendarItemByID 根据ID获取日历项
func (s *service) GetCalendarItemByID(id uint) (*CalendarItem, error) {
	item, err := s.repo.GetCalendarItemByID(id)
	if err != nil {
		return nil, ErrCalendarItemNotFound
	}
	return item, nil
}

// GetCalendarItemByUID 根据UID获取日历项
func (s *service) GetCalendarItemByUID(uid string) (*CalendarItem, error) {
	item, err := s.repo.GetCalendarItemByUID(uid)
	if err != nil {
		return nil, ErrCalendarItemNotFound
	}
	return item, nil
}

// UpdateCalendarItem 更新日历项
func (s *service) UpdateCalendarItem(id uint, req *UpdateCalendarItemRequest) (*CalendarItem, error) {
	item, err := s.repo.GetCalendarItemByID(id)
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

	if err := s.repo.UpdateCalendarItem(item); err != nil {
		return nil, fmt.Errorf("更新日历项失败: %w", err)
	}

	return item, nil
}

// DeleteCalendarItem 删除日历项
func (s *service) DeleteCalendarItem(id uint) error {
	_, err := s.repo.GetCalendarItemByID(id)
	if err != nil {
		return ErrCalendarItemNotFound
	}

	if err := s.repo.DeleteCalendarItem(id); err != nil {
		return fmt.Errorf("删除日历项失败: %w", err)
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

// SearchCalendarItems 搜索日历项（最多返回20条，按匹配程度排序）
func (s *service) SearchCalendarItems(userID *uint, req *SearchCalendarItemsRequest) ([]*CalendarItem, error) {
	// 验证搜索字段
	if len(req.Fields) == 0 {
		return nil, fmt.Errorf("%w: 至少需要指定一个搜索字段", ErrInvalidInput)
	}

	// 验证每个字段是否有效，并去重
	validFields := make([]string, 0, len(req.Fields))
	fieldSet := make(map[string]bool)
	invalidFields := make([]string, 0)

	for _, fieldStr := range req.Fields {
		field := SearchableField(fieldStr)
		if !field.IsValid() {
			invalidFields = append(invalidFields, fieldStr)
			continue
		}
		// 去重
		if !fieldSet[fieldStr] {
			validFields = append(validFields, fieldStr)
			fieldSet[fieldStr] = true
		}
	}

	// 如果有无效字段，返回错误
	if len(invalidFields) > 0 {
		return nil, fmt.Errorf("%w: 无效的搜索字段: %v", ErrInvalidSearchField, invalidFields)
	}

	// 验证去重后是否还有有效字段
	if len(validFields) == 0 {
		return nil, fmt.Errorf("%w: 没有有效的搜索字段", ErrInvalidInput)
	}

	// 验证关键字
	if req.Keyword == "" {
		return nil, fmt.Errorf("%w: 关键字不能为空", ErrInvalidInput)
	}

	items, err := s.repo.SearchCalendarItemsByKeyword(userID, validFields, req.Keyword)
	if err != nil {
		return nil, fmt.Errorf("搜索日历项失败: %w", err)
	}

	return items, nil
}

// CreateValarm 创建提醒
func (s *service) CreateValarm(calendarItemID uint, req *CreateValarmRequest) (*Valarm, error) {
	// 验证日历项是否存在
	_, err := s.repo.GetCalendarItemByID(calendarItemID)
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
	// 验证日历项是否存在
	_, err := s.repo.GetCalendarItemByID(calendarItemID)
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
