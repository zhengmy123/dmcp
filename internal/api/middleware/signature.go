package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// TokenAuthMiddleware MCP Token认证中间件（net/http 风格）
func TokenAuthMiddleware(tokenValidator func(token string) (valid bool, expired bool)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Mcp-Token")
			if token == "" {
				token = r.Header.Get("X-Token")
			}
			if token == "" {
				token = r.URL.Query().Get("token")
			}
			if token == "" {
				writeAuthError(w, "missing X-Mcp-Token header")
				return
			}

			valid, expired := tokenValidator(token)
			if !valid {
				if expired {
					writeAuthError(w, "token expired")
				} else {
					writeAuthError(w, "invalid token")
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminAuthMiddleware 后台管理认证中间件（net/http 风格）
func AdminAuthMiddleware(adminToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if adminToken == "" {
				next.ServeHTTP(w, r)
				return
			}
			token := r.Header.Get("X-Admin-Token")
			if token == "" || token != adminToken {
				writeAuthError(w, "invalid or missing admin token")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// VerifyHMACSignature 验证HMAC签名
func VerifyHMACSignature(secret, method, path, timestamp string) string {
	signStr := buildSignString(method, path, timestamp)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}

func buildSignString(method, path, timestamp string) string {
	return strings.ToUpper(method) + "\n" + path + "\n" + timestamp
}

func writeAuthError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, msg)))
}

// VerifySignatureWithTimestamp 验证带时间戳的签名
func VerifySignatureWithTimestamp(secret, signature, method, path, timestamp string) bool {
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		var unixTs int64
		if _, err := fmt.Sscanf(timestamp, "%d", &unixTs); err != nil {
			return false
		}
		ts = time.Unix(unixTs, 0)
	}

	now := time.Now()
	if ts.Before(now.Add(-5*time.Minute)) || ts.After(now.Add(5*time.Minute)) {
		return false
	}

	expectedSig := VerifyHMACSignature(secret, method, path, timestamp)
	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// SortedHeaderKeys 返回排序后的 header 键
func SortedHeaderKeys(headers map[string]string) []string {
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
