package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/infrastructure/store/tooldef"

	"github.com/bytedance/sonic"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

const toolsListChangedMethod = "notifications/tools/list_changed"

type DynamicRegistry struct {
	server        *server.MCPServer
	store         tooldef.Store
	interval      time.Duration
	logger        logger.Logger
	groupMCP      *MCPGroupManager
	serverName    string
	serverVersion string
	lastHash      string
	mu            sync.RWMutex
	lastDefs      []tooldef.ToolDefinition
	serverStore   repository.MCPServerStore
	buildSvc      *ServerBuildService
}

func NewDynamicRegistry(s *server.MCPServer, store tooldef.Store, interval time.Duration, log logger.Logger, groupMCP *MCPGroupManager, serverStore repository.MCPServerStore, buildSvc *ServerBuildService) *DynamicRegistry {
	serverName := "dynamic-mcp-go-server"
	serverVersion := "2.0.0"

	if name := os.Getenv("MCP_SERVER_NAME"); name != "" {
		serverName = name
	}
	if version := os.Getenv("MCP_SERVER_VERSION"); version != "" {
		serverVersion = version
	}

	return &DynamicRegistry{
		server:        s,
		store:         store,
		interval:      interval,
		logger:        log,
		groupMCP:      groupMCP,
		serverName:    serverName,
		serverVersion: serverVersion,
		serverStore:   serverStore,
		buildSvc:      buildSvc,
	}
}

func toServerTool(def tooldef.ToolDefinition) (server.ServerTool, error) {
	params, err := parseToolParams(def.Parameters)
	if err != nil {
		return server.ServerTool{}, fmt.Errorf("parse parameters for tool %q: %w", def.Name, err)
	}
	rawSchema, err := toRawInputSchema(params)
	if err != nil {
		return server.ServerTool{}, err
	}

	tool := mcp.NewToolWithRawSchema(def.Name, def.Description, rawSchema)
	return server.ServerTool{
		Tool:    tool,
		Handler: defaultHandler(def),
	}, nil
}

func parseToolParams(data []byte) ([]tooldef.ParameterDefinition, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var params []tooldef.ParameterDefinition
	if err := sonic.Unmarshal(data, &params); err != nil {
		return nil, err
	}
	return params, nil
}

