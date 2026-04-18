package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
	"time"

	"github.com/bytedance/sonic"
)

// HTTPService 定义HTTP服务配置
type HTTPService struct {
	ID                      uint              `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                    string            `json:"name" gorm:"size:128;not null"`
	Description             string            `json:"description,omitempty" gorm:"size:512"`
	TargetURL               string            `json:"target_url" gorm:"size:512;not null"`
	Method                  string            `json:"method" gorm:"size:16;not null;default:POST"`
	BodyType                 string            `json:"body_type" gorm:"size:32;default:JSON"` // 请求体类型: none, form-data, urlencoded, binary, msgpack, raw, JSON
	Headers                 map[string]string `json:"headers,omitempty" gorm:"-"`
	HeadersJSON             string            `json:"-" gorm="column:headers;type:text"`
	TimeoutSeconds          int               `json:"timeout_seconds" gorm:"default:30"`
	RetryCount              int               `json:"retry_count" gorm:"default:0"`
	ValidationScript        string            `json:"validation_script,omitempty" gorm:"type:text"`
	ValidationEnabled       bool              `json:"validation_enabled" gorm:"default:false"`
	RequestTransformScript  string            `json:"request_transform_script,omitempty" gorm:"type:text"`              // 请求转换脚本
	ResponseTransformScript string            `json:"response_transform_script,omitempty" gorm:"type:text"`             // 响应转换脚本
	InputSchema             JSONBytes         `json:"input_schema,omitempty" gorm:"type:text"`                          // 入参JSON Schema
	OutputSchema            JSONBytes         `json:"output_schema,omitempty" gorm:"type:text"`                         // 出参JSON Schema
	Enabled                 bool              `json:"enabled" gorm:"default:true;index"`
	CreatedAt               time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt               time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

// JSONBytes 自定义类型，用于将 []byte 直接序列化为 JSON 对象而不是 base64
type JSONBytes []byte

func (j JSONBytes) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	if !json.Valid(j) {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSONBytes) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` || string(data) == "" {
		*j = nil
		return nil
	}
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON: %s", string(data))
	}
	*j = data
	return nil
}

func (HTTPService) TableName() string {
	return "mcp_http_services"
}

// RequestData 定义HTTP请求数据
type RequestData struct {
	Headers  map[string]string `json:"headers,omitempty"`
	Body     interface{}       `json:"body,omitempty"`
	Query    map[string]string `json:"query,omitempty"`
}

// ResponseData 定义HTTP响应数据
type ResponseData struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       interface{}       `json:"body,omitempty"`
	Duration   time.Duration     `json:"duration_ms"`
	Error      string            `json:"error,omitempty"`
}

// DefaultTimeoutSeconds 默认超时秒数
const DefaultTimeoutSeconds = 30

// DefaultRetryCount 默认重试次数
const DefaultRetryCount = 0

// DefaultMethod 默认HTTP方法
const DefaultMethod = "POST"

// DefaultBodyType 默认请求体类型
const DefaultBodyType = "JSON"

// BodyType 常量定义
const (
	BodyTypeNone       = "none"
	BodyTypeFormData   = "form-data"
	BodyTypeURLEncoded = "urlencoded"
	BodyTypeBinary     = "binary"
	BodyTypeMsgpack    = "msgpack"
	BodyTypeRaw        = "raw"
	BodyTypeJSON       = "JSON"
)

// GetDefaultContentType 根据 body type 返回默认的 Content-Type
func (s *HTTPService) GetDefaultContentType() string {
	bodyType := s.BodyType
	if bodyType == "" {
		bodyType = DefaultBodyType
	}
	switch bodyType {
	case BodyTypeNone:
		return ""
	case BodyTypeFormData:
		return "multipart/form-data"
	case BodyTypeURLEncoded:
		return "application/x-www-form-urlencoded"
	case BodyTypeBinary:
		return "application/octet-stream"
	case BodyTypeMsgpack:
		return "application/msgpack"
	case BodyTypeRaw:
		return "text/plain"
	case BodyTypeJSON:
		fallthrough
	default:
		return "application/json"
	}
}

// EncodeBodyResult 是 EncodeBody 的返回结果
type EncodeBodyResult struct {
	Reader         io.Reader // 编码后的请求体
	OverrideContentType string // 需要覆盖的 Content-Type（如 multipart boundary），空表示不覆盖
}

