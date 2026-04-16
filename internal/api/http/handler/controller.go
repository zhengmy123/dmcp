package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/service"

	"github.com/gin-gonic/gin"
)

// Controller 处理HTTP服务的API请求
// CRUD 操作只操作数据库，内存通过定时同步获取；execute/debug 走内存 manager
type Controller struct {
	store   repository.ServiceStore
	manager *service.HTTPServiceManager
	logger  logger.Logger
}

// NewController 创建新的控制器
func NewController(store repository.ServiceStore, manager *service.HTTPServiceManager, log logger.Logger) *Controller {
	return &Controller{
		store:   store,
		manager: manager,
		logger:  log,
	}
}

// RegisterRoutes 注册路由到Gin引擎
func (c *Controller) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/services", c.listServices)
	router.GET("/services/:id", c.getService)
	router.POST("/services", c.createService)
	router.PUT("/services/:id", c.updateService)
	router.DELETE("/services/:id", c.deleteService)
	router.POST("/execute/:id", c.executeService)
	router.POST("/services/:id/debug", c.debugService)
}

func (c *Controller) listServices(ctx *gin.Context) {
	services, err := c.store.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list services",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"services": services,
		"count":    len(services),
	})
}

func (c *Controller) getService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service id",
			"id":    idParam,
		})
		return
	}

	service, err := c.store.Get(ctx.Request.Context(), uint(serviceID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "service not found",
			"id":    idParam,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"service": service,
	})
}

func (c *Controller) createService(ctx *gin.Context) {
	var service model.HTTPService
	if err := ctx.ShouldBindJSON(&service); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	service.Method = strings.ToUpper(service.Method)

	// 初始化默认Schema
	if len(service.InputSchema) == 0 {
		service.InputSchema = json.RawMessage(`{"type":"object","properties":{}}`)
	}
	if len(service.OutputSchema) == 0 {
		service.OutputSchema = json.RawMessage(`{"type":"object","properties":{}}`)
	}

	if err := c.store.Save(ctx.Request.Context(), &service); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "failed to create service",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "service created successfully",
		"service_id": service.ID,
		"service":    service,
	})
}

func (c *Controller) updateService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service id",
			"id":    idParam,
		})
		return
	}

	// 先从数据库获取现有服务
	existing, err := c.store.Get(ctx.Request.Context(), uint(serviceID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "service not found",
			"id":    idParam,
		})
		return
	}

	var updates model.HTTPService
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	if updates.Method != "" {
		updates.Method = strings.ToUpper(updates.Method)
	}

	// 合并更新
	updates.ID = existing.ID
	if updates.Name == "" {
		updates.Name = existing.Name
	}
	if updates.TargetURL == "" {
		updates.TargetURL = existing.TargetURL
	}

	if err := c.store.Save(ctx.Request.Context(), &updates); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update service",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "service updated successfully",
		"service_id": serviceID,
	})
}

func (c *Controller) deleteService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service id",
			"id":    idParam,
		})
		return
	}

	if err := c.store.Delete(ctx.Request.Context(), uint(serviceID)); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":   "failed to delete service",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "service deleted successfully",
		"service_id": serviceID,
	})
}

