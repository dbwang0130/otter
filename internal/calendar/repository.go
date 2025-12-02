package calendar

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Repository 日历项仓库接口
type Repository interface {
	// CalendarItem 相关方法
	CreateCalendarItem(item *CalendarItem) error
	GetCalendarItemByID(id uint) (*CalendarItem, error)
	GetCalendarItemByUID(uid string) (*CalendarItem, error)
	UpdateCalendarItem(item *CalendarItem) error
	DeleteCalendarItem(id uint) error
	ListCalendarItems(userID *uint, startTime, endTime *time.Time, itemType *CalendarItemType, offset, limit int) ([]*CalendarItem, int64, error)
	SearchCalendarItemsByKeyword(userID *uint, fields []string, keyword string) ([]*CalendarItem, error)

	// Valarm 相关方法
	CreateValarm(alarm *Valarm) error
	GetValarmByID(id uint) (*Valarm, error)
	GetValarmsByCalendarItemID(calendarItemID uint) ([]*Valarm, error)
	UpdateValarm(alarm *Valarm) error
	DeleteValarm(id uint) error
	DeleteValarmsByCalendarItemID(calendarItemID uint) error
}

type repository struct {
	db *gorm.DB
}

// NewRepository 创建新的仓库实例
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// CreateCalendarItem 创建日历项
func (r *repository) CreateCalendarItem(item *CalendarItem) error {
	return r.db.Create(item).Error
}

