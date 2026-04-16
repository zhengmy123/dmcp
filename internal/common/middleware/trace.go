package middleware

import (
	"context"
	"fmt"
	"runtime/trace"

	"github.com/gin-gonic/gin"
	"dynamic_mcp_go_server/internal/common/logger"
)

// Trace 链路追踪中间件
// 利用 Go runtime/trace 为每个请求创建 Task + Region
// trace_id 已由 RequestID 中间件注入 context，无需再手动传递
func Trace(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, _ := c.Get("request_id")
		taskName := fmt.Sprintf("%s %s [%v]", c.Request.Method, c.Request.URL.Path, requestID)

		ctx, task := trace.NewTask(c.Request.Context(), taskName)
		defer task.End()

		// 在 Region 内记录请求开始
		region := trace.StartRegion(ctx, "handle_request")
		trace.Logf(ctx, "info", "request started: %s %s", c.Request.Method, c.Request.URL.Path)

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		region.End()
		trace.Logf(ctx, "info", "request completed: %d", c.Writer.Status())
	}
}

// TraceContext 从 gin.Context 提取带链路的 context.Context
func TraceContext(c *gin.Context) context.Context {
	return c.Request.Context()
}

// TraceLog 从 gin.Context 提取带链路信息的 Logger
// 优先从 context 中的 trace_id 构建，兼容旧的 c.Get("logger") 方式
func TraceLog(c *gin.Context) logger.Logger {
	if log, ok := c.Get("logger"); ok {
		return log.(logger.Logger)
	}
	return nil
}
