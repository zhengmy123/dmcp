package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dynamic_mcp_go_server/internal/common/logger"

	"github.com/dop251/goja"
)

// 默认脚本执行超时时间
const DefaultScriptTimeout = 500 * time.Millisecond

// ScriptValidator 用于执行JavaScript验证脚本
type ScriptValidator struct {
	logger          logger.Logger
	scriptTimeout   time.Duration
}

// NewScriptValidator 创建新的脚本验证器
func NewScriptValidator(log logger.Logger) *ScriptValidator {
	return &ScriptValidator{
		logger:        log,
		scriptTimeout: DefaultScriptTimeout,
	}
}

// NewScriptValidatorWithTimeout 创建带自定义超时的脚本验证器
func NewScriptValidatorWithTimeout(log logger.Logger, timeout time.Duration) *ScriptValidator {
	return &ScriptValidator{
		logger:        log,
		scriptTimeout: timeout,
	}
}

// ValidationContext 验证脚本的上下文对象
type ValidationContext struct {
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
	Time    time.Time         `json:"time"`
}

// setupCommonVM 设置 VM 的公共对象（context, console, JSON, String）
func (v *ScriptValidator) setupCommonVM(vm *goja.Runtime, headers map[string]string, body interface{}, method, reqURL string) {
	now := time.Now()
	contextObj := vm.NewObject()
	contextObj.Set("headers", headers)
	contextObj.Set("body", body)
	contextObj.Set("time", now.UnixMilli()) // 毫秒时间戳
	contextObj.Set("method", method)         // 请求方法
	contextObj.Set("url", reqURL)            // 请求URL

	vm.Set("console", map[string]interface{}{
		"log": func(args ...interface{}) {
			v.logger.Debug("Script console.log", logger.Any("args", args))
		},
		"error": func(args ...interface{}) {
			v.logger.Error("Script console.error", logger.Any("args", args))
		},
	})

	vm.Set("JSON", map[string]interface{}{
		"stringify": func(args ...interface{}) (string, error) {
			if len(args) < 1 {
				return "", fmt.Errorf("JSON.stringify requires at least 1 argument")
			}
			data, err := json.Marshal(args[0])
			if err != nil {
				return "", err
			}
			return string(data), nil
		},
		"parse": func(str string) (interface{}, error) {
			var result interface{}
			err := json.Unmarshal([]byte(str), &result)
			return result, err
		},
	})

	vm.Set("String", func(obj interface{}) string {
		return fmt.Sprintf("%v", obj)
	})

	vm.Set("context", contextObj)
}

// runScriptWithTimeout 执行脚本并支持超时控制
func (v *ScriptValidator) runScriptWithTimeout(vm *goja.Runtime, script string) error {
	// 使用 channel 来处理超时
	done := make(chan error, 1)

	go func() {
		_, err := vm.RunString(script)
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(v.scriptTimeout):
		// 超时，中断脚本执行
		vm.Interrupt("script execution timeout")
		return fmt.Errorf("script execution timeout after %v", v.scriptTimeout)
	}
}

// ValidateRequest 使用JavaScript脚本验证请求
func (v *ScriptValidator) ValidateRequest(script string, headers map[string]string, body interface{}, method, reqURL string) (bool, string, error) {
	if strings.TrimSpace(script) == "" {
		return true, "", nil
	}

	vm := goja.New()
	v.setupCommonVM(vm, headers, body, method, reqURL)
	RegisterCryptoModule(vm)

	globalObj := vm.NewObject()
	vm.Set("global", globalObj)

	err := v.runScriptWithTimeout(vm, script)
	if err != nil {
		return false, fmt.Sprintf("Script execution error: %v", err), err
	}

	validValue := vm.Get("valid")
	if validValue == nil {
		return true, "", nil
	}

	valid, ok := validValue.Export().(bool)
	if !ok {
		validStr := fmt.Sprintf("%v", validValue.Export())
		valid = strings.ToLower(validStr) == "true" || validStr == "1"
	}

	message := ""
	messageValue := vm.Get("message")
	if messageValue != nil {
		if msgStr, ok := messageValue.Export().(string); ok {
			message = msgStr
		}
	}

	return valid, message, nil
}

// TransformResult 转换结果
type TransformResult struct {
	Headers map[string]string
	Body    interface{}
	URLParams map[string]string // URL查询参数
}

// TransformRequest 使用JavaScript脚本转换请求
func (v *ScriptValidator) TransformRequest(script string, headers map[string]string, body interface{}, method, reqURL string) (*TransformResult, error) {
	if strings.TrimSpace(script) == "" {
		v.logger.Debug("TransformRequest: empty script, returning original data")
		return &TransformResult{Headers: headers, Body: body, URLParams: nil}, nil
	}

	v.logger.Debug("TransformRequest: executing script",
		logger.String("method", method),
		logger.String("url", reqURL),
		logger.Any("headers", headers),
		logger.Any("body", body),
	)

	vm := goja.New()
	v.setupCommonVM(vm, headers, body, method, reqURL)
	RegisterCryptoModule(vm)

	err := v.runScriptWithTimeout(vm, script)
	if err != nil {
		v.logger.Error("TransformRequest: script execution failed",
			logger.Error(err),
			logger.String("script_preview", previewScript(script)),
		)
		return nil, fmt.Errorf("transform script execution error: %v", err)
	}

	result := &TransformResult{
		Headers:   headers,
		Body:      body,
		URLParams: nil,
	}

	headersValue := vm.Get("transformedHeaders")
	if headersValue != nil {
		if headersObj, ok := headersValue.Export().(map[string]interface{}); ok {
			result.Headers = make(map[string]string)
			for k, v := range headersObj {
				if str, ok := v.(string); ok {
					result.Headers[k] = str
				} else {
					result.Headers[k] = fmt.Sprintf("%v", v)
				}
			}
			v.logger.Debug("TransformRequest: transformed headers",
				logger.Any("headers", result.Headers),
			)
		}
	}

	bodyValue := vm.Get("transformedBody")
	if bodyValue != nil {
		result.Body = bodyValue.Export()
		v.logger.Debug("TransformRequest: transformed body",
			logger.Any("body", result.Body),
		)
	}

	// 支持 transformedURLParams 设置URL参数
	urlParamsValue := vm.Get("transformedURLParams")
	if urlParamsValue != nil {
		if urlParamsObj, ok := urlParamsValue.Export().(map[string]interface{}); ok {
			result.URLParams = make(map[string]string)
			for k, v := range urlParamsObj {
				if str, ok := v.(string); ok {
					result.URLParams[k] = str
				} else {
					result.URLParams[k] = fmt.Sprintf("%v", v)
				}
			}
			v.logger.Debug("TransformRequest: transformed URL params",
				logger.Any("url_params", result.URLParams),
			)
		}
	}

	v.logger.Debug("TransformRequest: completed successfully",
		logger.Bool("has_headers", result.Headers != nil && len(result.Headers) > 0),
		logger.Bool("has_body", result.Body != nil),
		logger.Bool("has_url_params", result.URLParams != nil && len(result.URLParams) > 0),
	)

	return result, nil
}

// previewScript 返回脚本的前缀用于日志（避免日志过长）
func previewScript(script string) string {
	const maxLen = 200
	if len(script) <= maxLen {
		return script
	}
	return script[:maxLen] + "..."
}
