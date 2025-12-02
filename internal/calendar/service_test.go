package calendar

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepository 是 Repository 接口的 mock 实现
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) CreateCalendarItem(item *CalendarItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *mockRepository) GetCalendarItemByID(id uint) (*CalendarItem, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CalendarItem), args.Error(1)
}

func (m *mockRepository) GetCalendarItemByUID(uid string) (*CalendarItem, error) {
	args := m.Called(uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CalendarItem), args.Error(1)
}

func (m *mockRepository) UpdateCalendarItem(item *CalendarItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *mockRepository) DeleteCalendarItem(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockRepository) ListCalendarItems(userID *uint, startTime, endTime *time.Time, itemType *CalendarItemType, offset, limit int) ([]*CalendarItem, int64, error) {
	args := m.Called(userID, startTime, endTime, itemType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*CalendarItem), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepository) CreateValarm(alarm *Valarm) error {
	args := m.Called(alarm)
	return args.Error(0)
}

func (m *mockRepository) GetValarmByID(id uint) (*Valarm, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Valarm), args.Error(1)
}

func (m *mockRepository) GetValarmsByCalendarItemID(calendarItemID uint) ([]*Valarm, error) {
	args := m.Called(calendarItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Valarm), args.Error(1)
}

func (m *mockRepository) UpdateValarm(alarm *Valarm) error {
	args := m.Called(alarm)
	return args.Error(0)
}

func (m *mockRepository) DeleteValarm(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockRepository) DeleteValarmsByCalendarItemID(calendarItemID uint) error {
	args := m.Called(calendarItemID)
	return args.Error(0)
}

func (m *mockRepository) SearchCalendarItemsByKeyword(userID *uint, fields []string, keyword string) ([]*CalendarItem, error) {
	args := m.Called(userID, fields, keyword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*CalendarItem), args.Error(1)
}

func (m *mockRepository) SearchCalendarItemsByFieldKeywords(userID *uint, fieldKeywords map[string]string, timeRanges map[string]TimeRange) ([]*CalendarItem, error) {
	args := m.Called(userID, fieldKeywords, timeRanges)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*CalendarItem), args.Error(1)
}

// TestService_CreateCalendarItem_Success 测试创建日历项成功
func TestService_CreateCalendarItem_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	now := time.Now()
	summary := "测试事件"
	req := &CreateCalendarItemRequest{
		Type:    CalendarItemTypeEvent,
		Summary: &summary,
		DtStart: now,
	}

	// 设置 mock 期望
	mockRepo.On("CreateCalendarItem", mock.AnythingOfType("*calendar.CalendarItem")).
		Return(nil).
		Run(func(args mock.Arguments) {
			item := args.Get(0).(*CalendarItem)
			assert.Equal(t, CalendarItemTypeEvent, item.Type)
			assert.Equal(t, summary, *item.Summary)
			assert.Equal(t, now, item.DtStart)
			assert.Equal(t, userID, *item.UserID)
			assert.NotEmpty(t, item.UID)
		})

	item, err := service.CreateCalendarItem(&userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, CalendarItemTypeEvent, item.Type)
	assert.Equal(t, summary, *item.Summary)
	mockRepo.AssertExpectations(t)
}

// TestService_CreateCalendarItem_InvalidType 测试无效的日历项类型
func TestService_CreateCalendarItem_InvalidType(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	now := time.Now()
	req := &CreateCalendarItemRequest{
		Type:    CalendarItemType("INVALID"),
		DtStart: now,
	}

	item, err := service.CreateCalendarItem(&userID, req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidType, err)
	assert.Nil(t, item)
	mockRepo.AssertExpectations(t)
}

