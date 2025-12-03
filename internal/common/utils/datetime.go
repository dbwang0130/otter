package utils

import (
	"fmt"
	"strings"
	"time"
)

// ParseDateTime 解析日期时间字符串，支持多种格式
// 支持的格式：
// 1. RFC3339 格式（包含时区）：2006-01-02T15:04:05Z07:00, 2006-01-02T15:04:05Z
// 2. RFC3339Nano 格式：2006-01-02T15:04:05.999999999Z07:00
// 3. ISO 8601 格式（无时区，假设为 UTC）：2006-01-02T15:04:05, 2006-01-02T15:04
// 4. 日期时间格式（空格分隔，假设为 UTC）：2006-01-02 15:04:05, 2006-01-02 15:04
// 5. 日期格式（假设为 UTC 00:00:00）：2006-01-02
func ParseDateTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("时间字符串不能为空")
	}

	// 尝试 RFC3339 格式（包含时区）
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, nil
	}

	// 尝试 RFC3339Nano 格式
	if t, err := time.Parse(time.RFC3339Nano, timeStr); err == nil {
		return t, nil
	}

	// 检查是否包含时区信息
	hasTimezone := strings.HasSuffix(timeStr, "Z") ||
		strings.Contains(timeStr, "+") ||
		(len(timeStr) > 19 && (strings.Contains(timeStr[19:], "+") || strings.Contains(timeStr[19:], "-")))

	// 尝试没有时区的 ISO 8601 格式：2006-01-02T15:04:05
	if strings.Contains(timeStr, "T") && !hasTimezone {
		layouts := []string{
			"2006-01-02T15:04:05",
			"2006-01-02T15:04",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, timeStr); err == nil {
				return t.UTC(), nil
			}
		}
	}

	// 尝试日期时间格式：2006-01-02 15:04:05
	if strings.Contains(timeStr, " ") {
		layouts := []string{
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, timeStr); err == nil {
				return t.UTC(), nil
			}
		}
	}

	// 尝试日期格式：2006-01-02
	if t, err := time.Parse("2006-01-02", timeStr); err == nil {
		return t.UTC(), nil
	}

	// 如果所有格式都失败，返回错误
	return time.Time{}, fmt.Errorf("无法解析时间字符串 '%s'，支持的格式：RFC3339 (2006-01-02T15:04:05Z07:00)、ISO 8601 (2006-01-02T15:04:05)、日期时间 (2006-01-02 15:04:05)、日期 (2006-01-02)", timeStr)
}
