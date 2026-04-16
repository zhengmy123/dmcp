package middleware

import (
	"fmt"
	"net/http"
	"runtime/trace"

	"github.com/gin-gonic/gin"
	"dynamic_mcp_go_server/internal/common/logger"
)

// Recovery 崩溃恢复中间件，同时记录堆栈和链路
func Recovery(log logger.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, err interface{}) {
		// 获取请求 ID 用于追踪
		requestID, _ := c.Get("request_id")

		loggerFields := []logger.Field{
			logger.Any("error", err),
			logger.String("request_id", fmt.Sprintf("%v", requestID)),
			logger.String("path", c.Request.URL.Path),
			logger.String("method", c.Request.Method),
		}

		// ★ 使用 Ctx(ctx) 自动注入 trace_id
		log.Ctx(c.Request.Context()).Error("panic recovered", loggerFields...)

		// 中止 Region
		trace.Logf(c.Request.Context(), "ERROR", "panic recovered: %v", err)

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":      "internal server error",
			"request_id": requestID,
		})
	})
}
