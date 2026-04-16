package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"dynamic_mcp_go_server/internal/common/logger"
)

// RequestID 链路追踪中间件 — 生成高效 UUID v4 作为 trace_id / request_id
// 如果上游已传入 X-Request-ID 或 X-Trace-ID 则复用，保证跨服务链路贯通
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = c.GetHeader("X-Request-ID")
		}
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// 同时写入 gin context（兼容旧代码读取 request_id）和标准 context
		c.Set("request_id", traceID)
		c.Set("trace_id", traceID)
		c.Header("X-Request-ID", traceID)
		c.Header("X-Trace-ID", traceID)

		// ★ 将 trace_id 注入 request context，下游通过 logger.Ctx(ctx) 自动获取
		ctx := logger.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
