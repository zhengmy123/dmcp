package service

import (
	"net/http"
	"strings"
	"time"

	"github.com/bytedance/sonic"
)

const mcpPathPrefix = "/mcp/"
const RootPath = "/mcp"

type mcpResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Detail  string      `json:"detail,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	mcpCodeSuccess       = 0
	mcpCodeBadRequest    = 400
	mcpCodeUnauthorized  = 401
	mcpCodeForbidden     = 403
	mcpCodeNotFound      = 404
	mcpCodeInternalError = 500
)

func mcpErrorResponse(w http.ResponseWriter, status int, code int, message string, detail string) {
	resp := mcpResponse{
		Code:    code,
		Message: message,
	}
	if detail != "" {
		resp.Detail = detail
	}
	writeMCPJSON(w, status, resp)
}

func writeMCPJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	data, err := sonic.Marshal(payload)
	if err != nil {
		http.Error(w, `{"code":500,"message":"json marshal error"}`, http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(data)
}

// NewHTTPHandler 元数据查询 handler
func NewHTTPHandler(registry *DynamicRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method != http.MethodGet:
			mcpErrorResponse(w, http.StatusMethodNotAllowed, mcpCodeBadRequest, "only GET is supported", "")
		case r.URL.Path == RootPath || r.URL.Path == RootPath+"/":
			writeMCPJSON(w, http.StatusOK, map[string]any{
				"code":    mcpCodeSuccess,
				"message": "success",
				"data": map[string]any{
					"updated_at": time.Now().UTC().Format(time.RFC3339),
					"message":    "Use /mcp/{vauth_key} to query scoped tools.",
				},
			})
		case strings.HasPrefix(r.URL.Path, mcpPathPrefix):
			handleGroupRoute(w, r, registry)
		default:
			mcpErrorResponse(w, http.StatusNotFound, mcpCodeNotFound, "unsupported path", r.URL.Path)
		}
	})
}

func handleGroupRoute(w http.ResponseWriter, r *http.Request, registry *DynamicRegistry) {
	path := strings.TrimPrefix(r.URL.Path, mcpPathPrefix)
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	switch len(parts) {
	case 1:
		vauthKey := strings.TrimSpace(parts[0])
		if vauthKey == "" {
			mcpErrorResponse(w, http.StatusBadRequest, mcpCodeBadRequest, "missing vauth_key in path", "")
			return
		}
		tools := registry.ListDefinitionsByVAuthKey(vauthKey)
		if len(tools) == 0 {
			mcpErrorResponse(w, http.StatusNotFound, mcpCodeNotFound, "mcp server not found", vauthKey)
			return
		}
		description, ok := registry.GetVAuthKeyDescription(vauthKey)
		if !ok {
			description = serverDescription(vauthKey)
		}
		writeMCPJSON(w, http.StatusOK, map[string]any{
			"code":    mcpCodeSuccess,
			"message": "success",
			"data": map[string]any{
				"updated_at":  time.Now().UTC().Format(time.RFC3339),
				"vauth_key":   vauthKey,
				"description": description,
				"tools":       tools,
			},
		})
	case 2:
		vauthKey := strings.TrimSpace(parts[0])
		toolName := strings.TrimSpace(parts[1])
		if vauthKey == "" || toolName == "" {
			mcpErrorResponse(w, http.StatusBadRequest, mcpCodeBadRequest, "path must be /mcp/{vauth_key}/{tool_name}", "")
			return
		}
		def, ok := registry.GetDefinitionByVAuthKey(vauthKey, toolName)
		if !ok {
			mcpErrorResponse(w, http.StatusNotFound, mcpCodeNotFound, "tool not found", "")
			return
		}
		description, ok := registry.GetVAuthKeyDescription(vauthKey)
		if !ok {
			description = serverDescription(vauthKey)
		}
		writeMCPJSON(w, http.StatusOK, map[string]any{
			"code":    mcpCodeSuccess,
			"message": "success",
			"data": map[string]any{
				"updated_at":  time.Now().UTC().Format(time.RFC3339),
				"vauth_key":   vauthKey,
				"description": description,
				"tool":        def,
			},
		})
	default:
		mcpErrorResponse(w, http.StatusNotFound, mcpCodeNotFound, "path must be /mcp/{vauth_key} or /mcp/{vauth_key}/{tool_name}", r.URL.Path)
	}
}

func serverDescription(vauthKey string) string {
	return "MCP Server for " + vauthKey
}