// TestService_GetCalendarItemByID_Success 测试根据ID获取日历项成功
func TestService_GetCalendarItemByID_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	itemID := uint(1)
	now := time.Now()
	summary := "测试事件"
	expectedItem := &CalendarItem{
		ID:      itemID,
		UID:     "test-uid-123",
		Type:    CalendarItemTypeEvent,
		Summary: &summary,
		DtStart: now,
	}

	mockRepo.On("GetCalendarItemByID", itemID).Return(expectedItem, nil)

	item, err := service.GetCalendarItemByID(itemID)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, itemID, item.ID)
	assert.Equal(t, "test-uid-123", item.UID)
	mockRepo.AssertExpectations(t)
}

// TestService_GetCalendarItemByID_NotFound 测试日历项不存在
func TestService_GetCalendarItemByID_NotFound(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	itemID := uint(999)

	mockRepo.On("GetCalendarItemByID", itemID).Return(nil, errors.New("not found"))

	item, err := service.GetCalendarItemByID(itemID)

	assert.Error(t, err)
	assert.Equal(t, ErrCalendarItemNotFound, err)
	assert.Nil(t, item)
	mockRepo.AssertExpectations(t)
}

// TestService_GetCalendarItemByUID_Success 测试根据UID获取日历项成功
func TestService_GetCalendarItemByUID_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	uid := "test-uid-123"
	now := time.Now()
	summary := "测试事件"
	expectedItem := &CalendarItem{
		ID:      1,
		UID:     uid,
		Type:    CalendarItemTypeEvent,
		Summary: &summary,
		DtStart: now,
	}

	mockRepo.On("GetCalendarItemByUID", uid).Return(expectedItem, nil)

	item, err := service.GetCalendarItemByUID(uid)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, uid, item.UID)
	mockRepo.AssertExpectations(t)
}

// TestService_UpdateCalendarItem_Success 测试更新日历项成功
func TestService_UpdateCalendarItem_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	itemID := uint(1)
	now := time.Now()
	summary := "原始标题"
	updatedSummary := "更新后的标题"
	existingItem := &CalendarItem{
		ID:      itemID,
		UID:     "test-uid-123",
		Type:    CalendarItemTypeEvent,
		Summary: &summary,
		DtStart: now,
	}

	req := &UpdateCalendarItemRequest{
		Summary: &updatedSummary,
	}

	// 第一次调用：获取现有项
	mockRepo.On("GetCalendarItemByID", itemID).Return(existingItem, nil)
	// 第二次调用：更新项
	mockRepo.On("UpdateCalendarItem", mock.AnythingOfType("*calendar.CalendarItem")).
		Return(nil).
		Run(func(args mock.Arguments) {
			item := args.Get(0).(*CalendarItem)
			assert.Equal(t, updatedSummary, *item.Summary)
			assert.NotNil(t, item.LastModified)
		})

	item, err := service.UpdateCalendarItem(itemID, req)

	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, updatedSummary, *item.Summary)
	mockRepo.AssertExpectations(t)
}

// TestService_UpdateCalendarItem_NotFound 测试更新不存在的日历项
func TestService_UpdateCalendarItem_NotFound(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	itemID := uint(999)
	req := &UpdateCalendarItemRequest{}

	mockRepo.On("GetCalendarItemByID", itemID).Return(nil, errors.New("not found"))

	item, err := service.UpdateCalendarItem(itemID, req)

	assert.Error(t, err)
	assert.Equal(t, ErrCalendarItemNotFound, err)
	assert.Nil(t, item)
	mockRepo.AssertExpectations(t)
}

