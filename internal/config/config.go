package config

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	StoreMemory = "memory"
	StoreMySQL  = "mysql"
)

type Config struct {
	ServerName     string
	ServerVersion  string
	Store          string
	RefreshSeconds int
	HTTPAddr       string
	MySQLDSN       string
	MySQLTable     string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int

	// 认证配置
	AdminToken    string // 后台管理 API Token (旧版兼容)
	JWTSecret     string // JWT 密钥
	JWTExpiration int    // JWT 过期时间（小时）
}

func Load() Config {
	return Config{
		ServerName:     getenv("MCP_SERVER_NAME", "dynamic-mcp-go-server"),
		ServerVersion:  getenv("MCP_SERVER_VERSION", "0.1.0"),
		Store:          strings.ToLower(getenv("TOOL_STORE", StoreMySQL)),
		RefreshSeconds: getenvInt("REFRESH_SECONDS", 10),
		HTTPAddr:       loadHTTPAddr(),
		MySQLDSN:       getenv("MYSQL_DSN", "root:1234qwer@tcp(127.0.0.1:3306)/mcp_server?charset=utf8mb4&parseTime=True&loc=Local"),
		MySQLTable:     getenv("MYSQL_TABLE", "mcp_tool_definitions"),
		RedisAddr:      getenv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword:  getenv("REDIS_PASSWORD", ""),
		RedisDB:        getenvInt("REDIS_DB", 0),
		AdminToken:     getenv("ADMIN_TOKEN", "admin-secret-token"),
		JWTSecret:      getenv("JWT_SECRET", "mcp-server-jwt-secret-key-change-in-production"),
		JWTExpiration:  getenvInt("JWT_EXPIRATION", 168),
	}
}

func (c Config) RefreshInterval() time.Duration {
	return time.Duration(c.RefreshSeconds) * time.Second
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func loadHTTPAddr() string {
	if raw := strings.TrimSpace(os.Getenv("HTTP_ADDR")); raw != "" {
		if isDigits(raw) {
			return ":" + raw
		}
		return raw
	}

	port := strings.TrimSpace(getenv("HTTP_PORT", "18080"))
	if isDigits(port) {
		return net.JoinHostPort("0.0.0.0", port)
	}
	return "0.0.0.0:18080"
}

func isDigits(v string) bool {
	if v == "" {
		return false
	}
	for _, r := range v {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
