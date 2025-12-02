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
	SearchCalendarItemsByFieldKeywords(userID *uint, fieldKeywords map[string]string, timeRanges map[string]TimeRange) ([]*CalendarItem, error)

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

// fieldColumnMap 字段名到 SQL 列名的映射
// 对于 JSONB 字段，列名包含 ::text 转换
var fieldColumnMap = map[string]string{
	"summary":     "summary",
	"description": "description",
	"location":    "location",
	"organizer":   "organizer",
	"comment":     "comment",
	"contact":     "contact",
	"categories":  "categories::text",
	"resources":   "resources::text",
}

// getFieldColumn 获取字段对应的 SQL 列名
func getFieldColumn(field string) (string, bool) {
	column, ok := fieldColumnMap[field]
	return column, ok
}

// getFieldMatchScoreSQL 获取字段匹配分数的 SQL 表达式
func getFieldMatchScoreSQL(field string, keywordExact, keywordPrefix string) string {
	column, ok := getFieldColumn(field)
	if !ok {
		return "1"
	}

	return fmt.Sprintf(`CASE 
		WHEN %s ILIKE '%s' THEN 3
		WHEN %s ILIKE '%s' THEN 2
		ELSE 1
	END`, column, keywordExact, column, keywordPrefix)
}

// isValidSearchField 验证搜索字段是否有效
func isValidSearchField(field string) bool {
	_, ok := fieldColumnMap[field]
	return ok
}

// SearchCalendarItemsByFieldKeywords 根据字段关键字映射搜索日历项（最多返回20条，按匹配程度排序）
// 支持多个字段的 AND 搜索，每个字段匹配对应的关键字
// 支持多个时间字段的范围过滤
// 允许只有时间范围而没有字段关键字
func (r *repository) SearchCalendarItemsByFieldKeywords(userID *uint, fieldKeywords map[string]string, timeRanges map[string]TimeRange) ([]*CalendarItem, error) {
	// 验证：至少需要指定一个搜索字段或时间范围
	if len(fieldKeywords) == 0 && len(timeRanges) == 0 {
		return nil, fmt.Errorf("至少需要指定一个搜索字段或时间范围")
	}

	// 验证所有字段是否有效，并去重
	validFieldKeywords := make(map[string]string)
	invalidFields := make([]string, 0)

	for field, keyword := range fieldKeywords {
		if !isValidSearchField(field) {
			invalidFields = append(invalidFields, field)
			continue
		}

		if keyword == "" {
			return nil, fmt.Errorf("关键字不能为空")
		}
		validFieldKeywords[field] = keyword
	}

	// 如果有无效字段，返回错误
	if len(invalidFields) > 0 {
		return nil, fmt.Errorf("不支持的搜索字段: %v", invalidFields)
	}

	// 如果指定了字段关键字但没有有效字段，且没有时间范围，则返回错误
	if len(validFieldKeywords) == 0 && len(timeRanges) == 0 {
		return nil, fmt.Errorf("没有有效的搜索字段或时间范围")
	}

	var items []*CalendarItem
	query := r.db.Model(&CalendarItem{})

	// 过滤用户ID
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 时间范围过滤：支持多个时间字段的范围搜索
	// 字段名到数据库列名的映射
	timeFieldColumnMap := map[string]string{
		"dtstart":   "dt_start",
		"dtend":     "dt_end",
		"due":       "due",
		"completed": "completed",
	}

	for field, timeRange := range timeRanges {
		column, ok := timeFieldColumnMap[field]
		if !ok {
			continue // 跳过无效的时间字段
		}

		// 处理开始时间
		if timeRange.Start != nil {
			// 对于可空字段（dtend, due, completed），需要考虑 NULL 值
			if field == "dtstart" {
				query = query.Where(column+" >= ?", *timeRange.Start)
			} else {
				query = query.Where("("+column+" >= ? OR "+column+" IS NULL)", *timeRange.Start)
			}
		}

		// 处理结束时间
		if timeRange.End != nil {
			// 对于可空字段（dtend, due, completed），需要考虑 NULL 值
			if field == "dtstart" {
				query = query.Where(column+" <= ?", *timeRange.End)
			} else {
				query = query.Where("("+column+" <= ? OR "+column+" IS NULL)", *timeRange.End)
			}
		}
	}

	// 构建 AND 条件：每个字段都必须匹配对应的关键字
	for field, keyword := range validFieldKeywords {
		column, ok := getFieldColumn(field)
		if !ok {
			// 理论上不会到达这里，因为已经验证过了，但为了安全起见保留
			return nil, fmt.Errorf("不支持的搜索字段: %s", field)
		}

		keywordEscaped := escapeSQLString(keyword)
		keywordPattern := "%" + keywordEscaped + "%"
		query = query.Where(column+" ILIKE ?", keywordPattern)
	}

	// 构建排序：按所有字段的平均匹配程度排序
	var scoreExpressions []string
	for field, keyword := range validFieldKeywords {
		keywordEscaped := escapeSQLString(keyword)
		keywordExact := keywordEscaped
		keywordPrefix := keywordEscaped + "%"
		scoreSQL := getFieldMatchScoreSQL(field, keywordExact, keywordPrefix)
		if scoreSQL != "1" {
			scoreExpressions = append(scoreExpressions, scoreSQL)
		}
	}

	var orderBy string
	if len(scoreExpressions) == 1 {
		// 单个字段，直接使用其匹配分数
		orderBy = fmt.Sprintf("%s DESC", scoreExpressions[0])
	} else if len(scoreExpressions) > 1 {
		// 多个字段，使用平均匹配分数
		avgScore := fmt.Sprintf("(%s) / %d", strings.Join(scoreExpressions, " + "), len(scoreExpressions))
		orderBy = fmt.Sprintf("%s DESC", avgScore)
	} else {
		// 如果没有有效的排序表达式，使用默认排序
		orderBy = "created_at DESC"
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