// TestService_DeleteCalendarItem_Success 测试删除日历项成功
func TestService_DeleteCalendarItem_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	itemID := uint(1)
	existingItem := &CalendarItem{
		ID:   itemID,
		UID:  "test-uid-123",
		Type: CalendarItemTypeEvent,
	}

	mockRepo.On("GetCalendarItemByID", itemID).Return(existingItem, nil)
	mockRepo.On("DeleteCalendarItem", itemID).Return(nil)

	err := service.DeleteCalendarItem(itemID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestService_DeleteCalendarItem_NotFound 测试删除不存在的日历项
func TestService_DeleteCalendarItem_NotFound(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	itemID := uint(999)

	mockRepo.On("GetCalendarItemByID", itemID).Return(nil, errors.New("not found"))

	err := service.DeleteCalendarItem(itemID)

	assert.Error(t, err)
	assert.Equal(t, ErrCalendarItemNotFound, err)
	mockRepo.AssertExpectations(t)
}

// TestService_ListCalendarItems_Success 测试列出日历项成功
func TestService_ListCalendarItems_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	now := time.Now()
	startTime := now
	endTime := now.Add(24 * time.Hour)
	summary := "测试事件"

	items := []*CalendarItem{
		{
			ID:      1,
			UID:     "uid-1",
			Type:    CalendarItemTypeEvent,
			Summary: &summary,
			DtStart: now,
			UserID:  &userID,
		},
		{
			ID:      2,
			UID:     "uid-2",
			Type:    CalendarItemTypeEvent,
			Summary: &summary,
			DtStart: now.Add(time.Hour),
			UserID:  &userID,
		},
	}

	req := &ListCalendarItemsRequest{
		Page:      1,
		PageSize:  10,
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	mockRepo.On("ListCalendarItems", &userID, &startTime, &endTime, (*CalendarItemType)(nil), 0, 10).
		Return(items, int64(2), nil)

	result, err := service.ListCalendarItems(&userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PageSize)
	mockRepo.AssertExpectations(t)
}

// TestService_ListCalendarItems_DefaultPagination 测试默认分页参数
func TestService_ListCalendarItems_DefaultPagination(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	req := &ListCalendarItemsRequest{
		Page:     0, // 无效值，应该使用默认值1
		PageSize: 0, // 无效值，应该使用默认值10
	}

	mockRepo.On("ListCalendarItems", &userID, (*time.Time)(nil), (*time.Time)(nil), (*CalendarItemType)(nil), 0, 10).
		Return([]*CalendarItem{}, int64(0), nil)

	result, err := service.ListCalendarItems(&userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PageSize)
	mockRepo.AssertExpectations(t)
}

// TestService_CreateValarm_Success 测试创建提醒成功
func TestService_CreateValarm_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	calendarItemID := uint(1)
	description := "提醒内容"
	req := &CreateValarmRequest{
		Action:      ValarmActionDisplay,
		Trigger:     "-PT15M",
		Description: &description,
	}

	existingItem := &CalendarItem{
		ID:   calendarItemID,
		UID:  "test-uid-123",
		Type: CalendarItemTypeEvent,
	}

	mockRepo.On("GetCalendarItemByID", calendarItemID).Return(existingItem, nil)
	mockRepo.On("CreateValarm", mock.AnythingOfType("*calendar.Valarm")).
		Return(nil).
		Run(func(args mock.Arguments) {
			alarm := args.Get(0).(*Valarm)
			assert.Equal(t, calendarItemID, alarm.CalendarItemID)
			assert.Equal(t, ValarmActionDisplay, alarm.Action)
			assert.Equal(t, "-PT15M", alarm.Trigger)
			assert.Equal(t, description, *alarm.Description)
		})

	alarm, err := service.CreateValarm(calendarItemID, req)

	assert.NoError(t, err)
	assert.NotNil(t, alarm)
	assert.Equal(t, ValarmActionDisplay, alarm.Action)
	mockRepo.AssertExpectations(t)
}

// TestService_CreateValarm_CalendarItemNotFound 测试日历项不存在
func TestService_CreateValarm_CalendarItemNotFound(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	calendarItemID := uint(999)
	description := "提醒内容"
	req := &CreateValarmRequest{
		Action:      ValarmActionDisplay,
		Trigger:     "-PT15M",
		Description: &description,
	}

	mockRepo.On("GetCalendarItemByID", calendarItemID).Return(nil, errors.New("not found"))

	alarm, err := service.CreateValarm(calendarItemID, req)

	assert.Error(t, err)
	assert.Equal(t, ErrCalendarItemNotFound, err)
	assert.Nil(t, alarm)
	mockRepo.AssertExpectations(t)
}

