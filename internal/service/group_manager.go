package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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
	httpServiceMgr *HTTPServiceManager

	cache *cache.TwoLevelLRU[string, http.Handler]

	mu sync.RWMutex
}

func NewMCPGroupManager(
	serverName, serverVersion string,
	authService *AuthService,
	buildInfoCache *BuildInfoCacheService,
	buildInfoStore repository.ServerBuildInfoStore,
	serverStore repository.MCPServerStore,
	httpServiceMgr *HTTPServiceManager,
	config MCPGroupManagerConfig,
) *MCPGroupManager {
	return &MCPGroupManager{
		serverName:     serverName,
		serverVersion:  serverVersion,
		authService:    authService,
		buildInfoCache: buildInfoCache,
		buildInfoStore: buildInfoStore,
		serverStore:    serverStore,
		httpServiceMgr: httpServiceMgr,
		cache:          cache.NewTwoLevelLRU[string, http.Handler](config.Cache),
	}
}

func (m *MCPGroupManager) GetHandler(vauthKey string) (http.Handler, error) {
	buildUUID, err := m.getBuildUUID(vauthKey)
	if err != nil {
		return nil, err
	}

	cached, ok := m.cache.Get(buildUUID)
	if ok {
		return cached, nil
	}

	buildInfo, err := m.loadBuildInfoByBuildUUID(context.Background(), buildUUID)
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

	m.cache.Set(buildUUID, handler)

	return handler, nil
}

func (m *MCPGroupManager) getBuildUUID(vauthKey string) (string, error) {
	if m.buildInfoCache != nil {
		cached, err := m.buildInfoCache.GetBuildUUID(context.Background(), vauthKey)
		if err == nil && cached != nil && cached.BuildUUID != "" {
			return cached.BuildUUID, nil
		}
	}

	if m.serverStore == nil {
		return "", fmt.Errorf("server store not available")
	}

	server, err := m.serverStore.GetByVAuthKey(context.Background(), vauthKey)
	if err != nil || server == nil {
		return "", ErrBuildInfoNotFound
	}

	buildInfo, err := m.buildInfoStore.GetActiveByServerID(context.Background(), server.ID)
	if err != nil || buildInfo == nil {
		return "", ErrBuildInfoNotFound
	}

	if m.buildInfoCache != nil {
		_ = m.buildInfoCache.SetBuildUUID(context.Background(), vauthKey, &CachedBuildUUID{
			BuildUUID: buildInfo.BuildUUID,
			Version:   buildInfo.Version,
		})
	}

	return buildInfo.BuildUUID, nil
}

func (m *MCPGroupManager) loadBuildInfoByBuildUUID(ctx context.Context, buildUUID string) (*model.ServerBuildInfo, error) {
	if m.buildInfoStore == nil {
		return nil, fmt.Errorf("build info store not available")
	}
	return m.buildInfoStore.GetByBuildUUID(ctx, buildUUID)
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
		if t.State != 1 {
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
	params := t.Parameters
	rawSchema, err := toRawInputSchema(params)
	if err != nil {
		return server.ServerTool{}, err
	}

	tool := mcp.NewToolWithRawSchema(t.Name, t.Description, rawSchema)
	return server.ServerTool{
		Tool:    tool,
		Handler: m.createToolHandler(t),
	}, nil
}

func (m *MCPGroupManager) createToolHandler(t model.ToolSnapshot) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if t.ServiceID == 0 {
			payload := map[string]any{
				"tool":      t.Name,
				"arguments": request.GetArguments(),
				"error":     "tool has no service_id configured",
			}
			result, err := mcp.NewToolResultJSON(payload)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err)), nil
			}
			return result, nil
		}

		if m.httpServiceMgr == nil {
			payload := map[string]any{
				"tool":      t.Name,
				"arguments": request.GetArguments(),
				"error":     "http service manager not available",
			}
			result, err := mcp.NewToolResultJSON(payload)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err)), nil
			}
			return result, nil
		}

		args := request.GetArguments()
		body := m.applyInputMapping(t.InputMapping, args)

		reqData := &model.RequestData{
			Body: body,
		}

		resp, err := m.httpServiceMgr.ExecuteService(ctx, t.ServiceID, reqData)
		if err != nil {
			payload := map[string]any{
				"tool":  t.Name,
				"error": err.Error(),
			}
			result, err := mcp.NewToolResultJSON(payload)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err)), nil
			}
			return result, nil
		}

		output := m.applyOutputMapping(t.OutputMapping, resp.Body)

		payload := map[string]any{
			"tool":       t.Name,
			"statusCode": resp.StatusCode,
		}
		if outputMap, ok := output.(map[string]any); ok {
			for k, v := range outputMap {
				payload[k] = v
			}
		}
		result, err := mcp.NewToolResultJSON(payload)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err)), nil
		}
		return result, nil
	}
}

func (m *MCPGroupManager) applyInputMapping(mapping []model.InputMappingField, args map[string]any) map[string]any {
	if len(mapping) == 0 {
		return args
	}

	result := make(map[string]any)
	for _, field := range mapping {
		if val, ok := args[field.Source]; ok {
			result[field.Target] = val
		}
	}
	if len(result) == 0 {
		return args
	}
	return result
}

func (m *MCPGroupManager) applyOutputMapping(mapping *model.OutputMappingConfig, body interface{}) interface{} {
	if mapping == nil || len(mapping.Fields) == 0 {
		return body
	}

	bodyMap, ok := body.(map[string]any)
	if !ok {
		return body
	}

	result := make(map[string]any)
	for _, field := range mapping.Fields {
		val := m.getNestedValue(bodyMap, field.SourceField)
		if val != nil {
			result[field.TargetField] = val
		} else if field.DefaultValue != "" {
			result[field.TargetField] = field.DefaultValue
		}
	}
	return result
}

func (m *MCPGroupManager) getNestedValue(body map[string]any, path string) interface{} {
	parts := strings.Split(path, ".")
	current := interface{}(body)
	for _, part := range parts {
		if m2, ok := current.(map[string]any); ok {
			if v, exists := m2[part]; exists {
				current = v
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
	return current
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