func (c *Controller) executeService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service id",
			"id":    idParam,
		})
		return
	}

	var reqData model.RequestData
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	response, err := c.manager.ExecuteService(ctx.Request.Context(), uint(serviceID), &reqData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to execute service",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

// DebugRequest 调试请求
type DebugRequest struct {
	Headers  map[string]string `json:"headers,omitempty"`
	Body     interface{}       `json:"body,omitempty"`
	Query    map[string]string `json:"query,omitempty"`
	BodyType string            `json:"body_type,omitempty"`
}

// DebugResponse 调试响应
type DebugResponse struct {
	Success         bool              `json:"success"`
	StatusCode      int               `json:"status_code,omitempty"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	RequestBody     interface{}       `json:"request_body,omitempty"`
	ResponseBody    interface{}       `json:"response_body,omitempty"`
	DurationMs      int64             `json:"duration_ms,omitempty"`
	Error           string            `json:"error,omitempty"`
	InputSchema     json.RawMessage   `json:"input_schema,omitempty"`
	OutputSchema    json.RawMessage   `json:"output_schema,omitempty"`
}

func (c *Controller) debugService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service id",
			"id":    idParam,
		})
		return
	}

	service, exists := c.manager.GetService(uint(serviceID))
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "service not found",
			"id":    idParam,
		})
		return
	}

	var debugReq DebugRequest
	if err := ctx.ShouldBindJSON(&debugReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	reqData := &model.RequestData{
		Headers: debugReq.Headers,
		Body:    debugReq.Body,
		Query:   debugReq.Query,
	}

	// 调试时用前端指定的 body_type，创建副本避免修改原始服务
	debugSvc := *service
	if debugReq.BodyType != "" {
		debugSvc.BodyType = debugReq.BodyType
	}

	if reqData.Headers == nil {
		reqData.Headers = make(map[string]string)
	}
	if reqData.Query == nil {
		reqData.Query = make(map[string]string)
	}

	// 记录debug请求输入
	log := c.logger.Ctx(ctx.Request.Context())
	log.Info("Debug request received",
		logger.Uint("service_id", uint(serviceID)),
		logger.String("service_name", service.Name),
		logger.String("target_url", service.TargetURL),
		logger.String("method", service.Method),
		logger.Any("request_headers", reqData.Headers),
		logger.Any("request_body", reqData.Body),
		logger.Any("request_query", reqData.Query),
	)

	response, err := c.manager.ExecuteServiceWithOverride(ctx.Request.Context(), &debugSvc, reqData)

	debugResp := DebugResponse{
		RequestHeaders:  service.Headers,
		InputSchema:     service.InputSchema,
		OutputSchema:    service.OutputSchema,
		RequestBody:     reqData.Body,
	}

	if err != nil {
		debugResp.Error = err.Error()
		debugResp.Success = false
		if response != nil {
			debugResp.StatusCode = response.StatusCode
			debugResp.ResponseBody = response.Body
			debugResp.ResponseHeaders = response.Headers
			debugResp.DurationMs = response.Duration.Milliseconds()
		}

		// 记录debug请求失败输出
		log.Error("Debug request failed",
			logger.Uint("service_id", uint(serviceID)),
			logger.String("service_name", service.Name),
			logger.Bool("success", debugResp.Success),
			logger.Int("status_code", debugResp.StatusCode),
			logger.Int64("duration_ms", debugResp.DurationMs),
			logger.Any("response_body", debugResp.ResponseBody),
			logger.String("error", debugResp.Error),
		)
	} else {
		debugResp.Success = response.StatusCode >= 200 && response.StatusCode < 300
		debugResp.StatusCode = response.StatusCode
		debugResp.ResponseHeaders = response.Headers
		debugResp.ResponseBody = response.Body
		debugResp.DurationMs = response.Duration.Milliseconds()
		if response.Error != "" {
			debugResp.Error = response.Error
		}

		// 记录debug请求成功输出
		log.Info("Debug request completed",
			logger.Uint("service_id", uint(serviceID)),
			logger.String("service_name", service.Name),
			logger.Bool("success", debugResp.Success),
			logger.Int("status_code", debugResp.StatusCode),
			logger.Int64("duration_ms", debugResp.DurationMs),
			logger.Any("response_body", debugResp.ResponseBody),
			logger.String("error", debugResp.Error),
		)
	}

	ctx.JSON(http.StatusOK, debugResp)
}

// WebhookHandler 处理Webhook请求
func (c *Controller) WebhookHandler(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid service id",
			"id":    idParam,
		})
		return
	}

	reqData := model.RequestData{
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	for k, v := range ctx.Request.Header {
		if len(v) > 0 {
			reqData.Headers[k] = v[0]
		}
	}

	for k, v := range ctx.Request.URL.Query() {
		if len(v) > 0 {
			reqData.Query[k] = v[0]
		}
	}

	var bodyObj interface{}
	bodyBytes, err := ctx.GetRawData()
	if err == nil && len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &bodyObj); err != nil {
			bodyObj = string(bodyBytes)
		}
		reqData.Body = bodyObj
	}

	response, err := c.manager.ExecuteService(ctx.Request.Context(), uint(serviceID), &reqData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to execute webhook",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"webhook_processed": true,
		"service_id":        serviceID,
		"response":          response,
	})
}