// TestService_CreateValarm_DisplayWithoutDescription 测试DISPLAY类型缺少description
func TestService_CreateValarm_DisplayWithoutDescription(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	calendarItemID := uint(1)
	existingItem := &CalendarItem{
		ID:   calendarItemID,
		UID:  "test-uid-123",
		Type: CalendarItemTypeEvent,
	}

	req := &CreateValarmRequest{
		Action:      ValarmActionDisplay,
		Trigger:     "-PT15M",
		Description: nil, // DISPLAY类型必须提供description
	}

	mockRepo.On("GetCalendarItemByID", calendarItemID).Return(existingItem, nil)

	alarm, err := service.CreateValarm(calendarItemID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrInvalidInput.Error())
	assert.Nil(t, alarm)
	mockRepo.AssertExpectations(t)
}

// TestService_CreateValarm_InvalidAction 测试无效的提醒动作类型
func TestService_CreateValarm_InvalidAction(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	calendarItemID := uint(1)
	existingItem := &CalendarItem{
		ID:   calendarItemID,
		UID:  "test-uid-123",
		Type: CalendarItemTypeEvent,
	}

	req := &CreateValarmRequest{
		Action:  ValarmAction("INVALID"),
		Trigger: "-PT15M",
	}

	mockRepo.On("GetCalendarItemByID", calendarItemID).Return(existingItem, nil)

	alarm, err := service.CreateValarm(calendarItemID, req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAction, err)
	assert.Nil(t, alarm)
	mockRepo.AssertExpectations(t)
}

// TestService_GetValarmByID_Success 测试根据ID获取提醒成功
func TestService_GetValarmByID_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	alarmID := uint(1)
	description := "提醒内容"
	expectedAlarm := &Valarm{
		ID:             alarmID,
		CalendarItemID: 1,
		Action:         ValarmActionDisplay,
		Trigger:        "-PT15M",
		Description:    &description,
	}

	mockRepo.On("GetValarmByID", alarmID).Return(expectedAlarm, nil)

	alarm, err := service.GetValarmByID(alarmID)

	assert.NoError(t, err)
	assert.NotNil(t, alarm)
	assert.Equal(t, alarmID, alarm.ID)
	mockRepo.AssertExpectations(t)
}

// TestService_GetValarmByID_NotFound 测试提醒不存在
func TestService_GetValarmByID_NotFound(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	alarmID := uint(999)

	mockRepo.On("GetValarmByID", alarmID).Return(nil, errors.New("not found"))

	alarm, err := service.GetValarmByID(alarmID)

	assert.Error(t, err)
	assert.Equal(t, ErrValarmNotFound, err)
	assert.Nil(t, alarm)
	mockRepo.AssertExpectations(t)
}

// TestService_GetValarmsByCalendarItemID_Success 测试获取日历项的所有提醒成功
func TestService_GetValarmsByCalendarItemID_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	calendarItemID := uint(1)
	description := "提醒内容"
	alarms := []*Valarm{
		{
			ID:             1,
			CalendarItemID: calendarItemID,
			Action:         ValarmActionDisplay,
			Trigger:        "-PT15M",
			Description:    &description,
		},
		{
			ID:             2,
			CalendarItemID: calendarItemID,
			Action:         ValarmActionEmail,
			Trigger:        "-PT1H",
		},
	}

	existingItem := &CalendarItem{
		ID:   calendarItemID,
		UID:  "test-uid-123",
		Type: CalendarItemTypeEvent,
	}

	mockRepo.On("GetCalendarItemByID", calendarItemID).Return(existingItem, nil)
	mockRepo.On("GetValarmsByCalendarItemID", calendarItemID).Return(alarms, nil)

	result, err := service.GetValarmsByCalendarItemID(calendarItemID)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, ValarmActionDisplay, result[0].Action)
	assert.Equal(t, ValarmActionEmail, result[1].Action)
	mockRepo.AssertExpectations(t)
}

