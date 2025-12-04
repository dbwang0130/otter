package calendar

import (
	"time"

	"github.com/galilio/otter/internal/calendar"
)

// OperationResult result of calendar item operations (create, update, delete)
type OperationResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	ID      *uint       `json:"id,omitempty"`
	UID     *string     `json:"uid,omitempty"`
	Created bool        `json:"created,omitempty"`
	Updated bool        `json:"updated,omitempty"`
	Deleted bool        `json:"deleted,omitempty"`
	Item    *ItemDetail `json:"item,omitempty"`
}

// CreateRequest create calendar item request
// Required fields (RFC 5545): type (VEVENT/VTODO/VJOURNAL/VFREEBUSY)
//   - VEVENT: dtstart required, dtend or duration (not both)
//   - VTODO: dtstart or due
//   - VJOURNAL: dtstart required
//   - VFREEBUSY: both dtstart and dtend required
//
// Optional: uid (for idempotency), all other fields
// Time format: RFC3339, e.g. "2024-01-15T14:30:00Z"
type CreateRequest struct {
	UID             *string  `json:"uid,omitempty"`                                                 // 唯一标识符（可选，用于幂等性：如果提供且已存在则返回现有项）
	Type            string   `json:"type" binding:"required,oneof=VEVENT VTODO VJOURNAL VFREEBUSY"` // 日历项类型（必填）
	DtStart         *string  `json:"dtstart,omitempty"`                                             // 开始时间，RFC3339 格式，例如: "2024-01-15T14:30:00Z"
	DtEnd           *string  `json:"dtend,omitempty"`                                               // 结束时间，RFC3339 格式
	Due             *string  `json:"due,omitempty"`                                                 // 截止时间（VTODO），RFC3339 格式
	Duration        *string  `json:"duration,omitempty"`                                            // 持续时间（VEVENT），与 dtend 二选一，格式如 "PT1H30M"
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

// GetRequest get calendar item request
// Requires at least one of id or uid
type GetRequest struct {
	ID  *uint   `json:"id,omitempty"`
	UID *string `json:"uid,omitempty"`
}

// UpdateRequest update calendar item request
type UpdateRequest struct {
	ID uint `json:"id" binding:"required"`
	calendar.UpdateCalendarItemRequest
}

// DeleteRequest delete calendar item request
// Idempotent: returns success even if item doesn't exist
type DeleteRequest struct {
	ID uint `json:"id" binding:"required"`
}

// SearchRequest search calendar items request
// Requires at least one of q or dtstart
// Search fields: summary, description, location, organizer, comment, contact, categories, resources
type SearchRequest struct {
	Q       *string    `json:"q,omitempty"`
	DtStart *TimeRange `json:"dtstart,omitempty"`
	Limit   *int       `json:"limit,omitempty"`
}

// TimeRange time range filter for dtstart field
type TimeRange struct {
	Start *string `json:"start,omitempty"`
	End   *string `json:"end,omitempty"`
}

// Item calendar item summary for list display
type Item struct {
	ID              uint                      `json:"id"`
	Type            calendar.CalendarItemType `json:"type,omitempty"`
	Summary         *string                   `json:"summary,omitempty"`
	Description     *string                   `json:"description,omitempty"`
	Location        *string                   `json:"location,omitempty"`
	Organizer       *string                   `json:"organizer,omitempty"`
	DtStart         time.Time                 `json:"dtstart,omitempty"`
	DtEnd           *time.Time                `json:"dtend,omitempty"`
	Due             *time.Time                `json:"due,omitempty"`
	Completed       *time.Time                `json:"completed,omitempty"`
	Status          *string                   `json:"status,omitempty"`
	Priority        *int                      `json:"priority,omitempty"`
	PercentComplete *int                      `json:"percent_complete,omitempty"`
	Categories      []string                  `json:"categories,omitempty"`
}

// ItemDetail calendar item detail response
// Embeds Item to reuse all base fields
type ItemDetail struct {
	Item

	UID          string            `json:"uid"`
	Duration     *string           `json:"duration,omitempty"`
	Sequence     *int              `json:"sequence,omitempty"`
	RRule        *string           `json:"rrule,omitempty"`
	ExDate       []string          `json:"exdate,omitempty"`
	RDate        []string          `json:"rdate,omitempty"`
	Comment      *string           `json:"comment,omitempty"`
	Contact      *string           `json:"contact,omitempty"`
	RelatedTo    *string           `json:"related_to,omitempty"`
	Resources    []string          `json:"resources,omitempty"`
	URL          *string           `json:"url,omitempty"`
	Class        *string           `json:"class,omitempty"`
	LastModified *time.Time        `json:"last_modified,omitempty"`
	CreatedAt    time.Time         `json:"created_at,omitempty"`
	UpdatedAt    time.Time         `json:"updated_at,omitempty"`
	Alarms       []calendar.Valarm `json:"alarms,omitempty"`
}

// SearchResponse search calendar items response
type SearchResponse struct {
	Items   []*Item `json:"items"`
	Total   int     `json:"total"`
	Limit   *int    `json:"limit,omitempty"`
	HasMore bool    `json:"has_more"`
}
