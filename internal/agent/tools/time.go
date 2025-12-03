package tools

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

type GetCurrentTimeRequest struct {
	TimeZone string `json:"timezone,omitempty"`
}

type GetCurrentTimeResponse struct {
	Time     string `json:"time"`
	Date     string `json:"date"`
	TimeZone string `json:"time_zone"`
	Unix     int64  `json:"unix"`
}

func getCurrentTime(ctx tool.Context, input GetCurrentTimeRequest) (GetCurrentTimeResponse, error) {
	// 默认时区为 Asia/Shanghai
	timezone := input.TimeZone
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}

	// 加载时区
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return GetCurrentTimeResponse{}, fmt.Errorf("无效的时区: %w", err)
	}

	// 获取当前时间
	now := time.Now().In(loc)

	// 格式化时间和日期
	timeStr := now.Format("15:04:05")
	dateStr := now.Format("2006-01-02")
	unix := now.Unix()

	slog.Debug("获取当前时间", "timezone", timezone, "time", timeStr, "date", dateStr)

	return GetCurrentTimeResponse{
		Time:     timeStr,
		Date:     dateStr,
		TimeZone: timezone,
		Unix:     unix,
	}, nil
}

func NewTimeTools() ([]tool.Tool, error) {
	tools := []tool.Tool{}

	timeTool, err := functiontool.New(functiontool.Config{
		Name:        "get_current_time",
		Description: "Get the current time and date in the specified timezone. Default timezone is Asia/Shanghai if not specified.",
		InputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"timezone": {
					Type:        "string",
					Description: "Timezone name (e.g., Asia/Shanghai, America/New_York, UTC). Default is Asia/Shanghai.",
				},
			},
		},
		OutputSchema: &jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"time": {
					Type:        "string",
					Description: "Current time in HH:MM:SS format",
				},
				"date": {
					Type:        "string",
					Description: "Current date in YYYY-MM-DD format",
				},
				"time_zone": {
					Type:        "string",
					Description: "Timezone used",
				},
				"unix": {
					Type:        "integer",
					Description: "Unix timestamp",
				},
			},
			Required: []string{"time", "date", "time_zone", "unix"},
		},
	}, getCurrentTime)
	if err != nil {
		slog.Error("Failed to create get_current_time tool", "error", err)
		return nil, err
	}
	tools = append(tools, timeTool)

	return tools, nil
}