// TestService_UpdateValarm_Success 测试更新提醒成功
func TestService_UpdateValarm_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	alarmID := uint(1)
	originalDescription := "原始提醒"
	updatedDescription := "更新后的提醒"
	existingAlarm := &Valarm{
		ID:             alarmID,
		CalendarItemID: 1,
		Action:         ValarmActionDisplay,
		Trigger:        "-PT15M",
		Description:    &originalDescription,
	}

	updatedTrigger := "-PT30M"
	req := &UpdateValarmRequest{
		Trigger:     &updatedTrigger,
		Description: &updatedDescription,
	}

	mockRepo.On("GetValarmByID", alarmID).Return(existingAlarm, nil)
	mockRepo.On("UpdateValarm", mock.AnythingOfType("*calendar.Valarm")).
		Return(nil).
		Run(func(args mock.Arguments) {
			alarm := args.Get(0).(*Valarm)
			assert.Equal(t, updatedTrigger, alarm.Trigger)
			assert.Equal(t, updatedDescription, *alarm.Description)
		})

	alarm, err := service.UpdateValarm(alarmID, req)

	assert.NoError(t, err)
	assert.NotNil(t, alarm)
	assert.Equal(t, updatedTrigger, alarm.Trigger)
	assert.Equal(t, updatedDescription, *alarm.Description)
	mockRepo.AssertExpectations(t)
}

// TestService_UpdateValarm_NotFound 测试更新不存在的提醒
func TestService_UpdateValarm_NotFound(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	alarmID := uint(999)
	req := &UpdateValarmRequest{}

	mockRepo.On("GetValarmByID", alarmID).Return(nil, errors.New("not found"))

	alarm, err := service.UpdateValarm(alarmID, req)

	assert.Error(t, err)
	assert.Equal(t, ErrValarmNotFound, err)
	assert.Nil(t, alarm)
	mockRepo.AssertExpectations(t)
}

