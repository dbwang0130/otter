package calendar

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Repository 日历项仓库接口
type Repository interface {
	// CalendarItem 相关方法
	CreateCalendarItem(item *CalendarItem) error
	GetCalendarItemByID(userID *uint, id uint) (*CalendarItem, error)
	GetCalendarItemByUID(userID *uint, uid string) (*CalendarItem, error)
	UpdateCalendarItem(userID *uint, item *CalendarItem) error
	DeleteCalendarItem(userID *uint, id uint) error
	ListCalendarItems(userID *uint, startTime, endTime *time.Time, itemType *CalendarItemType, offset, limit int) ([]*CalendarItem, int64, error)
	SearchCalendarItems(userID *uint, q string, timeRanges map[string]TimeRange, limit int) ([]*CalendarItem, error)

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

// GetCalendarItemByID 根据ID获取日历项（带用户ID过滤）
func (r *repository) GetCalendarItemByID(userID *uint, id uint) (*CalendarItem, error) {
	var item CalendarItem
	query := r.db.Preload("Alarms").Where("id = ?", id)

	// 过滤用户ID
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if err := query.First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// GetCalendarItemByUID 根据UID获取日历项（带用户ID过滤）
func (r *repository) GetCalendarItemByUID(userID *uint, uid string) (*CalendarItem, error) {
	var item CalendarItem
	query := r.db.Preload("Alarms").Where("uid = ?", uid)

	// 过滤用户ID
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if err := query.First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateCalendarItem 更新日历项（带用户ID过滤）
func (r *repository) UpdateCalendarItem(userID *uint, item *CalendarItem) error {
	query := r.db.Model(&CalendarItem{}).Where("id = ?", item.ID)

	// 过滤用户ID
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 先检查是否存在（通过更新影响行数）
	result := query.Updates(map[string]interface{}{
		"summary":          item.Summary,
		"description":      item.Description,
		"location":         item.Location,
		"organizer":        item.Organizer,
		"dt_start":         item.DtStart,
		"dt_end":           item.DtEnd,
		"due":              item.Due,
		"completed":        item.Completed,
		"duration":         item.Duration,
		"status":           item.Status,
		"priority":         item.Priority,
		"percent_complete": item.PercentComplete,
		"r_rule":           item.RRule,
		"ex_date":          item.ExDate,
		"r_date":           item.RDate,
		"categories":       item.Categories,
		"comment":          item.Comment,
		"contact":          item.Contact,
		"related_to":       item.RelatedTo,
		"resources":        item.Resources,
		"url":              item.URL,
		"class":            item.Class,
		"raw_ical":         item.RawIcal,
		"sequence":         item.Sequence,
		"last_modified":    item.LastModified,
	})

	if result.Error != nil {
		return result.Error
	}

	// 如果没有更新任何行，说明记录不存在或不属于该用户
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// DeleteCalendarItem 删除日历项（软删除，带用户ID过滤）
func (r *repository) DeleteCalendarItem(userID *uint, id uint) error {
	query := r.db.Model(&CalendarItem{}).Where("id = ?", id)

	// 过滤用户ID
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	result := query.Delete(&CalendarItem{})
	if result.Error != nil {
		return result.Error
	}

	// 如果没有删除任何行，说明记录不存在或不属于该用户
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
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

// applyTimeRangeFilters 应用时间范围过滤
func applyTimeRangeFilters(query *gorm.DB, timeRanges map[string]TimeRange) *gorm.DB {
	timeFieldColumnMap := map[string]string{
		"dtstart":   "dt_start",
		"dtend":     "dt_end",
		"due":       "due",
		"completed": "completed",
	}

	for field, timeRange := range timeRanges {
		column, ok := timeFieldColumnMap[field]
		if !ok {
			continue
		}

		isDtStart := field == "dtstart"
		if timeRange.Start != nil {
			if isDtStart {
				query = query.Where(column+" >= ?", *timeRange.Start)
			} else {
				query = query.Where("("+column+" >= ? OR "+column+" IS NULL)", *timeRange.Start)
			}
		}
		if timeRange.End != nil {
			if isDtStart {
				query = query.Where(column+" <= ?", *timeRange.End)
			} else {
				query = query.Where("("+column+" <= ? OR "+column+" IS NULL)", *timeRange.End)
			}
		}
	}

	return query
}

// applyFullTextSearch 应用全文搜索（在所有可搜索字段中搜索）
// 在所有可搜索字段中使用 OR 连接，任一字段匹配即可
func applyFullTextSearch(query *gorm.DB, q string) (*gorm.DB, error) {
	if q == "" {
		return query, nil
	}

	// 获取所有可搜索字段
	var searchQueries []string
	var args []interface{}

	keywordEscaped := escapeSQLString(q)
	keywordPattern := "%" + keywordEscaped + "%"

	// 在所有可搜索字段中搜索
	for field, column := range fieldColumnMap {
		_ = field // 字段名用于文档，实际使用列名
		searchQueries = append(searchQueries, column+" ILIKE ?")
		args = append(args, keywordPattern)
	}

	if len(searchQueries) > 0 {
		// 使用 OR 连接所有字段，任一字段匹配即可
		query = query.Where("("+strings.Join(searchQueries, " OR ")+")", args...)
	}

	return query, nil
}

// buildOrderBy 构建排序表达式
// 按开始时间（dtstart）升序排序
func buildOrderBy(q string) string {
	return "dt_start ASC"
}

// SearchCalendarItems 搜索日历项（按匹配程度排序）
// q: 搜索关键字，在所有可搜索字段中搜索
// 支持多个时间字段的范围过滤
// 注意：此方法假设参数已经由Service层验证，不再进行重复验证
func (r *repository) SearchCalendarItems(userID *uint, q string, timeRanges map[string]TimeRange, limit int) ([]*CalendarItem, error) {
	// 构建基础查询
	query := r.db.Model(&CalendarItem{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 应用时间范围过滤
	query = applyTimeRangeFilters(query, timeRanges)

	// 应用全文搜索（如果有关键字）
	if q != "" {
		var err error
		query, err = applyFullTextSearch(query, q)
		if err != nil {
			return nil, err
		}
	}

	// 应用排序
	query = query.Order(buildOrderBy(q))

	// 执行查询
	var items []*CalendarItem
	if err := query.Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}

	slog.Debug("SearchCalendarItems", "items_count", len(items))
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
