package service_test

import (
	"context"
	"testing"
	"time"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/service"
)

// mockLogger 是用于测试的日志实现
type mockLogger struct {
	debugs []string
	infos  []string
	errors []string
}

func (m *mockLogger) Debug(msg string, fields ...logger.Field) {
	m.debugs = append(m.debugs, msg)
}

func (m *mockLogger) Debugf(format string, args ...interface{}) {}

func (m *mockLogger) Info(msg string, fields ...logger.Field) {
	m.infos = append(m.infos, msg)
}

func (m *mockLogger) Infof(format string, args ...interface{}) {}

func (m *mockLogger) Error(msg string, fields ...logger.Field) {
	m.errors = append(m.errors, msg)
}

func (m *mockLogger) Errorf(format string, args ...interface{}) {}

func (m *mockLogger) Warn(msg string, fields ...logger.Field) {}

func (m *mockLogger) Warnf(format string, args ...interface{}) {}

func (m *mockLogger) Fatal(msg string, fields ...logger.Field) {}

func (m *mockLogger) With(fields ...logger.Field) logger.Logger {
	return m
}

func (m *mockLogger) Ctx(ctx context.Context) logger.Logger {
	return m
}

func (m *mockLogger) Sync() error {
	return nil
}

func TestTransformRequest_EmptyScript(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	result, err := validator.TransformRequest("", nil, nil, "POST", "http://example.com")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Headers != nil {
		t.Errorf("expected nil headers, got %v", result.Headers)
	}
	if result.Body != nil {
		t.Errorf("expected nil body, got %v", result.Body)
	}
	if result.URLParams != nil {
		t.Errorf("expected nil URLParams, got %v", result.URLParams)
	}
}

func TestTransformRequest_WhitespaceOnlyScript(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	result, err := validator.TransformRequest("   \n\t  ", nil, nil, "POST", "http://example.com")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.URLParams != nil {
		t.Errorf("expected nil URLParams, got %v", result.URLParams)
	}
}

func TestTransformRequest_TransformHeaders(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	script := `
transformedHeaders = {
  'X-Custom-Header': 'custom-value',
  'X-Request-Id': '12345',
  'Content-Type': 'application/json'  // 需要显式保留原始值
};
`
	headers := map[string]string{"Content-Type": "application/json"}
	result, err := validator.TransformRequest(script, headers, nil, "POST", "http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// 验证 transformedHeaders 被设置
	if result.Headers == nil {
		t.Fatal("expected headers to be set")
	}
	if result.Headers["X-Custom-Header"] != "custom-value" {
		t.Errorf("expected X-Custom-Header to be 'custom-value', got %s", result.Headers["X-Custom-Header"])
	}
	if result.Headers["X-Request-Id"] != "12345" {
		t.Errorf("expected X-Request-Id to be '12345', got %s", result.Headers["X-Request-Id"])
	}
	// 显式设置的值
	if result.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type to be 'application/json', got %s", result.Headers["Content-Type"])
	}
}

func TestTransformRequest_TransformBody(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	script := `
transformedBody = {
  ...context.body,
  sign: 'abc123',
  timestamp: 1234567890
};
`
	body := map[string]interface{}{"name": "test", "value": 100}
	result, err := validator.TransformRequest(script, nil, body, "POST", "http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.Body == nil {
		t.Fatal("expected body to be set")
	}
	bodyMap, ok := result.Body.(map[string]interface{})
	if !ok {
		t.Fatalf("expected body to be map, got %T", result.Body)
	}
	if bodyMap["sign"] != "abc123" {
		t.Errorf("expected sign to be 'abc123', got %v", bodyMap["sign"])
	}
	if bodyMap["timestamp"] != int64(1234567890) {
		t.Errorf("expected timestamp to be 1234567890, got %v", bodyMap["timestamp"])
	}
	if bodyMap["name"] != "test" {
		t.Errorf("expected original name to be preserved, got %v", bodyMap["name"])
	}
}

func TestTransformRequest_TransformURLParams(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	script := `
var timestamp = String(context.time);
var bodyStr = JSON.stringify(context.body || {});
var signature = hmacSha256('secret-key', bodyStr + timestamp);
transformedURLParams = {
  app_id: 'your-app-id',
  sign: signature,
  timestamp: timestamp
};
`
	body := map[string]interface{}{"key": "value"}
	result, err := validator.TransformRequest(script, nil, body, "POST", "http://example.com/api")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.URLParams == nil {
		t.Fatal("expected URLParams to be set")
	}
	if result.URLParams["app_id"] != "your-app-id" {
		t.Errorf("expected app_id to be 'your-app-id', got %s", result.URLParams["app_id"])
	}
	if result.URLParams["sign"] == "" {
		t.Error("expected sign to be non-empty")
	}
	if result.URLParams["timestamp"] == "" {
		t.Error("expected timestamp to be non-empty")
	}
}

