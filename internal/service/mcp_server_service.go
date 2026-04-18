package service

import (
	"context"
	"errors"
	"fmt"

	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"

	"gorm.io/gorm"
)

var (
	ErrMCPServerNotFound         = errors.New("mcp server not found")
	ErrMCPServerExists           = errors.New("mcp server with this vauth_key already exists")
	ErrServerTypeCannotBeChanged = errors.New("server type cannot be changed after creation")
)

// MCPServerService 提供 MCPServer 的业务逻辑
type MCPServerService struct {
	serverStore        repository.MCPServerStore
	bindingStore       repository.TokenServerBindingStore
	toolStore          repository.ToolStore
	toolServerBindingStore repository.ToolServerBindingStore
}

// NewMCPServerService 创建 MCPServerService
func NewMCPServerService(
	serverStore repository.MCPServerStore,
	bindingStore repository.TokenServerBindingStore,
	toolStore repository.ToolStore,
	toolServerBindingStore repository.ToolServerBindingStore,
) *MCPServerService {
	return &MCPServerService{
		serverStore:        serverStore,
		bindingStore:       bindingStore,
		toolStore:          toolStore,
		toolServerBindingStore: toolServerBindingStore,
	}
}

// ListServers 获取所有 MCPServer
func (s *MCPServerService) ListServers(ctx context.Context) ([]*model.MCPServer, error) {
	return s.serverStore.List(ctx)
}

// GetServer 获取单个 MCPServer
func (s *MCPServerService) GetServer(ctx context.Context, id uint) (*model.MCPServer, error) {
	server, err := s.serverStore.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMCPServerNotFound
		}
		return nil, err
	}
	return server, nil
}

// CreateServer 创建 MCPServer
func (s *MCPServerService) CreateServer(ctx context.Context, server *model.MCPServer) error {
	// 检查 vauth_key 是否已存在
	existing, err := s.serverStore.GetByVAuthKey(ctx, server.VAuthKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil {
		return ErrMCPServerExists
	}

	return s.serverStore.Save(ctx, server)
}

// UpdateServer 更新 MCPServer
func (s *MCPServerService) UpdateServer(ctx context.Context, server *model.MCPServer) error {
	// 检查是否存在
	existing, err := s.serverStore.GetByID(ctx, server.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMCPServerNotFound
		}
		return err
	}

	// 类型不可修改校验
	if existing.Type != server.Type {
		return ErrServerTypeCannotBeChanged
	}

	// 如果修改了 vauth_key，检查新 vauth_key 是否与其他服务器冲突
	if server.VAuthKey != existing.VAuthKey {
		conflict, err := s.serverStore.GetByVAuthKey(ctx, server.VAuthKey)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if conflict != nil {
			return ErrMCPServerExists
		}
	}

	return s.serverStore.Save(ctx, server)
}

// DeleteServer 删除 MCPServer
func (s *MCPServerService) DeleteServer(ctx context.Context, id uint) error {
	// 检查是否存在
	_, err := s.serverStore.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMCPServerNotFound
		}
		return err
	}

	return s.serverStore.Delete(ctx, id)
}

// GetServerTools 获取 Server 下的所有工具
func (s *MCPServerService) GetServerTools(ctx context.Context, serverID uint) ([]*model.ToolDefinition, error) {
	bindings, err := s.toolServerBindingStore.ListByServerID(ctx, serverID)
	if err != nil {
		return nil, err
	}

	toolMap := make(map[uint]*model.ToolDefinition)
	for _, binding := range bindings {
		tool, err := s.toolStore.GetByID(ctx, binding.ToolID)
		if err != nil {
			continue
		}
		toolMap[tool.ID] = tool
	}

	tools := make([]*model.ToolDefinition, 0, len(bindings))
	for _, binding := range bindings {
		if tool, ok := toolMap[binding.ToolID]; ok {
			tools = append(tools, tool)
		}
	}
	return tools, nil
}

// AddToolToServer 向 Server 添加工具
func (s *MCPServerService) AddToolToServer(ctx context.Context, serverID uint, tool *model.ToolDefinition) error {
	if err := model.ValidateToolName(tool.Name); err != nil {
		return err
	}

	_, err := s.serverStore.GetByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMCPServerNotFound
		}
		return err
	}

	tool.Enabled = true

	if err := s.toolStore.SaveTool(ctx, tool); err != nil {
		return err
	}

	binding := &model.ToolServerBinding{
		ToolID:   tool.ID,
		ServerID: serverID,
	}
	return s.toolServerBindingStore.Save(ctx, binding)
}

// RemoveToolFromServer 从 Server 移除工具
func (s *MCPServerService) RemoveToolFromServerByName(ctx context.Context, serverID uint, toolName string) error {
	tools, err := s.GetServerTools(ctx, serverID)
	if err != nil {
		return err
	}

	var toolID uint
	for _, tool := range tools {
		if tool.Name == toolName {
			toolID = tool.ID
			break
		}
	}
	if toolID == 0 {
		return fmt.Errorf("tool not found: %s", toolName)
	}

	binding, err := s.toolServerBindingStore.GetByToolAndServer(ctx, toolID, serverID)
	if err != nil {
		return err
	}

	return s.toolServerBindingStore.Delete(ctx, binding.ID)
}

// AddToolsToServer 向 Server 添加工具
func (s *MCPServerService) AddToolsToServer(ctx context.Context, serverID uint, toolNames []string) error {
	_, err := s.serverStore.GetByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMCPServerNotFound
		}
		return err
	}

	var toolIDs []uint
	for _, name := range toolNames {
		tools, err := s.toolStore.List(ctx)
		if err != nil {
			continue
		}
		for _, tool := range tools {
			if tool.Name == name {
				toolIDs = append(toolIDs, tool.ID)
				break
			}
		}
	}

	bindings := make([]*model.ToolServerBinding, 0, len(toolIDs))
	for _, toolID := range toolIDs {
		bindings = append(bindings, &model.ToolServerBinding{
			ToolID:   toolID,
			ServerID: serverID,
		})
	}

	return s.toolServerBindingStore.BatchSave(ctx, bindings)
}

// RemoveToolFromServer 从 Server 移除工具
func (s *MCPServerService) RemoveToolFromServer(ctx context.Context, serverID uint, toolName string) error {
	return s.RemoveToolFromServerByName(ctx, serverID, toolName)
}
