package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"dynamic_mcp_go_server/internal/common/cache"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"

	"github.com/bytedance/sonic"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var ErrBuildInfoNotFound = fmt.Errorf("build info not found")

type MCPGroupManagerConfig struct {
	Cache cache.Config
	Redis time.Duration
}

type MCPGroupManager struct {
	serverName    string
	serverVersion string
	authService   *AuthService

	buildInfoCache *BuildInfoCacheService
	buildInfoStore repository.ServerBuildInfoStore
	serverStore    repository.MCPServerStore

	cache *cache.TwoLevelLRU[string, http.Handler]

	mu sync.RWMutex
}

func NewMCPGroupManager(
	serverName, serverVersion string,
	authService *AuthService,
	buildInfoCache *BuildInfoCacheService,
	buildInfoStore repository.ServerBuildInfoStore,
	serverStore repository.MCPServerStore,
	config MCPGroupManagerConfig,
) *MCPGroupManager {
	return &MCPGroupManager{
		serverName:     serverName,
		serverVersion:  serverVersion,
		authService:    authService,
		buildInfoCache: buildInfoCache,
		buildInfoStore: buildInfoStore,
		serverStore:    serverStore,
		cache:          cache.NewTwoLevelLRU[string, http.Handler](config.Cache),
	}
}

func (m *MCPGroupManager) GetHandler(vauthKey string) (http.Handler, error) {
	cached, ok := m.cache.Get(vauthKey)
	if ok {
		return cached, nil
	}

	buildInfo, err := m.loadBuildInfo(context.Background(), vauthKey)
	if err != nil {
		return nil, err
	}
	if buildInfo == nil {
		return nil, ErrBuildInfoNotFound
	}

	handler, err := m.buildHandler(buildInfo)
	if err != nil {
		return nil, err
	}

	m.cache.Set(vauthKey, handler)

	return handler, nil
}

func (m *MCPGroupManager) loadBuildInfo(ctx context.Context, vauthKey string) (*model.ServerBuildInfo, error) {
	if m.buildInfoCache != nil {
		cached, err := m.buildInfoCache.GetBuildUUID(ctx, vauthKey)
		if err == nil && cached != nil {
			if m.buildInfoStore != nil {
				return m.buildInfoStore.GetByBuildUUID(ctx, cached.BuildUUID)
			}
		}
	}

	if m.serverStore == nil || m.buildInfoStore == nil {
		return nil, nil
	}

	server, err := m.serverStore.GetByVAuthKey(ctx, vauthKey)
	if err != nil {
		return nil, nil
	}

	buildInfo, err := m.buildInfoStore.GetActiveByServerID(ctx, server.ID)
	if err != nil {
		return nil, nil
	}

	if buildInfo != nil && m.buildInfoCache != nil {
		_ = m.buildInfoCache.SetBuildUUID(ctx, vauthKey, &CachedBuildUUID{
			BuildUUID: buildInfo.BuildUUID,
			Version:   buildInfo.Version,
		})
	}

	return buildInfo, nil
}

func (m *MCPGroupManager) buildHandler(info *model.ServerBuildInfo) (http.Handler, error) {
	if info == nil || info.BuildData == "" {
		return nil, ErrBuildInfoNotFound
	}

	var buildData model.BuildData
	if err := sonic.Unmarshal([]byte(info.BuildData), &buildData); err != nil {
		return nil, fmt.Errorf("unmarshal build_data: %w", err)
	}

	var tools []server.ServerTool
	for _, t := range buildData.Tools {
		if !t.Enabled {
			continue
		}
		tool, err := m.convertToServerTool(t)
		if err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}

	groupMCP := server.NewMCPServer(
		m.serverName+"::"+info.BuildUUID[:8],
		m.serverVersion,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)
	groupMCP.SetTools(tools...)

	return server.NewStreamableHTTPServer(
		groupMCP,
		server.WithStateLess(true),
	), nil
}

func (m *MCPGroupManager) convertToServerTool(t model.ToolSnapshot) (server.ServerTool, error) {
	params, err := parseToolParams(t.Parameters)
	if err != nil {
		return server.ServerTool{}, fmt.Errorf("parse parameters for tool %q: %w", t.Name, err)
	}
	rawSchema, err := toRawInputSchema(params)
	if err != nil {
		return server.ServerTool{}, err
	}

	tool := mcp.NewToolWithRawSchema(t.Name, t.Description, rawSchema)
	return server.ServerTool{
		Tool:    tool,
		Handler: mcpDefaultHandler(t.Name),
	}, nil
}

func (m *MCPGroupManager) Handler(vauthKey string) (http.Handler, bool) {
	h, err := m.GetHandler(vauthKey)
	if err != nil {
		return nil, false
	}
	return h, true
}

func (m *MCPGroupManager) ListGroups() []string {
	return nil
}

func mcpDefaultHandler(toolName string) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		payload := map[string]any{
			"tool":      toolName,
			"arguments": request.GetArguments(),
			"note":      "Replace defaultHandler with business logic.",
		}
		result, err := mcp.NewToolResultJSON(payload)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err)), nil
		}
		return result, nil
	}
}
