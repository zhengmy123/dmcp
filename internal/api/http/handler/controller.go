package handler

import (
	"strconv"
	"strings"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/common/response"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/service"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	store   repository.ServiceStore
	manager *service.HTTPServiceManager
	logger  logger.Logger
}

func NewController(store repository.ServiceStore, manager *service.HTTPServiceManager, log logger.Logger) *Controller {
	return &Controller{
		store:   store,
		manager: manager,
		logger:  log,
	}
}

func (c *Controller) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/services", c.listServices)
	router.GET("/services/simple", c.GetServicesSimple)
	router.GET("/services/:id", c.getService)
	router.POST("/services", c.createService)
	router.PUT("/services/:id", c.updateService)
	router.DELETE("/services/:id", c.deleteService)
	router.POST("/execute/:id", c.executeService)
	router.POST("/services/:id/debug", c.debugService)
}

func (c *Controller) listServices(ctx *gin.Context) {
	query := &model.ServiceQuery{}

	// 解析名称搜索参数
	name := ctx.Query("name")
	if name != "" {
		query.Name = &name
	}

	// 解析状态筛选参数
	stateStr := ctx.Query("state")
	if stateStr != "" {
		state, err := strconv.Atoi(stateStr)
		if err == nil && (state == 0 || state == 1) {
			query.State = &state
		}
	}

	// 解析分页参数
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	query.Page = page

	pageSize, err := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	query.PageSize = pageSize

	services, total, err := c.store.ListWithQuery(ctx.Request.Context(), query)
	if err != nil {
		response.InternalError(ctx, "failed to list services: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"services":  services,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (c *Controller) GetServicesSimple(ctx *gin.Context) {
	services, err := c.store.List(ctx.Request.Context())
	if err != nil {
		response.InternalError(ctx, "failed to list services: "+err.Error())
		return
	}

	type SimpleService struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	simpleServices := make([]SimpleService, 0, len(services))
	for _, svc := range services {
		simpleServices = append(simpleServices, SimpleService{
			ID:   svc.ID,
			Name: svc.Name,
		})
	}

	response.Success(ctx, gin.H{
		"services": simpleServices,
		"count":    len(simpleServices),
	})
}

func (c *Controller) getService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid service id")
		return
	}

	service, err := c.store.Get(ctx.Request.Context(), uint(serviceID))
	if err != nil {
		response.NotFound(ctx, "service not found")
		return
	}

	response.Success(ctx, gin.H{
		"service": service,
	})
}

func (c *Controller) createService(ctx *gin.Context) {
	var service model.HTTPService
	if err := ctx.ShouldBindJSON(&service); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	service.Method = strings.ToUpper(service.Method)

	if len(service.InputSchema) == 0 {
		service.InputSchema = []byte(`{"type":"object","properties":{}}`)
	}
	if len(service.OutputSchema) == 0 {
		service.OutputSchema = []byte(`{"type":"object","properties":{}}`)
	}

	if err := c.store.Save(ctx.Request.Context(), &service); err != nil {
		response.InternalError(ctx, "failed to create service: "+err.Error())
		return
	}

	response.Created(ctx, gin.H{
		"service_id": service.ID,
		"service":    service,
	})
}

func (c *Controller) updateService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid service id")
		return
	}

	existing, err := c.store.Get(ctx.Request.Context(), uint(serviceID))
	if err != nil {
		response.NotFound(ctx, "service not found")
		return
	}

	var updates model.HTTPService
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	if updates.Method != "" {
		updates.Method = strings.ToUpper(updates.Method)
	}

	updates.ID = existing.ID
	if updates.Name == "" {
		updates.Name = existing.Name
	}
	if updates.TargetURL == "" {
		updates.TargetURL = existing.TargetURL
	}

	if err := c.store.Save(ctx.Request.Context(), &updates); err != nil {
		response.InternalError(ctx, "failed to update service: "+err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "service updated successfully", gin.H{
		"service_id": serviceID,
	})
}

func (c *Controller) deleteService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid service id")
		return
	}

	if err := c.store.Delete(ctx.Request.Context(), uint(serviceID)); err != nil {
		response.NotFound(ctx, "failed to delete service: "+err.Error())
		return
	}

	response.SuccessWithMessage(ctx, "service deleted successfully", gin.H{
		"service_id": serviceID,
	})
}

func (c *Controller) executeService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid service id")
		return
	}

	var reqData model.RequestData
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	execResp, err := c.manager.ExecuteService(ctx.Request.Context(), uint(serviceID), &reqData)
	if err != nil {
		response.InternalError(ctx, "failed to execute service: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"response": execResp,
	})
}

type DebugRequest struct {
	Headers  map[string]string `json:"headers,omitempty"`
	Body     interface{}       `json:"body,omitempty"`
	Query    map[string]string `json:"query,omitempty"`
	BodyType string            `json:"body_type,omitempty"`
}

type DebugResponse struct {
	Success         bool              `json:"success"`
	StatusCode      int               `json:"status_code,omitempty"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	RequestBody     interface{}       `json:"request_body,omitempty"`
	ResponseBody    interface{}       `json:"response_body,omitempty"`
	DurationMs      int64             `json:"duration_ms,omitempty"`
	Error           string            `json:"error,omitempty"`
	InputSchema     []byte            `json:"input_schema,omitempty"`
	OutputSchema    []byte            `json:"output_schema,omitempty"`
}

func (c *Controller) debugService(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid service id")
		return
	}

	service, exists := c.manager.GetService(uint(serviceID))
	if !exists {
		response.NotFound(ctx, "service not found")
		return
	}

	var debugReq DebugRequest
	if err := ctx.ShouldBindJSON(&debugReq); err != nil {
		response.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	reqData := &model.RequestData{
		Headers: debugReq.Headers,
		Body:    debugReq.Body,
		Query:   debugReq.Query,
	}

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

	execResp, err := c.manager.ExecuteServiceWithOverride(ctx.Request.Context(), &debugSvc, reqData)

	debugResp := DebugResponse{
		RequestHeaders:  service.Headers,
		InputSchema:     service.InputSchema,
		OutputSchema:    service.OutputSchema,
		RequestBody:     reqData.Body,
	}

	if err != nil {
		debugResp.Error = err.Error()
		debugResp.Success = false
		if execResp != nil {
			debugResp.StatusCode = execResp.StatusCode
			debugResp.ResponseBody = execResp.Body
			debugResp.ResponseHeaders = execResp.Headers
			debugResp.DurationMs = execResp.Duration.Milliseconds()
		}

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
		debugResp.Success = execResp.StatusCode >= 200 && execResp.StatusCode < 300
		debugResp.StatusCode = execResp.StatusCode
		debugResp.ResponseHeaders = execResp.Headers
		debugResp.ResponseBody = execResp.Body
		debugResp.DurationMs = execResp.Duration.Milliseconds()
		if execResp.Error != "" {
			debugResp.Error = execResp.Error
		}

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

	response.Success(ctx, gin.H{
		"response": debugResp,
	})
}

func (c *Controller) WebhookHandler(ctx *gin.Context) {
	idParam := ctx.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid service id")
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
		if err := sonic.Unmarshal(bodyBytes, &bodyObj); err != nil {
			bodyObj = string(bodyBytes)
		}
		reqData.Body = bodyObj
	}

	execResp, err := c.manager.ExecuteService(ctx.Request.Context(), uint(serviceID), &reqData)
	if err != nil {
		response.InternalError(ctx, "failed to execute webhook: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"webhook_processed": true,
		"service_id":        serviceID,
		"response":          execResp,
	})
}