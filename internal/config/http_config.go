package config

import (
	"os"
	"strconv"
	"strings"
)

// HTTPServiceConfig 定义HTTP服务对接配置
type HTTPServiceConfig struct {
	Enabled                   bool
	ServiceID                 string
	ServiceName               string
	TargetURL                 string
	Method                    string
	Headers                   map[string]string
	TimeoutSeconds            int
	RetryCount                int
	ValidationScript          string
	ValidationEnabled         bool
	RequestTransformScript    string
	ResponseTransformScript   string
	InputSchema               string
	OutputSchema              string
}

// HTTPServiceStoreConfig 定义HTTP服务存储配置
type HTTPServiceStoreConfig struct {
	StoreType    string
	MySQLDSN     string
	MySQLTable   string
	SyncEnabled  bool
	SyncInterval int
}

// LoadHTTPServiceConfig 从环境变量加载HTTP服务配置
func LoadHTTPServiceConfig() HTTPServiceConfig {
	return HTTPServiceConfig{
		Enabled:                 getenvBool("HTTP_SERVICE_ENABLED", false),
		ServiceID:               getenv("HTTP_SERVICE_ID", ""),
		ServiceName:             getenv("HTTP_SERVICE_NAME", ""),
		TargetURL:               getenv("HTTP_TARGET_URL", ""),
		Method:                  getenv("HTTP_METHOD", "POST"),
		Headers:                 parseHeaders(getenv("HTTP_HEADERS", "")),
		TimeoutSeconds:          getenvInt("HTTP_TIMEOUT_SECONDS", 30),
		RetryCount:              getenvInt("HTTP_RETRY_COUNT", 3),
		ValidationScript:        getenv("HTTP_VALIDATION_SCRIPT", ""),
		ValidationEnabled:       getenvBool("HTTP_VALIDATION_ENABLED", false),
		RequestTransformScript:  getenv("HTTP_REQUEST_TRANSFORM_SCRIPT", ""),
		ResponseTransformScript: getenv("HTTP_RESPONSE_TRANSFORM_SCRIPT", ""),
		InputSchema:             getenv("HTTP_INPUT_SCHEMA", ""),
		OutputSchema:            getenv("HTTP_OUTPUT_SCHEMA", ""),
	}
}

// LoadHTTPServiceStoreConfig 从环境变量加载HTTP服务存储配置
func LoadHTTPServiceStoreConfig() HTTPServiceStoreConfig {
	return HTTPServiceStoreConfig{
		StoreType:    strings.ToLower(getenv("HTTP_SERVICE_STORE", "mysql")),
		MySQLDSN:     getenv("HTTP_MYSQL_DSN", "root:1234qwer@tcp(127.0.0.1:3306)/mcp_server?charset=utf8mb4&parseTime=True&loc=Local"),
		MySQLTable:   getenv("HTTP_MYSQL_TABLE", "mcp_http_services"),
		SyncEnabled:  getenvBool("HTTP_SYNC_ENABLED", true),
		SyncInterval: getenvInt("HTTP_SYNC_INTERVAL", 60),
	}
}

func parseHeaders(headersStr string) map[string]string {
	headers := make(map[string]string)
	if headersStr == "" {
		return headers
	}
	
	pairs := strings.Split(headersStr, ";")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			if key != "" && value != "" {
				headers[key] = value
			}
		}
	}
	return headers
}

func getenvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}