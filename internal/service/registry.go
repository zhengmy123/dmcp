package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/infrastructure/store/tooldef"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
}

func NewDynamicRegistry(s *server.MCPServer, store tooldef.Store, interval time.Duration, log logger.Logger, groupMCP *MCPGroupManager) *DynamicRegistry {
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
	}
}

func (d *DynamicRegistry) SyncOnce(ctx context.Context) error {
	defs, err := d.store.List(ctx)
	if err != nil {
		return fmt.Errorf("list tool definitions: %w", err)
	}

	currentHash, err := hashDefinitions(defs)
	if err != nil {
		return fmt.Errorf("hash definitions: %w", err)
	}
	if currentHash == d.lastHash {
		return nil
	}

	d.mu.RLock()
	prevDefs := cloneDefinitions(d.lastDefs)
	d.mu.RUnlock()

	prevGroupHashes, err := hashPerGroup(prevDefs)
	if err != nil {
		return fmt.Errorf("hash per-group (previous): %w", err)
	}
	nextGroupHashes, err := hashPerGroup(defs)
	if err != nil {
		return fmt.Errorf("hash per-group (current): %w", err)
	}
	changedGroups := changedGroupKeys(prevGroupHashes, nextGroupHashes)

	tools := make([]server.ServerTool, 0, len(defs))
	groupPartial := make(map[string][]server.ServerTool)
	for _, def := range defs {
		t, err := toServerTool(def)
		if err != nil {
			return fmt.Errorf("build tool %q: %w", def.Name, err)
		}
		tools = append(tools, t)
		vk := CanonicalVAuthKey(def.VAuthKey)
		if vk != "" && changedGroups[vk] {
			groupPartial[vk] = append(groupPartial[vk], t)
		}
	}

	d.server.SetTools(tools...)
	if d.groupMCP != nil && len(changedGroups) > 0 {
		var removeGroups []string
		for vk := range changedGroups {
			if len(groupPartial[vk]) == 0 {
				removeGroups = append(removeGroups, vk)
			}
		}
		d.groupMCP.PatchHandlers(groupPartial, removeGroups)
	}
	d.server.SendNotificationToAllClients(toolsListChangedMethod, map[string]any{})

	d.mu.Lock()
	d.lastDefs = cloneDefinitions(defs)
	d.mu.Unlock()

	d.lastHash = currentHash
	if len(changedGroups) > 0 {
		d.logger.Info("tool registry refreshed",
			logger.Int("tools", len(tools)),
			logger.Int("group_handlers_patched", len(changedGroups)),
		)
	} else {
		d.logger.Info("tool registry refreshed",
			logger.Int("tools", len(tools)),
			logger.String("group_handlers", "unchanged"),
		)
	}
	return nil
}

func (d *DynamicRegistry) Start(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.SyncOnce(ctx); err != nil {
				d.logger.Error("sync tool definitions failed", logger.Error(err))
			}
		}
	}
}

func toServerTool(def tooldef.ToolDefinition) (server.ServerTool, error) {
	rawSchema, err := toRawInputSchema(def.Parameters)
	if err != nil {
		return server.ServerTool{}, err
	}

	tool := mcp.NewToolWithRawSchema(def.Name, def.Description, rawSchema)
	return server.ServerTool{
		Tool:    tool,
		Handler: defaultHandler(def),
	}, nil
}

func defaultHandler(def tooldef.ToolDefinition) server.ToolHandlerFunc {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		payload := map[string]any{
			"tool":      def.Name,
			"vauth_key": CanonicalVAuthKey(def.VAuthKey),
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

func toRawInputSchema(params []tooldef.ParameterDefinition) (json.RawMessage, error) {
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
	raw, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func hashDefinitions(defs []tooldef.ToolDefinition) (string, error) {
	b, err := json.Marshal(defs)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func hashPerGroup(defs []tooldef.ToolDefinition) (map[string]string, error) {
	byGroup := make(map[string][]tooldef.ToolDefinition)
	for _, d := range defs {
		vk := CanonicalVAuthKey(d.VAuthKey)
		if vk == "" {
			continue
		}
		byGroup[vk] = append(byGroup[vk], d)
	}
	out := make(map[string]string, len(byGroup))
	for vk, subset := range byGroup {
		h, err := hashDefinitionSubset(subset)
		if err != nil {
			return nil, err
		}
		out[vk] = h
	}
	return out, nil
}

func hashDefinitionSubset(defs []tooldef.ToolDefinition) (string, error) {
	sorted := cloneDefinitions(defs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	b, err := json.Marshal(sorted)
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
	d.mu.RLock()
	defer d.mu.RUnlock()

	target := CanonicalVAuthKey(vauthKey)
	out := make([]tooldef.ToolDefinition, 0)
	for _, def := range d.lastDefs {
		if CanonicalVAuthKey(def.VAuthKey) != target {
			continue
		}
		out = append(out, cloneDefinition(def))
	}
	return out
}

// GetDefinitionByVAuthKey 按 vauthKey 和 toolName 查询工具定义
func (d *DynamicRegistry) GetDefinitionByVAuthKey(vauthKey, name string) (tooldef.ToolDefinition, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	target := CanonicalVAuthKey(vauthKey)
	for _, def := range d.lastDefs {
		if def.Name != name {
			continue
		}
		if CanonicalVAuthKey(def.VAuthKey) != target {
			continue
		}
		return cloneDefinition(def), true
	}
	return tooldef.ToolDefinition{}, false
}

// GetVAuthKeyDescription 获取 vauthKey 的描述
func (d *DynamicRegistry) GetVAuthKeyDescription(vauthKey string) (string, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	target := CanonicalVAuthKey(vauthKey)
	for _, def := range d.lastDefs {
		if CanonicalVAuthKey(def.VAuthKey) != target {
			continue
		}
		if strings.TrimSpace(def.ServerDesc) != "" {
			return def.ServerDesc, true
		}
	}
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
	cp.Parameters = append([]tooldef.ParameterDefinition(nil), def.Parameters...)
	cp.VAuthKey = CanonicalVAuthKey(cp.VAuthKey)
	return cp
}

// ServerName 返回服务器名称
func (d *DynamicRegistry) ServerName() string {
	if d.serverName != "" {
		return d.serverName
	}
	return "dynamic-mcp-go-server"
}

// ServerVersion 返回服务器版本
func (d *DynamicRegistry) ServerVersion() string {
	if d.serverVersion != "" {
		return d.serverVersion
	}
	return "2.0.0"
}