func defaultHandler(def tooldef.ToolDefinition) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		payload := map[string]any{
			"tool":      def.Name,
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

func toRawInputSchema(params []tooldef.ParameterDefinition) ([]byte, error) {
	properties := make(map[string]any, len(params))
	required := make([]string, 0)

	for _, p := range params {
		prop := map[string]any{
			"type": string(p.Type),
		}
		if p.Description != "" {
			prop["description"] = p.Description
		}
		if p.Default != nil {
			prop["default"] = p.Default
		}
		if len(p.Enum) > 0 {
			prop["enum"] = p.Enum
		}
		if p.Minimum != nil {
			prop["minimum"] = *p.Minimum
		}
		if p.Maximum != nil {
			prop["maximum"] = *p.Maximum
		}
		properties[p.Name] = prop
		if p.Required {
			required = append(required, p.Name)
		}
	}

	sort.Strings(required)
	schema := map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
	raw, err := sonic.Marshal(schema)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func hashDefinitions(defs []tooldef.ToolDefinition) (string, error) {
	b, err := sonic.Marshal(defs)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func hashPerGroup(_ []tooldef.ToolDefinition) (map[string]string, error) {
	return nil, nil
}

func hashDefinitionSubset(defs []tooldef.ToolDefinition) (string, error) {
	sorted := cloneDefinitions(defs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	b, err := sonic.Marshal(sorted)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func changedGroupKeys(prev, next map[string]string) map[string]bool {
	out := make(map[string]bool)
	for k, v := range next {
		if prev[k] != v {
			out[k] = true
		}
	}
	for k := range prev {
		if _, ok := next[k]; !ok {
			out[k] = true
		}
	}
	return out
}

func (d *DynamicRegistry) ListDefinitions() []tooldef.ToolDefinition {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return cloneDefinitions(d.lastDefs)
}

func (d *DynamicRegistry) GetDefinition(name string) (tooldef.ToolDefinition, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, def := range d.lastDefs {
		if def.Name == name {
			return cloneDefinition(def), true
		}
	}
	return tooldef.ToolDefinition{}, false
}

// ListDefinitionsByVAuthKey 按 vauthKey 查询工具定义
func (d *DynamicRegistry) ListDefinitionsByVAuthKey(vauthKey string) []tooldef.ToolDefinition {
	if d.buildSvc == nil || d.serverStore == nil {
		return nil
	}

	ctx := context.Background()
	mcpServer, err := d.serverStore.GetByVAuthKey(ctx, vauthKey)
	if err != nil || mcpServer == nil {
		return nil
	}

	buildInfo, err := d.buildSvc.GetActiveBuild(ctx, mcpServer.ID)
	if err != nil || buildInfo == nil {
		d.logger.Debug("ListDefinitionsByVAuthKey: no active build", zap.String("vauthKey", vauthKey), zap.Uint("serverID", mcpServer.ID))
		return nil
	}

	d.logger.Info("ListDefinitionsByVAuthKey",
		zap.String("vauthKey", vauthKey),
		zap.Uint("serverID", mcpServer.ID),
		zap.Uint("buildID", buildInfo.ID),
		zap.String("buildHash", buildInfo.Hash),
		zap.Int("buildDataLen", len(buildInfo.BuildData)))

	var buildData model.BuildData
	if err := sonic.Unmarshal([]byte(buildInfo.BuildData), &buildData); err != nil {
		d.logger.Error("ListDefinitionsByVAuthKey: unmarshal build data failed",
			zap.String("vauthKey", vauthKey),
			zap.Error(err))
		return nil
	}

	d.logger.Info("ListDefinitionsByVAuthKey buildData",
		zap.String("vauthKey", vauthKey),
		zap.Int("toolsCount", len(buildData.Tools)))

	defs := make([]tooldef.ToolDefinition, 0, len(buildData.Tools))
	for _, t := range buildData.Tools {
		if t.State == 1 {
			paramsBytes, _ := sonic.Marshal(t.Parameters)
			defs = append(defs, tooldef.ToolDefinition{
				ID:          t.ID,
				Name:        t.Name,
				Description: t.Description,
				Parameters:  paramsBytes,
				State:       t.State,
			})
		}
	}
	d.logger.Info("ListDefinitionsByVAuthKey result",
		zap.String("vauthKey", vauthKey),
		zap.Int("defsCount", len(defs)))
	return defs
}

// GetDefinitionByVAuthKey 按 vauthKey 和 toolName 查询工具定义
func (d *DynamicRegistry) GetDefinitionByVAuthKey(vauthKey, name string) (tooldef.ToolDefinition, bool) {
	defs := d.ListDefinitionsByVAuthKey(vauthKey)
	for _, def := range defs {
		if def.Name == name {
			return def, true
		}
	}
	return tooldef.ToolDefinition{}, false
}

// GetVAuthKeyDescription 获取 vauthKey 的描述
func (d *DynamicRegistry) GetVAuthKeyDescription(vauthKey string) (string, bool) {
	return "", false
}

func cloneDefinitions(defs []tooldef.ToolDefinition) []tooldef.ToolDefinition {
	out := make([]tooldef.ToolDefinition, 0, len(defs))
	for _, def := range defs {
		out = append(out, cloneDefinition(def))
	}
	return out
}

func cloneDefinition(def tooldef.ToolDefinition) tooldef.ToolDefinition {
	cp := def
	if def.Parameters != nil {
		cp.Parameters = make([]byte, len(def.Parameters))
		copy(cp.Parameters, def.Parameters)
	}
	return cp
}

// ServerName 返回服务器名称
func (d *DynamicRegistry) ServerName() string {
	if d.serverName != "" {
		return d.serverName
	}
	return "dynamic-mcp-go-server"
}

// GetLogger 返回日志器
func (d *DynamicRegistry) GetLogger() logger.Logger {
	return d.logger
}

// ServerVersion 返回服务器版本
func (d *DynamicRegistry) ServerVersion() string {
	if d.serverVersion != "" {
		return d.serverVersion
	}
	return "2.0.0"
}
