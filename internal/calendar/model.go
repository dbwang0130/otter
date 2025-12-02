package calendar

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// CalendarItemType 日历项类型
type CalendarItemType string

const (
	CalendarItemTypeEvent    CalendarItemType = "VEVENT"
	CalendarItemTypeTodo     CalendarItemType = "VTODO"
	CalendarItemTypeJournal  CalendarItemType = "VJOURNAL"
	CalendarItemTypeFreeBusy CalendarItemType = "VFREEBUSY"
)

// CalendarItem 日历项模型
type CalendarItem struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UID             string           `json:"uid" gorm:"not null;size:255;index:idx_uid,unique"`
	Type            CalendarItemType `json:"type" gorm:"not null;type:varchar(20);check:type IN ('VEVENT','VTODO','VJOURNAL','VFREEBUSY')"`
	Summary         *string          `json:"summary" gorm:"size:500"`
	Description     *string          `json:"description" gorm:"type:text"`
	Location        *string          `json:"location" gorm:"size:500"`
	Organizer       *string          `json:"organizer" gorm:"size:500"`
	DtStart         time.Time        `json:"dtstart" gorm:"not null;index"`
	DtEnd           *time.Time       `json:"dtend" gorm:"index"`
	Due             *time.Time       `json:"due" gorm:"index"`
	Completed       *time.Time       `json:"completed"`
	Duration        *string          `json:"duration" gorm:"size:100"`
	Status          *string          `json:"status" gorm:"size:50"`
	Priority        *int             `json:"priority" gorm:"check:priority >= 0 AND priority <= 9"`
	PercentComplete *int             `json:"percent_complete" gorm:"check:percent_complete >= 0 AND percent_complete <= 100"`
	Sequence        *int             `json:"sequence"`
	RRule           *string          `json:"rrule" gorm:"size:500"`
	ExDate          StringArray      `json:"exdate" gorm:"type:jsonb"`
	RDate           StringArray      `json:"rdate" gorm:"type:jsonb"`
	Categories      StringArray      `json:"categories" gorm:"type:jsonb"`
	Comment         *string          `json:"comment" gorm:"type:text"`
	Contact         *string          `json:"contact" gorm:"size:500"`
	RelatedTo       *string          `json:"related_to" gorm:"size:255"`
	Resources       StringArray      `json:"resources" gorm:"type:jsonb"`
	URL             *string          `json:"url" gorm:"size:1000"`
	Class           *string          `json:"class" gorm:"size:50"`
	LastModified    *time.Time       `json:"last_modified"`
	RawIcal         *string          `json:"raw_ical" gorm:"type:text"`

	// 关联用户（如果需要）
	UserID *uint `json:"user_id" gorm:"index"`

	// 关联的提醒
	Alarms []Valarm `json:"alarms" gorm:"foreignKey:CalendarItemID;constraint:OnDelete:CASCADE"`
}

func (CalendarItem) TableName() string {
	return "calendar_items"
}

// ValarmAction 提醒动作类型
type ValarmAction string

const (
	ValarmActionDisplay ValarmAction = "DISPLAY"
	ValarmActionAudio   ValarmAction = "AUDIO"
	ValarmActionEmail   ValarmAction = "EMAIL"
)

// Valarm 提醒模型
type Valarm struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	CalendarItemID uint         `json:"calendar_item_id" gorm:"not null;index"`
	Action         ValarmAction `json:"action" gorm:"not null;type:varchar(20);check:action IN ('DISPLAY','AUDIO','EMAIL')"`
	Trigger        string       `json:"trigger" gorm:"not null;size:500"`
	Description    *string      `json:"description" gorm:"size:500"`
	Summary        *string      `json:"summary" gorm:"size:500"`
	Attendee       *string      `json:"attendee" gorm:"size:500"`
	Duration       *string      `json:"duration" gorm:"size:100"`
	RepeatCount    *int         `json:"repeat_count"`
	XProperty      JSONB        `json:"x_property" gorm:"type:jsonb"`
}

func (Valarm) TableName() string {
	return "valarms"
}

// StringArray 字符串数组类型，用于 JSONB 存储
type StringArray []string

// Value 实现 driver.Valuer 接口
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(bytes, a)
}

// JSONB 通用 JSONB 类型
type JSONB map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(bytes, j)
}
