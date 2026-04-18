package service

import (
	"github.com/bytedance/sonic"
	"net/http"
	"strings"
	"time"
)

const mcpPathPrefix = "/mcp/"
const RootPath = "/mcp"

// NewHTTPHandler 元数据查询 handler
func NewHTTPHandler(registry *DynamicRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method != http.MethodGet:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{
				"error": "only GET is supported",
			})
		case r.URL.Path == RootPath || r.URL.Path == RootPath+"/":
			writeJSON(w, http.StatusOK, map[string]any{
				"updated_at": time.Now().UTC().Format(time.RFC3339),
				"message":    "Use /mcp/{vauth_key} to query scoped tools.",
			})
		case strings.HasPrefix(r.URL.Path, mcpPathPrefix):
			handleGroupRoute(w, r, registry)
		default:
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error": "unsupported path",
				"path":  r.URL.Path,
			})
		}
	})
}

func writeJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	data, err := sonic.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = w.Write(data)
}

func handleGroupRoute(w http.ResponseWriter, r *http.Request, registry *DynamicRegistry) {
	path := strings.TrimPrefix(r.URL.Path, mcpPathPrefix)
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	switch len(parts) {
	case 1:
		vauthKey := strings.TrimSpace(parts[0])
		if vauthKey == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error": "missing vauth_key in path",
			})
			return
		}
		tools := registry.ListDefinitionsByVAuthKey(vauthKey)
		if len(tools) == 0 {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error":     "mcp server not found",
				"vauth_key": vauthKey,
			})
			return
		}
		description, ok := registry.GetVAuthKeyDescription(vauthKey)
		if !ok {
			description = serverDescription(vauthKey)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"updated_at":  time.Now().UTC().Format(time.RFC3339),
			"vauth_key":   vauthKey,
			"description": description,
			"tools":       tools,
		})
	case 2:
		vauthKey := strings.TrimSpace(parts[0])
		toolName := strings.TrimSpace(parts[1])
		if vauthKey == "" || toolName == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error": "path must be /mcp/{vauth_key}/{tool_name}",
			})
			return
		}
		def, ok := registry.GetDefinitionByVAuthKey(vauthKey, toolName)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"error":     "tool not found",
				"vauth_key": vauthKey,
				"name":      toolName,
			})
			return
		}
		description, ok := registry.GetVAuthKeyDescription(vauthKey)
		if !ok {
			description = serverDescription(vauthKey)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"updated_at":  time.Now().UTC().Format(time.RFC3339),
			"vauth_key":   vauthKey,
			"description": description,
			"tool":        def,
		})
	default:
		writeJSON(w, http.StatusNotFound, map[string]any{
			"error": "path must be /mcp/{vauth_key} or /mcp/{vauth_key}/{tool_name}",
			"path":  r.URL.Path,
		})
	}
}

func serverDescription(vauthKey string) string {
	return "MCP Server for " + vauthKey
}