// GetCalendarItemByID 根据ID获取日历项
func (r *repository) GetCalendarItemByID(id uint) (*CalendarItem, error) {
	var item CalendarItem
	if err := r.db.Preload("Alarms").First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// GetCalendarItemByUID 根据UID获取日历项
func (r *repository) GetCalendarItemByUID(uid string) (*CalendarItem, error) {
	var item CalendarItem
	if err := r.db.Preload("Alarms").Where("uid = ?", uid).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateCalendarItem 更新日历项
func (r *repository) UpdateCalendarItem(item *CalendarItem) error {
	return r.db.Save(item).Error
}

// DeleteCalendarItem 删除日历项（软删除）
func (r *repository) DeleteCalendarItem(id uint) error {
	return r.db.Delete(&CalendarItem{}, id).Error
}

// ListCalendarItems 列出日历项
func (r *repository) ListCalendarItems(userID *uint, startTime, endTime *time.Time, itemType *CalendarItemType, offset, limit int) ([]*CalendarItem, int64, error) {
	var items []*CalendarItem
	var total int64

	query := r.db.Model(&CalendarItem{})

	// 过滤用户ID
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 过滤类型
	if itemType != nil {
		query = query.Where("type = ?", *itemType)
	}

	// 时间范围过滤
	if startTime != nil && endTime != nil {
		// 查找在时间范围内有重叠的日历项
		// dt_start <= endTime AND (dt_end >= startTime OR dt_end IS NULL)
		query = query.Where("dt_start <= ? AND (dt_end >= ? OR dt_end IS NULL)", *endTime, *startTime)
	} else if startTime != nil {
		query = query.Where("dt_start >= ?", *startTime)
	} else if endTime != nil {
		query = query.Where("dt_start <= ?", *endTime)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	if err := query.Preload("Alarms").Offset(offset).Limit(limit).Order("dt_start ASC").Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// escapeSQLString 转义 SQL 字符串中的单引号，防止 SQL 注入
func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// getFieldMatchScoreSQL 获取字段匹配分数的 SQL 表达式
func getFieldMatchScoreSQL(field string, keywordExact, keywordPrefix string) string {
	switch field {
	case "summary":
		return fmt.Sprintf(`CASE 
			WHEN summary ILIKE '%s' THEN 3
			WHEN summary ILIKE '%s' THEN 2
			ELSE 1
		END`, keywordExact, keywordPrefix)
	case "description":
		return fmt.Sprintf(`CASE 
			WHEN description ILIKE '%s' THEN 3
			WHEN description ILIKE '%s' THEN 2
			ELSE 1
		END`, keywordExact, keywordPrefix)
	case "location":
		return fmt.Sprintf(`CASE 
			WHEN location ILIKE '%s' THEN 3
			WHEN location ILIKE '%s' THEN 2
			ELSE 1
		END`, keywordExact, keywordPrefix)
	case "organizer":
		return fmt.Sprintf(`CASE 
			WHEN organizer ILIKE '%s' THEN 3
			WHEN organizer ILIKE '%s' THEN 2
			ELSE 1
		END`, keywordExact, keywordPrefix)
	case "comment":
		return fmt.Sprintf(`CASE 
			WHEN comment ILIKE '%s' THEN 3
			WHEN comment ILIKE '%s' THEN 2
			ELSE 1
		END`, keywordExact, keywordPrefix)
	case "contact":
		return fmt.Sprintf(`CASE 
			WHEN contact ILIKE '%s' THEN 3
			WHEN contact ILIKE '%s' THEN 2
			ELSE 1
		END`, keywordExact, keywordPrefix)
	case "categories":
		return fmt.Sprintf(`CASE 
			WHEN categories::text ILIKE '%s' THEN 3
			WHEN categories::text ILIKE '%s' THEN 2
			ELSE 1
		END`, keywordExact, keywordPrefix)
	case "resources":
		return fmt.Sprintf(`CASE 
			WHEN resources::text ILIKE '%s' THEN 3
			WHEN resources::text ILIKE '%s' THEN 2
			ELSE 1
		END`, keywordExact, keywordPrefix)
	default:
		return "1"
	}
}

// isValidSearchField 验证搜索字段是否有效
func isValidSearchField(field string) bool {
	switch field {
	case "summary", "description", "location", "organizer",
		"comment", "contact", "categories", "resources":
		return true
	default:
		return false
	}
}

// SearchCalendarItemsByKeyword 根据关键字模糊搜索日历项（最多返回20条，按匹配程度排序）
// 支持多个字段的 AND 搜索，所有指定字段都必须匹配关键字
func (r *repository) SearchCalendarItemsByKeyword(userID *uint, fields []string, keyword string) ([]*CalendarItem, error) {
	if keyword == "" {
		return nil, fmt.Errorf("关键字不能为空")
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("至少需要指定一个搜索字段")
	}

	// 验证所有字段是否有效
	validFields := make([]string, 0, len(fields))
	for _, field := range fields {
		if !isValidSearchField(field) {
			return nil, fmt.Errorf("不支持的搜索字段: %s", field)
		}
		// 去重：避免重复字段
		duplicate := false
		for _, vf := range validFields {
			if vf == field {
				duplicate = true
				break
			}
		}
		if !duplicate {
			validFields = append(validFields, field)
		}
	}

	if len(validFields) == 0 {
		return nil, fmt.Errorf("没有有效的搜索字段")
	}

	var items []*CalendarItem
	keywordEscaped := escapeSQLString(keyword)
	keywordPattern := "%" + keywordEscaped + "%"
	keywordExact := keywordEscaped
	keywordPrefix := keywordEscaped + "%"

	query := r.db.Model(&CalendarItem{})

	// 过滤用户ID
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 构建 AND 条件：所有指定字段都必须匹配
	// 使用 GORM 的参数化查询，每个字段都要匹配关键字
	for _, field := range validFields {
		switch field {
		case "summary":
			query = query.Where("summary ILIKE ?", keywordPattern)
		case "description":
			query = query.Where("description ILIKE ?", keywordPattern)
		case "location":
			query = query.Where("location ILIKE ?", keywordPattern)
		case "organizer":
			query = query.Where("organizer ILIKE ?", keywordPattern)
		case "comment":
			query = query.Where("comment ILIKE ?", keywordPattern)
		case "contact":
			query = query.Where("contact ILIKE ?", keywordPattern)
		case "categories":
			query = query.Where("categories::text ILIKE ?", keywordPattern)
		case "resources":
			query = query.Where("resources::text ILIKE ?", keywordPattern)
		default:
			// 理论上不会到达这里，因为已经验证过了，但为了安全起见保留
			return nil, fmt.Errorf("不支持的搜索字段: %s", field)
		}
	}

	// 构建排序：按所有字段的平均匹配程度排序
	var scoreExpressions []string
	for _, field := range validFields {
		scoreSQL := getFieldMatchScoreSQL(field, keywordExact, keywordPrefix)
		if scoreSQL != "1" {
			scoreExpressions = append(scoreExpressions, scoreSQL)
		}
	}

	var orderBy string
	if len(scoreExpressions) == 1 {
		// 单个字段，直接使用其匹配分数
		orderBy = fmt.Sprintf("%s DESC", scoreExpressions[0])
	} else {
		// 多个字段，使用平均匹配分数
		avgScore := fmt.Sprintf("(%s) / %d", strings.Join(scoreExpressions, " + "), len(scoreExpressions))
		orderBy = fmt.Sprintf("%s DESC", avgScore)
	}

	// 添加排序
	query = query.Order(orderBy)

	// 查询列表，最多返回20条
	if err := query.Preload("Alarms").Limit(20).Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

// CreateValarm 创建提醒
func (r *repository) CreateValarm(alarm *Valarm) error {
	return r.db.Create(alarm).Error
}

// GetValarmByID 根据ID获取提醒
func (r *repository) GetValarmByID(id uint) (*Valarm, error) {
	var alarm Valarm
	if err := r.db.First(&alarm, id).Error; err != nil {
		return nil, err
	}
	return &alarm, nil
}

// GetValarmsByCalendarItemID 根据日历项ID获取所有提醒
func (r *repository) GetValarmsByCalendarItemID(calendarItemID uint) ([]*Valarm, error) {
	var alarms []*Valarm
	if err := r.db.Where("calendar_item_id = ?", calendarItemID).Find(&alarms).Error; err != nil {
		return nil, err
	}
	return alarms, nil
}

// UpdateValarm 更新提醒
func (r *repository) UpdateValarm(alarm *Valarm) error {
	return r.db.Save(alarm).Error
}

// DeleteValarm 删除提醒（软删除）
func (r *repository) DeleteValarm(id uint) error {
	return r.db.Delete(&Valarm{}, id).Error
}

// DeleteValarmsByCalendarItemID 删除日历项的所有提醒
func (r *repository) DeleteValarmsByCalendarItemID(calendarItemID uint) error {
	return r.db.Where("calendar_item_id = ?", calendarItemID).Delete(&Valarm{}).Error
}
