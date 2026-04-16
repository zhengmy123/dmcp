package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"
)

// RequestData 是 model.RequestData 的别名
type RequestData = model.RequestData

// ResponseData 是 model.ResponseData 的别名
type ResponseData = model.ResponseData

// HTTPServiceManager 管理HTTP服务
type HTTPServiceManager struct {
	services   map[uint]*model.HTTPService
	mu         sync.RWMutex
	logger     logger.Logger
	httpClient *http.Client
	validator  *ScriptValidator
}

// NewHTTPServiceManager 创建新的服务管理器
func NewHTTPServiceManager(log logger.Logger) *HTTPServiceManager {
	return &HTTPServiceManager{
		services: make(map[uint]*model.HTTPService),
		logger:   log,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		validator: NewScriptValidator(log),
	}
}

// RegisterService 注册新的HTTP服务
func (m *HTTPServiceManager) RegisterService(service *model.HTTPService) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if service.TargetURL == "" {
		return fmt.Errorf("target URL is required")
	}

	if service.Method == "" {
		service.Method = "POST"
	}

	if service.TimeoutSeconds <= 0 {
		service.TimeoutSeconds = 30
	}

	now := time.Now()
	if service.CreatedAt.IsZero() {
		service.CreatedAt = now
	}
	service.UpdatedAt = now

	m.services[service.ID] = service

	m.logger.Info("HTTP service registered",
		logger.Uint("service_id", service.ID),
		logger.String("service_name", service.Name),
		logger.String("target_url", service.TargetURL),
	)

	return nil
}

// GetService 获取服务配置
func (m *HTTPServiceManager) GetService(serviceID uint) (*model.HTTPService, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	service, exists := m.services[serviceID]
	return service, exists
}

// ListServices 列出所有服务
func (m *HTTPServiceManager) ListServices() []*model.HTTPService {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]*model.HTTPService, 0, len(m.services))
	for _, service := range m.services {
		services = append(services, service)
	}
	return services
}

// UpdateService 更新服务配置
func (m *HTTPServiceManager) UpdateService(serviceID uint, updates *model.HTTPService) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service not found: %d", serviceID)
	}

	// 更新字段
	if updates.Name != "" {
		service.Name = updates.Name
	}
	if updates.Description != "" {
		service.Description = updates.Description
	}
	if updates.TargetURL != "" {
		service.TargetURL = updates.TargetURL
	}
	if updates.Method != "" {
		service.Method = updates.Method
	}
	if updates.BodyType != "" {
		service.BodyType = updates.BodyType
	}
	if updates.Headers != nil {
		service.Headers = updates.Headers
	}
	if updates.TimeoutSeconds > 0 {
		service.TimeoutSeconds = updates.TimeoutSeconds
	}
	if updates.RetryCount >= 0 {
		service.RetryCount = updates.RetryCount
	}
	if updates.ValidationScript != "" {
		service.ValidationScript = updates.ValidationScript
	}
	service.ValidationEnabled = updates.ValidationEnabled
	if updates.RequestTransformScript != "" {
		service.RequestTransformScript = updates.RequestTransformScript
	}
	if updates.ResponseTransformScript != "" {
		service.ResponseTransformScript = updates.ResponseTransformScript
	}
	if updates.InputSchema != nil {
		service.InputSchema = updates.InputSchema
	}
	if updates.OutputSchema != nil {
		service.OutputSchema = updates.OutputSchema
	}
	service.Enabled = updates.Enabled
	service.UpdatedAt = time.Now()

	m.logger.Info("HTTP service updated",
		logger.Uint("service_id", serviceID),
		logger.String("service_name", service.Name),
	)

	return nil
}

// DeleteService 删除服务
func (m *HTTPServiceManager) DeleteService(serviceID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service not found: %d", serviceID)
	}

	delete(m.services, serviceID)

	m.logger.Info("HTTP service deleted",
		logger.Uint("service_id", serviceID),
		logger.String("service_name", service.Name),
	)

	return nil
}

