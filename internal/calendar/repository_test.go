package calendar

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB 创建测试用的数据库连接（使用sqlmock）
func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建sqlmock失败: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建GORM连接失败: %v", err)
	}

	return gormDB, mock
}

// TestRepository_CreateCalendarItem 测试创建日历项
func TestRepository_CreateCalendarItem(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	now := time.Now()
	summary := "测试事件"
	item := &CalendarItem{
		UID:     "test-uid-123",
		Type:    CalendarItemTypeEvent,
		Summary: &summary,
		DtStart: now,
	}

	// 设置期望：INSERT 语句
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "calendar_items"`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			sqlmock.AnyArg(), // DeletedAt
			item.UID,
			item.Type,
			item.Summary,
			sqlmock.AnyArg(), // Description
			sqlmock.AnyArg(), // Location
			sqlmock.AnyArg(), // Organizer
			item.DtStart,
			sqlmock.AnyArg(), // DtEnd
			sqlmock.AnyArg(), // Due
			sqlmock.AnyArg(), // Completed
			sqlmock.AnyArg(), // Duration
			sqlmock.AnyArg(), // Status
			sqlmock.AnyArg(), // Priority
			sqlmock.AnyArg(), // PercentComplete
			sqlmock.AnyArg(), // Sequence
			sqlmock.AnyArg(), // RRule
			sqlmock.AnyArg(), // ExDate
			sqlmock.AnyArg(), // RDate
			sqlmock.AnyArg(), // Categories
			sqlmock.AnyArg(), // Comment
			sqlmock.AnyArg(), // Contact
			sqlmock.AnyArg(), // RelatedTo
			sqlmock.AnyArg(), // Resources
			sqlmock.AnyArg(), // URL
			sqlmock.AnyArg(), // Class
			sqlmock.AnyArg(), // LastModified
			sqlmock.AnyArg(), // RawIcal
			sqlmock.AnyArg(), // UserID
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.CreateCalendarItem(item)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetCalendarItemByID_Success 测试根据ID获取日历项成功
func TestRepository_GetCalendarItemByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	itemID := uint(1)
	now := time.Now()
	summary := "测试事件"
	exDateJSON, _ := json.Marshal([]string{})
	categoriesJSON, _ := json.Marshal([]string{"工作"})
	resourcesJSON, _ := json.Marshal([]string{})

	// 主查询：获取日历项
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "uid", "type", "summary",
		"description", "location", "organizer", "dtstart", "dtend", "due",
		"completed", "duration", "status", "priority", "percent_complete",
		"sequence", "rrule", "exdate", "rdate", "categories", "comment",
		"contact", "related_to", "resources", "url", "class", "last_modified",
		"raw_ical", "user_id",
	}).AddRow(
		itemID, now, now, nil, "test-uid-123", CalendarItemTypeEvent,
		summary, nil, nil, nil, now, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, exDateJSON, nil, categoriesJSON, nil, nil, nil,
		resourcesJSON, nil, nil, nil, nil, nil,
	)

	mock.ExpectQuery(`SELECT \* FROM "calendar_items"`).
		WithArgs(itemID, 1).
		WillReturnRows(rows)

	// Preload Alarms 查询（即使没有alarms也会查询）
	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	item, err := repo.GetCalendarItemByID(itemID)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, itemID, item.ID)
	assert.Equal(t, "test-uid-123", item.UID)
	assert.Equal(t, CalendarItemTypeEvent, item.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetCalendarItemByID_NotFound 测试日历项不存在
func TestRepository_GetCalendarItemByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	itemID := uint(999)

	mock.ExpectQuery(`SELECT \* FROM "calendar_items"`).
		WithArgs(itemID, 1).
		WillReturnError(sql.ErrNoRows)

	item, err := repo.GetCalendarItemByID(itemID)

	assert.Error(t, err)
	assert.Nil(t, item)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetCalendarItemByUID_Success 测试根据UID获取日历项成功
func TestRepository_GetCalendarItemByUID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	uid := "test-uid-123"
	now := time.Now()
	summary := "测试事件"
	exDateJSON, _ := json.Marshal([]string{})
	categoriesJSON, _ := json.Marshal([]string{"工作"})
	resourcesJSON, _ := json.Marshal([]string{})

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "uid", "type", "summary",
		"description", "location", "organizer", "dtstart", "dtend", "due",
		"completed", "duration", "status", "priority", "percent_complete",
		"sequence", "rrule", "exdate", "rdate", "categories", "comment",
		"contact", "related_to", "resources", "url", "class", "last_modified",
		"raw_ical", "user_id",
	}).AddRow(
		1, now, now, nil, uid, CalendarItemTypeEvent,
		summary, nil, nil, nil, now, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, exDateJSON, nil, categoriesJSON, nil, nil, nil,
		resourcesJSON, nil, nil, nil, nil, nil,
	)

	mock.ExpectQuery(`SELECT \* FROM "calendar_items"`).
		WithArgs(uid, 1).
		WillReturnRows(rows)

	// Preload Alarms 查询
	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	item, err := repo.GetCalendarItemByUID(uid)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, uid, item.UID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetCalendarItemByUID_NotFound 测试UID不存在
func TestRepository_GetCalendarItemByUID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	uid := "non-existent-uid"

	mock.ExpectQuery(`SELECT \* FROM "calendar_items"`).
		WithArgs(uid, 1).
		WillReturnError(sql.ErrNoRows)

	item, err := repo.GetCalendarItemByUID(uid)

	assert.Error(t, err)
	assert.Nil(t, item)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_UpdateCalendarItem 测试更新日历项
func TestRepository_UpdateCalendarItem(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	now := time.Now()
	summary := "更新后的标题"
	item := &CalendarItem{
		ID:      1,
		UID:     "test-uid-123",
		Type:    CalendarItemTypeEvent,
		Summary: &summary,
		DtStart: now,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "calendar_items" SET`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			sqlmock.AnyArg(), // DeletedAt
			item.UID,
			item.Type,
			item.Summary,
			sqlmock.AnyArg(), // Description
			sqlmock.AnyArg(), // Location
			sqlmock.AnyArg(), // Organizer
			item.DtStart,
			sqlmock.AnyArg(), // DtEnd
			sqlmock.AnyArg(), // Due
			sqlmock.AnyArg(), // Completed
			sqlmock.AnyArg(), // Duration
			sqlmock.AnyArg(), // Status
			sqlmock.AnyArg(), // Priority
			sqlmock.AnyArg(), // PercentComplete
			sqlmock.AnyArg(), // Sequence
			sqlmock.AnyArg(), // RRule
			sqlmock.AnyArg(), // ExDate
			sqlmock.AnyArg(), // RDate
			sqlmock.AnyArg(), // Categories
			sqlmock.AnyArg(), // Comment
			sqlmock.AnyArg(), // Contact
			sqlmock.AnyArg(), // RelatedTo
			sqlmock.AnyArg(), // Resources
			sqlmock.AnyArg(), // URL
			sqlmock.AnyArg(), // Class
			sqlmock.AnyArg(), // LastModified
			sqlmock.AnyArg(), // RawIcal
			sqlmock.AnyArg(), // UserID
			item.ID,          // WHERE条件中的ID
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateCalendarItem(item)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_DeleteCalendarItem 测试删除日历项
func TestRepository_DeleteCalendarItem(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	itemID := uint(1)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "calendar_items" SET`).
		WithArgs(sqlmock.AnyArg(), itemID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeleteCalendarItem(itemID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_ListCalendarItems_Success 测试列出日历项成功
func TestRepository_ListCalendarItems_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(1)
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour)
	offset := 0
	limit := 10
	total := int64(2)

	// COUNT 查询
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(total)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "calendar_items"`).
		WillReturnRows(countRows)

	// SELECT 查询
	exDateJSON, _ := json.Marshal([]string{})
	categoriesJSON, _ := json.Marshal([]string{"工作"})
	resourcesJSON, _ := json.Marshal([]string{})
	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "uid", "type", "summary",
		"description", "location", "organizer", "dtstart", "dtend", "due",
		"completed", "duration", "status", "priority", "percent_complete",
		"sequence", "rrule", "exdate", "rdate", "categories", "comment",
		"contact", "related_to", "resources", "url", "class", "last_modified",
		"raw_ical", "user_id",
	})
	for i := 1; i <= 2; i++ {
		summary := "事件" + string(rune(i+'0'))
		rows.AddRow(
			uint(i), startTime, startTime, nil, "uid-"+string(rune(i+'0')),
			CalendarItemTypeEvent, summary, nil, nil, nil, startTime, nil, nil,
			nil, nil, nil, nil, nil, nil, nil, exDateJSON, nil, categoriesJSON,
			nil, nil, nil, resourcesJSON, nil, nil, nil, nil, userID,
		)
	}

	mock.ExpectQuery(`SELECT \* FROM "calendar_items"`).
		WithArgs(userID, endTime, startTime, limit).
		WillReturnRows(rows)

	// Preload Alarms 查询
	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	items, totalCount, err := repo.ListCalendarItems(&userID, &startTime, &endTime, nil, offset, limit)

	assert.NoError(t, err)
	assert.Equal(t, total, totalCount)
	assert.Len(t, items, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_CreateValarm 测试创建提醒
func TestRepository_CreateValarm(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	description := "提醒内容"
	xPropertyJSON, _ := json.Marshal(map[string]interface{}{"key": "value"})
	alarm := &Valarm{
		CalendarItemID: 1,
		Action:         ValarmActionDisplay,
		Trigger:        "-PT15M",
		Description:    &description,
		XProperty:      JSONB{"key": "value"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "valarms"`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			sqlmock.AnyArg(), // DeletedAt
			alarm.CalendarItemID,
			alarm.Action,
			alarm.Trigger,
			alarm.Description,
			sqlmock.AnyArg(), // Summary
			sqlmock.AnyArg(), // Attendee
			sqlmock.AnyArg(), // Duration
			sqlmock.AnyArg(), // RepeatCount
			xPropertyJSON,    // XProperty
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.CreateValarm(alarm)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetValarmByID_Success 测试根据ID获取提醒成功
func TestRepository_GetValarmByID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	alarmID := uint(1)
	description := "提醒内容"
	xPropertyJSON, _ := json.Marshal(map[string]interface{}{"key": "value"})

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "calendar_item_id",
		"action", "trigger", "description", "summary", "attendee",
		"duration", "repeat_count", "x_property",
	}).AddRow(
		alarmID, time.Now(), time.Now(), nil, 1,
		ValarmActionDisplay, "-PT15M", description, nil, nil,
		nil, nil, xPropertyJSON,
	)

	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WithArgs(alarmID, 1).
		WillReturnRows(rows)

	alarm, err := repo.GetValarmByID(alarmID)

	assert.NoError(t, err)
	assert.NotNil(t, alarm)
	assert.Equal(t, alarmID, alarm.ID)
	assert.Equal(t, ValarmActionDisplay, alarm.Action)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetValarmByID_NotFound 测试提醒不存在
func TestRepository_GetValarmByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	alarmID := uint(999)

	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WithArgs(alarmID, 1).
		WillReturnError(sql.ErrNoRows)

	alarm, err := repo.GetValarmByID(alarmID)

	assert.Error(t, err)
	assert.Nil(t, alarm)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_GetValarmsByCalendarItemID_Success 测试获取日历项的所有提醒成功
func TestRepository_GetValarmsByCalendarItemID_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	calendarItemID := uint(1)
	description := "提醒内容"
	xPropertyJSON, _ := json.Marshal(map[string]interface{}{})

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "calendar_item_id",
		"action", "trigger", "description", "summary", "attendee",
		"duration", "repeat_count", "x_property",
	}).AddRow(
		1, time.Now(), time.Now(), nil, calendarItemID,
		ValarmActionDisplay, "-PT15M", description, nil, nil,
		nil, nil, xPropertyJSON,
	).AddRow(
		2, time.Now(), time.Now(), nil, calendarItemID,
		ValarmActionEmail, "-PT1H", nil, nil, "test@example.com",
		nil, nil, xPropertyJSON,
	)

	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WithArgs(calendarItemID).
		WillReturnRows(rows)

	alarms, err := repo.GetValarmsByCalendarItemID(calendarItemID)

	assert.NoError(t, err)
	assert.Len(t, alarms, 2)
	assert.Equal(t, ValarmActionDisplay, alarms[0].Action)
	assert.Equal(t, ValarmActionEmail, alarms[1].Action)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_UpdateValarm 测试更新提醒
func TestRepository_UpdateValarm(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	description := "更新后的提醒内容"
	xPropertyJSON, _ := json.Marshal(map[string]interface{}{"key": "updated"})
	alarm := &Valarm{
		ID:             1,
		CalendarItemID: 1,
		Action:         ValarmActionDisplay,
		Trigger:        "-PT30M",
		Description:    &description,
		XProperty:      JSONB{"key": "updated"},
	}

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "valarms" SET`).
		WithArgs(
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			sqlmock.AnyArg(), // DeletedAt
			alarm.CalendarItemID,
			alarm.Action,
			alarm.Trigger,
			alarm.Description,
			sqlmock.AnyArg(), // Summary
			sqlmock.AnyArg(), // Attendee
			sqlmock.AnyArg(), // Duration
			sqlmock.AnyArg(), // RepeatCount
			xPropertyJSON,    // XProperty
			alarm.ID,         // WHERE条件中的ID
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateValarm(alarm)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_DeleteValarm 测试删除提醒
func TestRepository_DeleteValarm(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	alarmID := uint(1)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "valarms" SET`).
		WithArgs(sqlmock.AnyArg(), alarmID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeleteValarm(alarmID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_DeleteValarmsByCalendarItemID 测试删除日历项的所有提醒
func TestRepository_DeleteValarmsByCalendarItemID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	calendarItemID := uint(1)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "valarms" SET`).
		WithArgs(sqlmock.AnyArg(), calendarItemID).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectCommit()

	err := repo.DeleteValarmsByCalendarItemID(calendarItemID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_SearchCalendarItemsByKeyword_SingleField 测试单字段搜索
func TestRepository_SearchCalendarItemsByKeyword_SingleField(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(1)
	fieldKeywords := map[string]string{
		"summary": "测试",
	}
	keywordPattern := "%测试%"

	summary := "测试事件"
	exDateJSON, _ := json.Marshal([]string{})
	categoriesJSON, _ := json.Marshal([]string{})
	resourcesJSON, _ := json.Marshal([]string{})

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "uid", "type", "summary",
		"description", "location", "organizer", "dtstart", "dtend", "due",
		"completed", "duration", "status", "priority", "percent_complete",
		"sequence", "rrule", "exdate", "rdate", "categories", "comment",
		"contact", "related_to", "resources", "url", "class", "last_modified",
		"raw_ical", "user_id",
	}).AddRow(
		1, time.Now(), time.Now(), nil, "uid-1", CalendarItemTypeEvent,
		summary, nil, nil, nil, time.Now(), nil, nil, nil, nil, nil,
		nil, nil, nil, nil, exDateJSON, nil, categoriesJSON, nil, nil,
		nil, resourcesJSON, nil, nil, nil, nil, userID,
	)

	// GORM 会自动添加 deleted_at IS NULL 和 LIMIT 条件
	mock.ExpectQuery(`SELECT \* FROM "calendar_items"`).
		WithArgs(userID, keywordPattern, sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Preload Alarms 查询
	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	items, err := repo.SearchCalendarItemsByFieldKeywords(&userID, fieldKeywords)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, summary, *items[0].Summary)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_SearchCalendarItemsByKeyword_MultipleFields 测试多字段搜索
func TestRepository_SearchCalendarItemsByKeyword_MultipleFields(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(1)
	fieldKeywords := map[string]string{
		"summary":  "测试",
		"location": "测试",
	}
	keywordPattern := "%测试%"

	summary := "测试事件"
	location := "测试地点"
	exDateJSON, _ := json.Marshal([]string{})
	categoriesJSON, _ := json.Marshal([]string{})
	resourcesJSON, _ := json.Marshal([]string{})

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "uid", "type", "summary",
		"description", "location", "organizer", "dtstart", "dtend", "due",
		"completed", "duration", "status", "priority", "percent_complete",
		"sequence", "rrule", "exdate", "rdate", "categories", "comment",
		"contact", "related_to", "resources", "url", "class", "last_modified",
		"raw_ical", "user_id",
	}).AddRow(
		1, time.Now(), time.Now(), nil, "uid-1", CalendarItemTypeEvent,
		summary, nil, location, nil, time.Now(), nil, nil, nil, nil, nil,
		nil, nil, nil, nil, exDateJSON, nil, categoriesJSON, nil, nil,
		nil, resourcesJSON, nil, nil, nil, nil, userID,
	)

	// 期望两个字段的 AND 条件，GORM 会自动添加 deleted_at IS NULL 和 LIMIT 条件
	mock.ExpectQuery(`SELECT \* FROM "calendar_items"`).
		WithArgs(userID, keywordPattern, keywordPattern, sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Preload Alarms 查询
	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	items, err := repo.SearchCalendarItemsByFieldKeywords(&userID, fieldKeywords)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, summary, *items[0].Summary)
	assert.Equal(t, location, *items[0].Location)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_SearchCalendarItemsByKeyword_EmptyKeyword 测试空关键字
func TestRepository_SearchCalendarItemsByKeyword_EmptyKeyword(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(1)
	fieldKeywords := map[string]string{
		"summary": "",
	}

	items, err := repo.SearchCalendarItemsByFieldKeywords(&userID, fieldKeywords)

	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "关键字不能为空")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_SearchCalendarItemsByKeyword_EmptyFields 测试空字段列表
func TestRepository_SearchCalendarItemsByKeyword_EmptyFields(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(1)
	fieldKeywords := map[string]string{}

	items, err := repo.SearchCalendarItemsByFieldKeywords(&userID, fieldKeywords)

	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "至少需要指定一个搜索字段")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_SearchCalendarItemsByKeyword_InvalidField 测试无效字段
func TestRepository_SearchCalendarItemsByKeyword_InvalidField(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(1)
	fieldKeywords := map[string]string{
		"invalid_field": "测试",
	}

	items, err := repo.SearchCalendarItemsByFieldKeywords(&userID, fieldKeywords)

	assert.Error(t, err)
	assert.Nil(t, items)
	assert.Contains(t, err.Error(), "不支持的搜索字段")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_SearchCalendarItemsByKeyword_DuplicateFields 测试重复字段去重
func TestRepository_SearchCalendarItemsByKeyword_DuplicateFields(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(1)
	fieldKeywords := map[string]string{
		"summary":  "测试",
		"location": "测试",
	}
	keywordPattern := "%测试%"

	summary := "测试事件"
	location := "测试地点"
	exDateJSON, _ := json.Marshal([]string{})
	categoriesJSON, _ := json.Marshal([]string{})
	resourcesJSON, _ := json.Marshal([]string{})

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "uid", "type", "summary",
		"description", "location", "organizer", "dtstart", "dtend", "due",
		"completed", "duration", "status", "priority", "percent_complete",
		"sequence", "rrule", "exdate", "rdate", "categories", "comment",
		"contact", "related_to", "resources", "url", "class", "last_modified",
		"raw_ical", "user_id",
	}).AddRow(
		1, time.Now(), time.Now(), nil, "uid-1", CalendarItemTypeEvent,
		summary, nil, location, nil, time.Now(), nil, nil, nil, nil, nil,
		nil, nil, nil, nil, exDateJSON, nil, categoriesJSON, nil, nil,
		nil, resourcesJSON, nil, nil, nil, nil, userID,
	)

	// map 自动去重，GORM 会自动添加 deleted_at IS NULL 和 LIMIT 条件
	mock.ExpectQuery(`SELECT \* FROM "calendar_items"`).
		WithArgs(userID, keywordPattern, keywordPattern, sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Preload Alarms 查询
	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	items, err := repo.SearchCalendarItemsByFieldKeywords(&userID, fieldKeywords)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_SearchCalendarItemsByKeyword_NoUserID 测试无用户ID搜索
func TestRepository_SearchCalendarItemsByKeyword_NoUserID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	fieldKeywords := map[string]string{
		"summary": "测试",
	}
	keywordPattern := "%测试%"

	summary := "测试事件"
	exDateJSON, _ := json.Marshal([]string{})
	categoriesJSON, _ := json.Marshal([]string{})
	resourcesJSON, _ := json.Marshal([]string{})

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "uid", "type", "summary",
		"description", "location", "organizer", "dtstart", "dtend", "due",
		"completed", "duration", "status", "priority", "percent_complete",
		"sequence", "rrule", "exdate", "rdate", "categories", "comment",
		"contact", "related_to", "resources", "url", "class", "last_modified",
		"raw_ical", "user_id",
	}).AddRow(
		1, time.Now(), time.Now(), nil, "uid-1", CalendarItemTypeEvent,
		summary, nil, nil, nil, time.Now(), nil, nil, nil, nil, nil,
		nil, nil, nil, nil, exDateJSON, nil, categoriesJSON, nil, nil,
		nil, resourcesJSON, nil, nil, nil, nil, nil,
	)

	// GORM 会自动添加 deleted_at IS NULL 和 LIMIT 条件
	mock.ExpectQuery(`SELECT \* FROM "calendar_items"`).
		WithArgs(keywordPattern, sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Preload Alarms 查询
	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	items, err := repo.SearchCalendarItemsByFieldKeywords(nil, fieldKeywords)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRepository_SearchCalendarItemsByKeyword_JSONBFields 测试 JSONB 字段搜索
func TestRepository_SearchCalendarItemsByKeyword_JSONBFields(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRepository(db)

	userID := uint(1)
	fieldKeywords := map[string]string{
		"categories": "工作",
	}
	keywordPattern := "%工作%"

	summary := "测试事件"
	categoriesJSON, _ := json.Marshal([]string{"工作", "重要"})
	exDateJSON, _ := json.Marshal([]string{})
	resourcesJSON, _ := json.Marshal([]string{})

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "deleted_at", "uid", "type", "summary",
		"description", "location", "organizer", "dtstart", "dtend", "due",
		"completed", "duration", "status", "priority", "percent_complete",
		"sequence", "rrule", "exdate", "rdate", "categories", "comment",
		"contact", "related_to", "resources", "url", "class", "last_modified",
		"raw_ical", "user_id",
	}).AddRow(
		1, time.Now(), time.Now(), nil, "uid-1", CalendarItemTypeEvent,
		summary, nil, nil, nil, time.Now(), nil, nil, nil, nil, nil,
		nil, nil, nil, nil, exDateJSON, nil, categoriesJSON, nil, nil,
		nil, resourcesJSON, nil, nil, nil, nil, userID,
	)

	// GORM 会自动添加 deleted_at IS NULL 和 LIMIT 条件
	mock.ExpectQuery(`SELECT \* FROM "calendar_items"`).
		WithArgs(userID, keywordPattern, sqlmock.AnyArg()).
		WillReturnRows(rows)

	// Preload Alarms 查询
	mock.ExpectQuery(`SELECT \* FROM "valarms"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	items, err := repo.SearchCalendarItemsByFieldKeywords(&userID, fieldKeywords)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
