package service

import (
	"net/http"
	"strings"
)

// NewScopedMCPHandler 创建按 vauthKey 分发的 MCP handler
// URL: /mcp/{vauth_key}
func NewScopedMCPHandler(manager *MCPGroupManager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		groupKey := strings.Trim(strings.TrimPrefix(r.URL.Path, mcpPathPrefix), "/")
		if groupKey == "" || strings.Contains(groupKey, "/") {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error": "path must be /mcp/{vauth_key}",
				"path":  r.URL.Path,
			})
			return
		}

		h, ok := manager.Handler(groupKey)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error":     "mcp server not found",
				"vauth_key": groupKey,
			})
			return
		}

		h.ServeHTTP(w, r)
	})
}

// CanonicalVAuthKey 规范化 vauthKey
func CanonicalVAuthKey(vauthKey string) string {
	return strings.TrimSpace(vauthKey)
}