// EncodeBody 根据指定的 bodyType 将 body 编码为 io.Reader
// bodyType 为空时使用 DefaultBodyType
func EncodeBody(bodyType string, body interface{}) (*EncodeBodyResult, error) {
	if body == nil {
		return &EncodeBodyResult{Reader: nil}, nil
	}

	if bodyType == "" {
		bodyType = DefaultBodyType
	}

	switch bodyType {
	case BodyTypeFormData:
		return encodeFormData(body)
	case BodyTypeURLEncoded:
		return encodeURLEncoded(body)
	case BodyTypeNone:
		return &EncodeBodyResult{Reader: nil}, nil
	case BodyTypeRaw, BodyTypeBinary, BodyTypeMsgpack:
		return encodeRawBody(body), nil
	case BodyTypeJSON:
		fallthrough
	default:
		return encodeJSONBody(body)
	}
}

// encodeFormData 将 body 编码为 multipart/form-data
func encodeFormData(body interface{}) (*EncodeBodyResult, error) {
	formData, ok := body.(map[string]interface{})
	if !ok {
		// 降级为 JSON
		bodyBytes, _ := sonic.Marshal(body)
		return &EncodeBodyResult{Reader: strings.NewReader(string(bodyBytes))}, nil
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for k, v := range formData {
		if err := writer.WriteField(k, fmt.Sprintf("%v", v)); err != nil {
			return nil, fmt.Errorf("write form field %s failed: %w", k, err)
		}
	}
	writer.Close()

	return &EncodeBodyResult{
		Reader:              &buf,
		OverrideContentType: writer.FormDataContentType(),
	}, nil
}

// encodeURLEncoded 将 body 编码为 application/x-www-form-urlencoded
func encodeURLEncoded(body interface{}) (*EncodeBodyResult, error) {
	formData, ok := body.(map[string]interface{})
	if !ok {
		bodyBytes, _ := sonic.Marshal(body)
		return &EncodeBodyResult{Reader: strings.NewReader(string(bodyBytes))}, nil
	}

	values := url.Values{}
	for k, v := range formData {
		values.Set(k, fmt.Sprintf("%v", v))
	}
	return &EncodeBodyResult{Reader: strings.NewReader(values.Encode())}, nil
}

// encodeRawBody 将 body 透传为字符串
func encodeRawBody(body interface{}) *EncodeBodyResult {
	if str, ok := body.(string); ok {
		return &EncodeBodyResult{Reader: strings.NewReader(str)}
	}
	bodyBytes, _ := sonic.Marshal(body)
	return &EncodeBodyResult{Reader: strings.NewReader(string(bodyBytes))}
}

// encodeJSONBody 将 body 编码为 JSON
func encodeJSONBody(body interface{}) (*EncodeBodyResult, error) {
	bodyBytes, err := sonic.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %w", err)
	}
	return &EncodeBodyResult{Reader: strings.NewReader(string(bodyBytes))}, nil
}

// NewHTTPService 创建新的HTTP服务配置
func NewHTTPService(name, targetURL, method string) *HTTPService {
	now := time.Now()
	return &HTTPService{
		Name:           name,
		TargetURL:      targetURL,
		Method:         method,
		BodyType:       DefaultBodyType,
		Headers:        make(map[string]string),
		TimeoutSeconds: DefaultTimeoutSeconds,
		RetryCount:     DefaultRetryCount,
		InputSchema:    []byte(`{"type":"object","properties":{}}`),
		OutputSchema:   []byte(`{"type":"object","properties":{}}`),
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// GetID 获取ID的字符串表示
func (s *HTTPService) GetID() string {
	return fmt.Sprintf("%d", s.ID)
}

// HasInputSchema 检查是否有入参Schema定义
func (s *HTTPService) HasInputSchema() bool {
	return len(s.InputSchema) > 0 && string(s.InputSchema) != `{"type":"object","properties":{}}`
}

// HasOutputSchema 检查是否有出参Schema定义
func (s *HTTPService) HasOutputSchema() bool {
	return len(s.OutputSchema) > 0 && string(s.OutputSchema) != `{"type":"object","properties":{}}`
}

// HasRequestTransform 检查是否有请求转换脚本
func (s *HTTPService) HasRequestTransform() bool {
	return s.RequestTransformScript != ""
}

// HasResponseTransform 检查是否有响应转换脚本
func (s *HTTPService) HasResponseTransform() bool {
	return s.ResponseTransformScript != ""
}

// IsValid 检查服务配置是否有效
func (s *HTTPService) IsValid() bool {
	return s.ID > 0 && s.Name != "" && s.TargetURL != ""
}
