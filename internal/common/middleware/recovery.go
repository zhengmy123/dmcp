package middleware

import (
	"fmt"
	"runtime/trace"

	"github.com/gin-gonic/gin"
	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/common/response"
)

func Recovery(log logger.Logger) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, err interface{}) {
		requestID, _ := c.Get("request_id")

		loggerFields := []logger.Field{
			logger.Any("error", err),
			logger.String("request_id", fmt.Sprintf("%v", requestID)),
			logger.String("path", c.Request.URL.Path),
			logger.String("method", c.Request.Method),
		}

		log.Ctx(c.Request.Context()).Error("panic recovered", loggerFields...)

		trace.Logf(c.Request.Context(), "ERROR", "panic recovered: %v", err)

		errMsg := "internal server error"
		if err != nil {
			errMsg = fmt.Sprintf("%v", err)
		}

		response.InternalError(c, errMsg)
	})
}
