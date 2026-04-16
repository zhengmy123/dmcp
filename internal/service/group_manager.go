package service

import (
	"net/http"
	"sync"

	"github.com/mark3labs/mcp-go/server"
)

// MCPGroupManager 按 vauthKey 聚合 tools，为每个 vauthKey 构建独立的 MCP Server
type MCPGroupManager struct {
	serverName    string
	serverVersion string
	authService   *AuthService

	mu       sync.RWMutex
	handlers map[string]http.Handler // vauthKey -> StreamableHTTPServer
}

// NewMCPGroupManager 创建 MCP 分组管理器
func NewMCPGroupManager(serverName, serverVersion string, authService *AuthService) *MCPGroupManager {
	return &MCPGroupManager{
		serverName:    serverName,
		serverVersion: serverVersion,
		authService:   authService,
		handlers:      make(map[string]http.Handler),
	}
}

// Rebuild 重建所有 vauthKey 的 MCP handler
func (m *MCPGroupManager) Rebuild(toolsByGroup map[string][]server.ServerTool) {
	newHandlers := make(map[string]http.Handler, len(toolsByGroup))
	for vauthKey, groupTools := range toolsByGroup {
		if len(groupTools) == 0 {
			continue
		}
		groupMCP := server.NewMCPServer(
			m.serverName+"::"+vauthKey,
			m.serverVersion,
			server.WithToolCapabilities(true),
			server.WithRecovery(),
		)
		groupMCP.SetTools(groupTools...)
		newHandlers[vauthKey] = server.NewStreamableHTTPServer(
			groupMCP,
			server.WithStateLess(true),
		)
	}

	m.mu.Lock()
	m.handlers = newHandlers
	m.mu.Unlock()
}

// PatchHandlers 增量更新指定 vauthKey 的 handler
func (m *MCPGroupManager) PatchHandlers(updates map[string][]server.ServerTool, remove []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.handlers == nil {
		m.handlers = make(map[string]http.Handler)
	}
	for _, key := range remove {
		delete(m.handlers, key)
	}
	for vauthKey, groupTools := range updates {
		if len(groupTools) == 0 {
			delete(m.handlers, vauthKey)
			continue
		}
		groupMCP := server.NewMCPServer(
			m.serverName+"::"+vauthKey,
			m.serverVersion,
			server.WithToolCapabilities(true),
			server.WithRecovery(),
		)
		groupMCP.SetTools(groupTools...)
		m.handlers[vauthKey] = server.NewStreamableHTTPServer(
			groupMCP,
			server.WithStateLess(true),
		)
	}
}

// Handler 获取指定 vauthKey 的 MCP handler
func (m *MCPGroupManager) Handler(vauthKey string) (http.Handler, bool) {
	key := CanonicalVAuthKey(vauthKey)
	m.mu.RLock()
	defer m.mu.RUnlock()
	h, ok := m.handlers[key]
	return h, ok
}

// ListGroups 列出所有已注册的 vauthKey
func (m *MCPGroupManager) ListGroups() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]string, 0, len(m.handlers))
	for k := range m.handlers {
		keys = append(keys, k)
	}
	return keys
}