// ExecuteService 执行HTTP服务调用
func (m *HTTPServiceManager) ExecuteService(ctx context.Context, serviceID uint, reqData *model.RequestData) (*model.ResponseData, error) {
	service, exists := m.GetService(serviceID)
	if !exists {
		return nil, fmt.Errorf("service not found: %d", serviceID)
	}

	return m.executeWithService(ctx, service, reqData)
}

// ExecuteServiceWithOverride 使用指定的服务配置执行请求（用于调试时覆盖 body_type 等配置）
func (m *HTTPServiceManager) ExecuteServiceWithOverride(ctx context.Context, service *model.HTTPService, reqData *model.RequestData) (*model.ResponseData, error) {
	return m.executeWithService(ctx, service, reqData)
}

func (m *HTTPServiceManager) executeWithService(ctx context.Context, service *model.HTTPService, reqData *model.RequestData) (*model.ResponseData, error) {
	startTime := time.Now()

	// 请求转换脚本
	if service.HasRequestTransform() {
		transformResult, err := m.validator.TransformRequest(
			service.RequestTransformScript, reqData.Headers, reqData.Body, service.Method, service.TargetURL,
		)
		if err != nil {
			m.logger.Warn("request transform script failed, using original data",
				logger.Error(err),
				logger.Uint("service_id", service.ID),
			)
		} else {
			m.logger.Info("request transform: script found, executing",
				logger.Uint("service_id", service.ID),
				logger.String("script_preview", previewScript(service.RequestTransformScript)),
				logger.Any("transformResult", transformResult),
			)
			if transformResult.Headers != nil {
				reqData.Headers = transformResult.Headers
			}
			if transformResult.Body != nil {
				reqData.Body = transformResult.Body
			}
			if transformResult.URLParams != nil {
				m.logger.Debug("setting URL params from transform",
					logger.Any("url_params", transformResult.URLParams),
				)
				reqData.Query = transformResult.URLParams
			}
		}
	}

	// 入参JSON Schema校验
	if service.HasInputSchema() {
		if err := m.validateJSONSchema(service.InputSchema, reqData.Body, "input"); err != nil {
			return nil, fmt.Errorf("input schema validation failed: %w", err)
		}
	}

	// 根据 InputSchema 填充默认值
	reqData.Body = fillDefaultsFromSchema(service.InputSchema, reqData.Body)

	// 构建请求
	req, err := m.buildRequest(ctx, service, reqData)
	if err != nil {
		return nil, fmt.Errorf("build request failed: %w", err)
	}

	// 执行请求
	resp, err := m.executeWithRetry(req, service.RetryCount)
	duration := time.Since(startTime)

	response := &model.ResponseData{
		Duration: duration,
	}

	if err != nil {
		response.Error = err.Error()
		return response, err
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Error = fmt.Sprintf("read response body failed: %v", err)
		return response, err
	}

	// 解析响应体
	var bodyObj interface{}
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &bodyObj); err != nil {
			// 如果不是JSON，保持为字符串
			bodyObj = string(bodyBytes)
		}
	}

	// 收集响应头
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	response.StatusCode = resp.StatusCode
	response.Headers = headers
	response.Body = bodyObj

	// 响应转换脚本
	if service.HasResponseTransform() {
		transformResult, err := m.validator.TransformRequest(
			service.ResponseTransformScript, response.Headers, response.Body, service.Method, service.TargetURL,
		)
		if err != nil {
			m.logger.Warn("response transform script failed, using original data",
				logger.Error(err),
				logger.Uint("service_id", service.ID),
			)
		} else {
			if transformResult.Headers != nil {
				response.Headers = transformResult.Headers
			}
			if transformResult.Body != nil {
				response.Body = transformResult.Body
			}
		}
	}

	// 出参JSON Schema校验
	if service.HasOutputSchema() {
		if err := m.validateJSONSchema(service.OutputSchema, response.Body, "output"); err != nil {
			m.logger.Warn("output schema validation failed",
				logger.Error(err),
				logger.Uint("service_id", service.ID),
			)
			// 出参校验失败不阻断，仅记录警告
		}
	}

	return response, nil
}