// TestService_DeleteValarm_Success 测试删除提醒成功
func TestService_DeleteValarm_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	alarmID := uint(1)
	existingAlarm := &Valarm{
		ID:             alarmID,
		CalendarItemID: 1,
		Action:         ValarmActionDisplay,
	}

	mockRepo.On("GetValarmByID", alarmID).Return(existingAlarm, nil)
	mockRepo.On("DeleteValarm", alarmID).Return(nil)

	err := service.DeleteValarm(alarmID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestService_DeleteValarm_NotFound 测试删除不存在的提醒
func TestService_DeleteValarm_NotFound(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	alarmID := uint(999)

	mockRepo.On("GetValarmByID", alarmID).Return(nil, errors.New("not found"))

	err := service.DeleteValarm(alarmID)

	assert.Error(t, err)
	assert.Equal(t, ErrValarmNotFound, err)
	mockRepo.AssertExpectations(t)
}

// TestService_SearchCalendarItems_Success 测试搜索日历项成功
func TestService_SearchCalendarItems_Success(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	summary := "测试事件"
	keyword := "测试"
	req := &SearchCalendarItemsRequest{
		FieldKeywords: map[string]string{
			"summary": keyword,
		},
	}

	expectedItems := []*CalendarItem{
		{
			ID:      1,
			UID:     "test-uid-1",
			Type:    CalendarItemTypeEvent,
			Summary: &summary,
		},
	}

	mockRepo.On("SearchCalendarItemsByFieldKeywords", &userID, map[string]string{"summary": keyword}, map[string]TimeRange(nil)).Return(expectedItems, nil)

	items, err := service.SearchCalendarItems(&userID, req)

	assert.NoError(t, err)
	assert.Equal(t, expectedItems, items)
	mockRepo.AssertExpectations(t)
}

// TestService_SearchCalendarItems_MultipleFields 测试多字段搜索
func TestService_SearchCalendarItems_MultipleFields(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	summary := "测试事件"
	location := "北京"
	keyword := "测试"
	req := &SearchCalendarItemsRequest{
		FieldKeywords: map[string]string{
			"summary":  keyword,
			"location": keyword,
		},
	}

	expectedItems := []*CalendarItem{
		{
			ID:       1,
			UID:      "test-uid-1",
			Type:     CalendarItemTypeEvent,
			Summary:  &summary,
			Location: &location,
		},
	}

	mockRepo.On("SearchCalendarItemsByFieldKeywords", &userID, map[string]string{"summary": keyword, "location": keyword}, map[string]TimeRange(nil)).Return(expectedItems, nil)

	items, err := service.SearchCalendarItems(&userID, req)

	assert.NoError(t, err)
	assert.Equal(t, expectedItems, items)
	mockRepo.AssertExpectations(t)
}

// TestService_SearchCalendarItems_DuplicateFields 测试重复字段去重
func TestService_SearchCalendarItems_DuplicateFields(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	summary := "测试事件"
	keyword := "测试"
	req := &SearchCalendarItemsRequest{
		FieldKeywords: map[string]string{
			"summary":  keyword,
			"location": keyword,
		},
	}

	expectedItems := []*CalendarItem{
		{
			ID:      1,
			UID:     "test-uid-1",
			Type:    CalendarItemTypeEvent,
			Summary: &summary,
		},
	}

	// map 自动去重，所以直接使用
	mockRepo.On("SearchCalendarItemsByFieldKeywords", &userID, map[string]string{"summary": keyword, "location": keyword}, map[string]TimeRange(nil)).Return(expectedItems, nil)

	items, err := service.SearchCalendarItems(&userID, req)

	assert.NoError(t, err)
	assert.Equal(t, expectedItems, items)
	mockRepo.AssertExpectations(t)
}

// TestService_SearchCalendarItems_InvalidField 测试无效字段
func TestService_SearchCalendarItems_InvalidField(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	req := &SearchCalendarItemsRequest{
		FieldKeywords: map[string]string{
			"invalid_field": "测试",
		},
	}

	_, err := service.SearchCalendarItems(&userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrInvalidSearchField.Error())
	mockRepo.AssertExpectations(t)
}

// TestService_SearchCalendarItems_EmptyFields 测试空字段列表
func TestService_SearchCalendarItems_EmptyFields(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	req := &SearchCalendarItemsRequest{
		FieldKeywords: map[string]string{},
	}

	_, err := service.SearchCalendarItems(&userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrInvalidInput.Error())
	mockRepo.AssertExpectations(t)
}

// TestService_SearchCalendarItems_EmptyKeyword 测试空关键字
func TestService_SearchCalendarItems_EmptyKeyword(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	req := &SearchCalendarItemsRequest{
		FieldKeywords: map[string]string{
			"summary": "",
		},
	}

	_, err := service.SearchCalendarItems(&userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), ErrInvalidInput.Error())
	mockRepo.AssertExpectations(t)
}

// TestService_SearchCalendarItems_RepositoryError 测试 Repository 返回错误
func TestService_SearchCalendarItems_RepositoryError(t *testing.T) {
	mockRepo := new(mockRepository)
	service := NewService(mockRepo)

	userID := uint(1)
	req := &SearchCalendarItemsRequest{
		FieldKeywords: map[string]string{
			"summary": "测试",
		},
	}

	repoError := errors.New("数据库错误")
	mockRepo.On("SearchCalendarItemsByFieldKeywords", &userID, map[string]string{"summary": "测试"}, map[string]TimeRange(nil)).Return(nil, repoError)

	_, err := service.SearchCalendarItems(&userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "搜索日历项失败")
	mockRepo.AssertExpectations(t)
}
