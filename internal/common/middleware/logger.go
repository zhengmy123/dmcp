package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"dynamic_mcp_go_server/internal/common/logger"
)

// Logger 请求日志中间件
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查接口的日志
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []logger.Field{
			logger.Int("status", status),
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.String("query", query),
			logger.Duration("latency", latency),
			logger.String("ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
		}

		// trace_id 已通过 Ctx(ctx) 自动注入，这里只追加 request_id 兼容旧日志检索
		if requestID, exists := c.Get("request_id"); exists {
			fields = append(fields, logger.String("request_id", fmt.Sprintf("%v", requestID)))
		}

		// 添加用户信息
		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, logger.String("user_id", fmt.Sprintf("%v", userID)))
		}

		// ★ 使用 Ctx(ctx) 让 trace_id 自动注入
		ctxLog := log.Ctx(c.Request.Context())

		// 根据状态码选择日志级别
		if status >= 500 {
			ctxLog.Error("Server error", fields...)
		} else if status >= 400 {
			ctxLog.Warn("Client error", fields...)
		} else {
			ctxLog.Info("Request completed", fields...)
		}
	}
}