func (m *HTTPServiceManager) buildRequest(ctx context.Context, service *model.HTTPService, reqData *model.RequestData) (*http.Request, error) {
	// 序列化请求体
	var body io.Reader
	if reqData.Body != nil {
		result, err := model.EncodeBody(service.BodyType, reqData.Body)
		if err != nil {
			return nil, fmt.Errorf("encode request body failed: %w", err)
		}
		body = result.Reader

		// 如果编码器指定了需要覆盖的 Content-Type（如 multipart boundary），设置到请求头
		if result.OverrideContentType != "" {
			if reqData.Headers == nil {
				reqData.Headers = make(map[string]string)
			}
			reqData.Headers["Content-Type"] = result.OverrideContentType
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, service.Method, service.TargetURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	if reqData.Headers != nil {
		for k, v := range reqData.Headers {
			req.Header.Set(k, v)
		}
	}

	// 设置服务配置的请求头
	for k, v := range service.Headers {
		req.Header.Set(k, v)
	}

	// 如果未设置 Content-Type，根据 BodyType 自动设置默认值
	if req.Header.Get("Content-Type") == "" {
		defaultContentType := service.GetDefaultContentType()
		if defaultContentType != "" {
			req.Header.Set("Content-Type", defaultContentType)
			m.logger.Debug("set default Content-Type",
				logger.String("content_type", defaultContentType),
				logger.String("body_type", service.BodyType),
			)
		}
	}

	// 设置URL查询参数（由请求转换脚本设置）
	if reqData.Query != nil {
		m.logger.Debug("building request with URL params",
			logger.Any("query", reqData.Query),
		)
		q := req.URL.Query()
		for k, v := range reqData.Query {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
		m.logger.Debug("URL after setting query params",
			logger.String("url", req.URL.String()),
		)
	}

	// 设置超时
	m.httpClient.Timeout = time.Duration(service.TimeoutSeconds) * time.Second

	return req, nil
}

// fillDefaultsFromSchema 根据 JSON Schema 的 default 字段，为 body 中缺失的属性填充默认值
func fillDefaultsFromSchema(schema json.RawMessage, body interface{}) interface{} {
	if len(schema) == 0 {
		return body
	}

	// 解析 schema
	var schemaObj struct {
		Type       string                 `json:"type"`
		Properties map[string]interface{} `json:"properties"`
	}
	if err := json.Unmarshal(schema, &schemaObj); err != nil {
		return body
	}
	if schemaObj.Type != "object" || len(schemaObj.Properties) == 0 {
		return body
	}

	// 将 body 转为 map
	var bodyMap map[string]interface{}
	switch b := body.(type) {
	case map[string]interface{}:
		bodyMap = b
	default:
		// body 不是 object，无法填充
		return body
	}

	// 填充缺失的默认值
	changed := false
	for name, prop := range schemaObj.Properties {
		if _, exists := bodyMap[name]; exists {
			continue // 已有值，跳过
		}
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			continue
		}
		if defVal, hasDefault := propMap["default"]; hasDefault {
			bodyMap[name] = defVal
			changed = true
		}
	}

	if changed {
		return bodyMap
	}
	return body
}

func (m *HTTPServiceManager) executeWithRetry(req *http.Request, retryCount int) (*http.Response, error) {
	var lastErr error

	for i := 0; i <= retryCount; i++ {
		if i > 0 {
			// 重试前等待
			backoff := time.Duration(i*100) * time.Millisecond
			time.Sleep(backoff)

			m.logger.Debug("Retrying HTTP request",
				logger.Int("attempt", i+1),
				logger.String("url", req.URL.String()),
			)
		}

		resp, err := m.httpClient.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// 如果是网络错误，继续重试
		if isNetworkError(err) && i < retryCount {
			continue
		}

		// 其他错误直接返回
		break
	}

	return nil, fmt.Errorf("HTTP request failed after %d retries: %w", retryCount+1, lastErr)
}

func isNetworkError(err error) bool {
	return strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "refused")
}

// ServiceManagerWithDAO 带持久化支持的服务管理器
type ServiceManagerWithDAO struct {
	*HTTPServiceManager
	dao interface {
		List(ctx context.Context) ([]*model.HTTPService, error)
		Save(ctx context.Context, service *model.HTTPService) error
		Delete(ctx context.Context, id uint) error
	}
}

// WithDAO 设置持久化存储（可选）
func (m *HTTPServiceManager) WithDAO(dao interface {
	List(ctx context.Context) ([]*model.HTTPService, error)
	Save(ctx context.Context, service *model.HTTPService) error
	Delete(ctx context.Context, id uint) error
}) *ServiceManagerWithDAO {
	return &ServiceManagerWithDAO{
		HTTPServiceManager: m,
		dao:                dao,
	}
}

// LoadFromDB 从数据库加载服务
func (m *ServiceManagerWithDAO) LoadFromDB(ctx context.Context) error {
	services, err := m.dao.List(ctx)
	if err != nil {
		return fmt.Errorf("load services from db failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, service := range services {
		m.services[service.ID] = service
	}

	m.logger.Info("Loaded services from database",
		logger.Int("count", len(services)),
	)

	return nil
}

// SaveToDB 保存服务到数据库
func (m *ServiceManagerWithDAO) SaveToDB(ctx context.Context, service *model.HTTPService) error {
	if err := m.dao.Save(ctx, service); err != nil {
		return fmt.Errorf("save service to db failed: %w", err)
	}

	// 同时更新内存中的服务
	m.mu.Lock()
	m.services[service.ID] = service
	m.mu.Unlock()

	return nil
}

// DeleteFromDB 从数据库删除服务
func (m *ServiceManagerWithDAO) DeleteFromDB(ctx context.Context, serviceID uint) error {
	if err := m.dao.Delete(ctx, serviceID); err != nil {
		return fmt.Errorf("delete service from db failed: %w", err)
	}

	// 同时从内存中删除
	m.mu.Lock()
	delete(m.services, serviceID)
	m.mu.Unlock()

	return nil
}

// validateJSONSchema 使用JSON Schema校验数据
func (m *HTTPServiceManager) validateJSONSchema(schema json.RawMessage, data interface{}, direction string) error {
	if len(schema) == 0 {
		return nil
	}

	// 将data序列化为JSON再与schema进行基本校验
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("%s data marshal failed: %w", direction, err)
	}

	// 基本校验：确保schema本身是合法的JSON
	var schemaObj interface{}
	if err := json.Unmarshal(schema, &schemaObj); err != nil {
		return fmt.Errorf("%s schema is invalid JSON: %w", direction, err)
	}

	// 基本校验：确保数据可以被解析
	var dataObj interface{}
	if err := json.Unmarshal(dataBytes, &dataObj); err != nil {
		return fmt.Errorf("%s data is invalid JSON: %w", direction, err)
	}

	// 检查schema中required字段是否在data中存在
	schemaMap, ok := schemaObj.(map[string]interface{})
	if !ok {
		return nil
	}

	if required, ok := schemaMap["required"].([]interface{}); ok {
		dataMap, ok := dataObj.(map[string]interface{})
		if !ok {
			return fmt.Errorf("%s: data is not an object, cannot check required fields", direction)
		}
		for _, req := range required {
			if reqStr, ok := req.(string); ok {
				if _, exists := dataMap[reqStr]; !exists {
					return fmt.Errorf("%s: missing required field '%s'", direction, reqStr)
				}
			}
		}
	}

	return nil
}