func TestTransformRequest_NoTransform(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	// 脚本不设置任何 transformed* 变量
	script := `
// 只是打印一些日志
console.log('do nothing');
`
	headers := map[string]string{"X-Original": "header"}
	body := map[string]interface{}{"original": "body"}
	result, err := validator.TransformRequest(script, headers, body, "POST", "http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// 应该保留原始值
	if result.Headers["X-Original"] != "header" {
		t.Errorf("expected headers to be preserved, got %v", result.Headers)
	}
	if result.Body.(map[string]interface{})["original"] != "body" {
		t.Errorf("expected body to be preserved, got %v", result.Body)
	}
	if result.URLParams != nil {
		t.Errorf("expected URLParams to be nil, got %v", result.URLParams)
	}
}

func TestTransformRequest_CombinedTransform(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	script := `
var timestamp = String(context.time);
transformedHeaders = {
  'X-Timestamp': timestamp,
  'Authorization': 'Bearer token123'
};
transformedBody = {
  ...context.body,
  signed: true
};
transformedURLParams = {
  v: '1',
  ts: timestamp
};
`
	body := map[string]interface{}{"data": "test"}
	headers := map[string]string{"Content-Type": "application/json"}
	result, err := validator.TransformRequest(script, headers, body, "POST", "http://example.com/api")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// 验证所有转换
	if result.Headers["X-Timestamp"] == "" {
		t.Error("expected X-Timestamp to be set")
	}
	if result.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("expected Authorization to be 'Bearer token123', got %s", result.Headers["Authorization"])
	}

	bodyMap := result.Body.(map[string]interface{})
	if bodyMap["signed"] != true {
		t.Errorf("expected signed to be true, got %v", bodyMap["signed"])
	}

	if result.URLParams["v"] != "1" {
		t.Errorf("expected v to be '1', got %s", result.URLParams["v"])
	}
}

func TestTransformRequest_ScriptTimeout(t *testing.T) {
	log := &mockLogger{}
	// 创建超时很短的验证器
	validator := service.NewScriptValidatorWithTimeout(log, 1*time.Millisecond)

	// 死循环脚本
	script := `
while(true) {
  // infinite loop
}
`
	_, err := validator.TransformRequest(script, nil, nil, "POST", "http://example.com")
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestTransformRequest_CryptoFunctions(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	script := `
var bodyStr = JSON.stringify(context.body || {});
var signature = hmacSha256('secret-key', bodyStr);
var md5hash = md5(bodyStr + 'salt');
transformedBody = {
  ...context.body,
  sign: signature,
  md5: md5hash
};
`
	body := map[string]interface{}{"test": "data"}
	result, err := validator.TransformRequest(script, nil, body, "POST", "http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	bodyMap := result.Body.(map[string]interface{})
	if bodyMap["sign"] == "" {
		t.Error("expected sign to be non-empty")
	}
	if bodyMap["md5"] == "" {
		t.Error("expected md5 to be non-empty")
	}
}

func TestTransformRequest_HexEncodeBase64Encode(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	script := `
var data = 'hello world';
var hex = hexEncode(data);
var b64 = base64Encode(data);
transformedBody = {
  original: data,
  hex: hex,
  base64: b64
};
`
	result, err := validator.TransformRequest(script, nil, nil, "POST", "http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bodyMap := result.Body.(map[string]interface{})
	if bodyMap["hex"] != "68656c6c6f20776f726c64" { // hello world in hex
		t.Errorf("expected hex to be '68656c6c6f20776f726c64', got %v", bodyMap["hex"])
	}
	if bodyMap["base64"] != "aGVsbG8gd29ybGQ=" { // hello world in base64
		t.Errorf("expected base64 to be 'aGVsbG8gd29ybGQ=', got %v", bodyMap["base64"])
	}
}

func TestTransformRequest_UrlEncode(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	script := `
transformedURLParams = {
  query: urlEncode('hello world & 123'),
  raw: 'hello world & 123'
};
`
	result, err := validator.TransformRequest(script, nil, nil, "POST", "http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.URLParams["query"] != "hello+world+%26+123" {
		t.Errorf("expected urlEncoded query, got %s", result.URLParams["query"])
	}
}

func TestTransformRequest_SortByKeys(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	script := `
var obj = { z: 1, a: 2, m: 3 };
var sorted = sortByKeys(obj);
transformedBody = {
  original_z: obj.z,
  original_a: obj.a,
  original_m: obj.m,
  sorted_keys: Object.keys(sorted)
};
`
	result, err := validator.TransformRequest(script, nil, nil, "POST", "http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bodyMap := result.Body.(map[string]interface{})
	keysRaw := bodyMap["sorted_keys"]
	keys, ok := keysRaw.([]interface{})
	if !ok {
		t.Fatalf("expected sorted_keys to be array, got %T", keysRaw)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	// Object.keys returns keys in insertion order (which is sorted order for sortByKeys result)
	if keys[0] != "a" || keys[1] != "m" || keys[2] != "z" {
		t.Errorf("expected sorted keys [a, m, z], got %v", keys)
	}
}

func TestTransformRequest_ContextVariables(t *testing.T) {
	log := &mockLogger{}
	validator := service.NewScriptValidator(log)

	script := `
transformedBody = {
  method: context.method,
  url: context.url,
  hasHeaders: !!context.headers,
  hasBody: !!context.body,
  timePositive: context.time > 0
};
`
	body := map[string]interface{}{"test": true}
	headers := map[string]string{"X-Test": "1"}
	result, err := validator.TransformRequest(script, headers, body, "GET", "http://example.com/api/test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	bodyMap := result.Body.(map[string]interface{})
	if bodyMap["method"] != "GET" {
		t.Errorf("expected method to be 'GET', got %v", bodyMap["method"])
	}
	if bodyMap["url"] != "http://example.com/api/test" {
		t.Errorf("expected url to be 'http://example.com/api/test', got %v", bodyMap["url"])
	}
	if bodyMap["hasHeaders"] != true {
		t.Error("expected hasHeaders to be true")
	}
	if bodyMap["hasBody"] != true {
		t.Error("expected hasBody to be true")
	}
	if bodyMap["timePositive"] != true {
		t.Error("expected timePositive to be true")
	}
}


