package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/galilio/otter/internal/common/config"
	"github.com/galilio/otter/internal/common/middleware"
	"github.com/gin-gonic/gin"
)

// traceIDHandler 包装 slog handler，自动从 context 中提取 traceid
type traceIDHandler struct {
	slog.Handler
}

// Handle 处理日志记录，自动添加 traceid
func (h *traceIDHandler) Handle(ctx context.Context, r slog.Record) error {
	// 从 context 中提取 traceid
	if traceID := ctx.Value(middleware.TraceIDKey{}); traceID != nil {
		if id, ok := traceID.(string); ok && id != "" {
			r.AddAttrs(slog.String("trace_id", id))
		}
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs 返回带有属性的新 handler
func (h *traceIDHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceIDHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup 返回带有组的新 handler
func (h *traceIDHandler) WithGroup(name string) slog.Handler {
	return &traceIDHandler{Handler: h.Handler.WithGroup(name)}
}

// Init 初始化日志系统
func Init(cfg config.LogConfig) error {
	// 解析日志级别
	var logLevel slog.Level
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// 选择输出目标
	var writer io.Writer = os.Stdout

	if cfg.LogFile != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.LogFile)
		if logDir != "" && logDir != "." {
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return fmt.Errorf("创建日志目录失败: %w", err)
			}
		}

		// 配置日志轮转（默认值已在 applyLogDefaults 中设置）
		writer = &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    cfg.MaxSize,    // MB
			MaxBackups: cfg.MaxBackups, // 保留的旧文件数量
			MaxAge:     cfg.MaxAge,     // 保留天数
			Compress:   cfg.Compress,   // 是否压缩轮转后的文件
			LocalTime:  cfg.LocalTime,  // 是否使用本地时间
		}
	}

	// 创建 logger
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	baseHandler := slog.NewTextHandler(writer, opts)
	// 包装 handler 以支持自动提取 traceid
	traceHandler := &traceIDHandler{Handler: baseHandler}
	logger := slog.New(traceHandler)
	slog.SetDefault(logger)

	return nil
}

// FromContext 从 gin.Context 中获取带 traceid 的 logger
// 使用方式：logger.FromContext(c).Info("消息", "key", "value")
// 如果 traceid 存在，会自动添加到所有日志记录中
func FromContext(c *gin.Context) *slog.Logger {
	if c == nil {
		return slog.Default()
	}
	// 从 gin context 或 request context 中获取 traceid
	ctx := c.Request.Context()
	traceID := getTraceIDFromContext(c, ctx)
	if traceID != "" {
		return slog.Default().With("trace_id", traceID)
	}
	return slog.Default()
}

// getTraceIDFromContext 从 gin context 或 request context 中获取 traceid
func getTraceIDFromContext(c *gin.Context, ctx context.Context) string {
	// 先尝试从 gin context 获取
	if traceIDVal, exists := c.Get("trace_id"); exists {
		if id, ok := traceIDVal.(string); ok && id != "" {
			return id
		}
	}
	// 再尝试从 request context 获取
	if id := ctx.Value(middleware.TraceIDKey{}); id != nil {
		if idStr, ok := id.(string); ok && idStr != "" {
			return idStr
		}
	}
	return ""
}
