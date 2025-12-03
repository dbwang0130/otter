package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceIDKey context 中 traceid 的键
type TraceIDKey struct{}

// TraceID 提取或生成 traceid 的中间件
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头中获取 traceid
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = c.GetHeader("X-Request-Id")
		}
		if traceID == "" {
			// 如果没有，生成一个新的 traceid
			traceID = uuid.New().String()
		}

		// 将 traceid 存储到 context 中
		ctx := context.WithValue(c.Request.Context(), TraceIDKey{}, traceID)
		c.Request = c.Request.WithContext(ctx)

		// 将 traceid 存储到 gin context 中，方便后续使用
		c.Set("trace_id", traceID)

		// 在响应头中返回 traceid
		c.Writer.Header().Set("X-Trace-Id", traceID)

		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 从 context 中获取 traceid
		traceID := ""
		if traceIDVal, exists := param.Keys["trace_id"]; exists {
			if id, ok := traceIDVal.(string); ok {
				traceID = id
			}
		}

		// 如果没有从 gin context 获取到，尝试从 request context 获取
		if traceID == "" {
			if id := param.Request.Context().Value(TraceIDKey{}); id != nil {
				if idStr, ok := id.(string); ok {
					traceID = idStr
				}
			}
		}

		attrs := []any{
			"timestamp", param.TimeStamp.Format(time.RFC1123),
			"client_ip", param.ClientIP,
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"user_agent", param.Request.UserAgent(),
		}

		// 如果有 traceid，添加到日志属性中
		if traceID != "" {
			attrs = append(attrs, "trace_id", traceID)
		}

		if param.ErrorMessage != "" {
			attrs = append(attrs, "error", param.ErrorMessage)
		}

		slog.Debug("HTTP请求", attrs...)
		return ""
	})
}

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 从 context 中获取 traceid
		traceID := ""
		if traceIDVal, exists := c.Get("trace_id"); exists {
			if id, ok := traceIDVal.(string); ok {
				traceID = id
			}
		}
		if traceID == "" {
			if id := c.Request.Context().Value(TraceIDKey{}); id != nil {
				if idStr, ok := id.(string); ok {
					traceID = idStr
				}
			}
		}

		attrs := []any{"panic", recovered}
		if traceID != "" {
			attrs = append(attrs, "trace_id", traceID)
		}
		slog.Error("Panic recovered", attrs...)
		c.JSON(500, gin.H{
			"error": "内部服务器错误",
		})
		c.Abort()
	})
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
