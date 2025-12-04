package calendar

import (
	"log/slog"
	"time"

	"github.com/galilio/otter/internal/calendar"
	"github.com/galilio/otter/internal/common/utils"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

type calendarTools struct {
	service calendar.Service
}

func SetupTools(service calendar.Service) ([]tool.Tool, error) {
	ct := &calendarTools{service: service}
	tools := []tool.Tool{}

	createItemTool, err := functiontool.New(functiontool.Config{
		Name:         "create_calendar_item",
		Description:  "Create a calendar item. Supports VEVENT (requires dtstart, and either dtend or duration), VTODO (requires dtstart or due), VJOURNAL (requires dtstart), VFREEBUSY (requires both dtstart and dtend). If uid is provided and already exists, returns the existing item (idempotent).",
		InputSchema:  utils.SchemaFromStruct(CreateRequest{}),
		OutputSchema: utils.SchemaFromStruct(OperationResult{}),
	}, ct.CreateCalendarItem)
	if err != nil {
		slog.Error("Failed to create create_calendar_item tool", "error", err)
		return nil, err
	}
	tools = append(tools, createItemTool)

	getItemTool, err := functiontool.New(functiontool.Config{
		Name:         "get_calendar_item",
		Description:  "Get a calendar item by ID or UID. Requires either id or uid.",
		InputSchema:  utils.SchemaFromStruct(GetRequest{}),
		OutputSchema: utils.SchemaFromStruct(ItemDetail{}),
	}, ct.GetCalendarItem)
	if err != nil {
		slog.Error("Failed to create get_calendar_item tool", "error", err)
		return nil, err
	}
	tools = append(tools, getItemTool)

	updateItemTool, err := functiontool.New(functiontool.Config{
		Name:         "update_calendar_item",
		Description:  "Update a calendar item. Requires id and optional fields to update.",
		InputSchema:  utils.SchemaFromStruct(UpdateRequest{}),
		OutputSchema: utils.SchemaFromStruct(OperationResult{}),
	}, ct.UpdateCalendarItem)
	if err != nil {
		slog.Error("Failed to create update_calendar_item tool", "error", err)
		return nil, err
	}
	tools = append(tools, updateItemTool)

	deleteItemTool, err := functiontool.New(functiontool.Config{
		Name:         "delete_calendar_item",
		Description:  "Delete a calendar item by ID. Returns success even if the item doesn't exist (idempotent).",
		InputSchema:  utils.SchemaFromStruct(DeleteRequest{}),
		OutputSchema: utils.SchemaFromStruct(OperationResult{}),
	}, ct.DeleteCalendarItem)
	if err != nil {
		slog.Error("Failed to create delete_calendar_item tool", "error", err)
		return nil, err
	}
	tools = append(tools, deleteItemTool)

	searchItemsTool, err := functiontool.New(functiontool.Config{
		Name:         "search_calendar_items",
		Description:  "Search calendar items by keyword (q) and/or time ranges. The keyword will be searched across all searchable fields (summary, description, location, organizer, comment, contact, categories, resources). At least one search criteria (q or dtstart) is required.",
		InputSchema:  utils.SchemaFromStruct(SearchRequest{}),
		OutputSchema: utils.SchemaFromStruct(SearchResponse{}),
	}, ct.SearchCalendarItems)
	if err != nil {
		slog.Error("Failed to create search_calendar_items tool", "error", err)
		return nil, err
	}
	tools = append(tools, searchItemsTool)

	return tools, nil
}

func (ct *calendarTools) CreateCalendarItem(ctx tool.Context, input CreateRequest) (*OperationResult, error) {
	userID := getUserID(ctx)
	slog.Info("Creating calendar item", "type", input.Type, "uid", input.UID)

	if input.UID != nil && *input.UID != "" {
		existingItem, err := ct.service.GetCalendarItemByUID(&userID, *input.UID)
		if err == nil && existingItem != nil {
			slog.Info("Calendar item already exists, returning existing item", "uid", *input.UID, "id", existingItem.ID)
			return &OperationResult{
				Success: true,
				Message: "Calendar item already exists, returning existing item",
				ID:      &existingItem.ID,
				UID:     &existingItem.UID,
				Created: false,
				Item:    convertToDetailResponse(existingItem),
			}, nil
		}
	}

	itemType := calendar.CalendarItemType(input.Type)
	if !isValidCalendarItemType(itemType) {
		slog.Warn("Invalid calendar item type", "type", input.Type)
		return &OperationResult{
			Success: false,
			Message: "Invalid calendar item type",
		}, calendar.ErrInvalidType
	}

	dtStart, err := parseOptionalTime(input.DtStart)
	if err != nil {
		slog.Warn("Failed to parse dtstart", "error", err)
		return &OperationResult{
			Success: false,
			Message: "Failed to parse dtstart: " + err.Error(),
		}, err
	}

	dtEnd, err := parseOptionalTime(input.DtEnd)
	if err != nil {
		slog.Warn("Failed to parse dtend", "error", err)
		return &OperationResult{
			Success: false,
			Message: "Failed to parse dtend: " + err.Error(),
		}, err
	}

	due, err := parseOptionalTime(input.Due)
	if err != nil {
		slog.Warn("Failed to parse due", "error", err)
		return &OperationResult{
			Success: false,
			Message: "Failed to parse due: " + err.Error(),
		}, err
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
		return &OperationResult{
			Success: false,
			Message: "Failed to create calendar item: " + err.Error(),
		}, err
	}

	item, err := ct.service.GetCalendarItemByID(&userID, resp.ID)
	if err != nil {
		slog.Warn("Failed to get created item details", "id", resp.ID, "error", err)
		return &OperationResult{
			Success: true,
			Message: "Calendar item created successfully",
			ID:      &resp.ID,
			UID:     &resp.UID,
			Created: true,
		}, nil
	}

	slog.Info("Calendar item created successfully", "id", item.ID, "uid", item.UID, "type", item.Type)
	return &OperationResult{
		Success: true,
		Message: "Calendar item created successfully",
		ID:      &item.ID,
		UID:     &item.UID,
		Created: true,
		Item:    convertToDetailResponse(item),
	}, nil
}

func (ct *calendarTools) GetCalendarItem(ctx tool.Context, input GetRequest) (*ItemDetail, error) {
	userID := getUserID(ctx)

	if input.ID == nil && input.UID == nil {
		slog.Warn("GetCalendarItem: neither id nor uid provided")
		return nil, calendar.ErrInvalidInput
	}

	var item *calendar.CalendarItem
	var err error

	if input.ID != nil {
		slog.Debug("Getting calendar item by ID", "id", *input.ID)
		item, err = ct.service.GetCalendarItemByID(&userID, *input.ID)
	} else if input.UID != nil {
		slog.Debug("Getting calendar item by UID", "uid", *input.UID)
		item, err = ct.service.GetCalendarItemByUID(&userID, *input.UID)
	}

	if err != nil {
		slog.Error("Failed to get calendar item", "id", input.ID, "uid", input.UID, "error", err)
		return nil, err
	}

	slog.Debug("Calendar item retrieved successfully", "id", item.ID, "uid", item.UID)
	return convertToDetailResponse(item), nil
}

func (ct *calendarTools) UpdateCalendarItem(ctx tool.Context, input UpdateRequest) (*OperationResult, error) {
	userID := getUserID(ctx)
	slog.Info("Updating calendar item", "id", input.ID)

	item, err := ct.service.UpdateCalendarItem(&userID, input.ID, &input.UpdateCalendarItemRequest)
	if err != nil {
		slog.Error("Failed to update calendar item", "id", input.ID, "error", err)
		return &OperationResult{
			Success: false,
			Message: "Failed to update calendar item: " + err.Error(),
			ID:      &input.ID,
		}, err
	}

	slog.Info("Calendar item updated successfully", "id", item.ID, "uid", item.UID)
	return &OperationResult{
		Success: true,
		Message: "Calendar item updated successfully",
		ID:      &item.ID,
		UID:     &item.UID,
		Updated: true,
		Item:    convertToDetailResponse(item),
	}, nil
}

func (ct *calendarTools) DeleteCalendarItem(ctx tool.Context, input DeleteRequest) (*OperationResult, error) {
	userID := getUserID(ctx)
	slog.Info("Deleting calendar item", "id", input.ID)

	err := ct.service.DeleteCalendarItem(&userID, input.ID)
	if err != nil && err != calendar.ErrCalendarItemNotFound {
		slog.Error("Failed to delete calendar item", "id", input.ID, "error", err)
		return &OperationResult{
			Success: false,
			Message: "Failed to delete calendar item: " + err.Error(),
			ID:      &input.ID,
		}, err
	}

	message := "Calendar item deleted successfully"
	if err == calendar.ErrCalendarItemNotFound {
		slog.Info("Calendar item not found (idempotent operation)", "id", input.ID)
		message = "Calendar item not found or already deleted (idempotent)"
	} else {
		slog.Info("Calendar item deleted successfully", "id", input.ID)
	}

	return &OperationResult{
		Success: true,
		Message: message,
		ID:      &input.ID,
		Deleted: true,
	}, nil
}

func (ct *calendarTools) SearchCalendarItems(ctx tool.Context, input SearchRequest) (*SearchResponse, error) {
	userID := getUserID(ctx)

	if input.Q == nil && input.DtStart == nil {
		slog.Warn("SearchCalendarItems: no search criteria provided")
		return &SearchResponse{
			Items:   []*Item{},
			Total:   0,
			Limit:   input.Limit,
			HasMore: false,
		}, calendar.ErrInvalidInput
	}

	slog.Debug("Searching calendar items", "q", input.Q, "limit", input.Limit)

	req := calendar.SearchCalendarItemsRequest{
		Q:     input.Q,
		Limit: input.Limit,
	}

	if input.DtStart != nil {
		dtStart, err := convertTimeRangeInput(input.DtStart)
		if err != nil {
			slog.Warn("Failed to parse time range", "error", err)
			return &SearchResponse{
				Items:   []*Item{},
				Total:   0,
				HasMore: false,
			}, err
		}
		if dtStart != nil {
			req.DtStart = dtStart
		}
	}

	items, err := ct.service.SearchCalendarItems(&userID, &req)
	if err != nil {
		slog.Error("Failed to search calendar items", "error", err)
		return &SearchResponse{
			Items:   []*Item{},
			Total:   0,
			HasMore: false,
		}, err
	}

	summaries := make([]*Item, 0, len(items))
	for _, item := range items {
		summaries = append(summaries, convertToResponse(item))
	}

	limit := 20
	if input.Limit != nil {
		limit = *input.Limit
	}
	hasMore := len(items) >= limit

	slog.Info("Calendar items search completed", "total", len(summaries), "has_more", hasMore)
	return &SearchResponse{
		Items:   summaries,
		Total:   len(summaries),
		Limit:   input.Limit,
		HasMore: hasMore,
	}, nil
}

func getUserID(ctx tool.Context) uint {
	// TODO: extract user ID from context
	return 2
}

func isValidCalendarItemType(t calendar.CalendarItemType) bool {
	return t == calendar.CalendarItemTypeEvent || t == calendar.CalendarItemTypeTodo ||
		t == calendar.CalendarItemTypeJournal || t == calendar.CalendarItemTypeFreeBusy
}

func parseOptionalTime(timeStr *string) (*time.Time, error) {
	if timeStr == nil {
		return nil, nil
	}
	parsed, err := utils.ParseDateTime(*timeStr)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func convertTimeRangeInput(tr *TimeRange) (*calendar.TimeRange, error) {
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

func convertToDetailResponse(item *calendar.CalendarItem) *ItemDetail {
	base := convertToResponse(item)
	return &ItemDetail{
		Item:         *base,
		UID:          item.UID,
		Duration:     item.Duration,
		Sequence:     item.Sequence,
		RRule:        item.RRule,
		ExDate:       []string(item.ExDate),
		RDate:        []string(item.RDate),
		Comment:      item.Comment,
		Contact:      item.Contact,
		RelatedTo:    item.RelatedTo,
		Resources:    []string(item.Resources),
		URL:          item.URL,
		Class:        item.Class,
		LastModified: item.LastModified,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
		Alarms:       item.Alarms,
	}
}

func convertToResponse(item *calendar.CalendarItem) *Item {
	return &Item{
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
		Categories:      []string(item.Categories),
	}
}
