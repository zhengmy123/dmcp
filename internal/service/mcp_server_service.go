package service

import (
	"context"
	"errors"

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
	serverStore  repository.MCPServerStore
	bindingStore repository.TokenServerBindingStore
	toolStore    repository.ToolStore
}

// NewMCPServerService 创建 MCPServerService
func NewMCPServerService(
	serverStore repository.MCPServerStore,
	bindingStore repository.TokenServerBindingStore,
	toolStore repository.ToolStore,
) *MCPServerService {
	return &MCPServerService{
		serverStore:  serverStore,
		bindingStore: bindingStore,
		toolStore:    toolStore,
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
	// 验证服务器存在
	server, err := s.serverStore.GetByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMCPServerNotFound
		}
		return nil, err
	}

	return s.toolStore.ListByVAuthKey(ctx, server.VAuthKey)
}

// AddToolToServer 向 Server 添加工具
func (s *MCPServerService) AddToolToServer(ctx context.Context, serverID uint, tool *model.ToolDefinition) error {
	// 验证服务器存在
	server, err := s.serverStore.GetByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMCPServerNotFound
		}
		return err
	}

	tool.VAuthKey = server.VAuthKey
	tool.Enabled = true

	return s.toolStore.AddToolWithVAuthKey(ctx, tool)
}

// RemoveToolFromServer 从 Server 移除工具
func (s *MCPServerService) RemoveToolFromServerByName(ctx context.Context, serverID uint, toolName string) error {
	// 验证服务器存在
	server, err := s.serverStore.GetByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMCPServerNotFound
		}
		return err
	}

	return s.toolStore.RemoveToolByNameAndVAuthKey(ctx, toolName, server.VAuthKey)
}

// AddToolsToServer 向 Server 添加工具
func (s *MCPServerService) AddToolsToServer(ctx context.Context, serverID uint, toolNames []string) error {
	// 验证服务器存在
	_, err := s.serverStore.GetByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMCPServerNotFound
		}
		return err
	}

	return s.toolStore.AddToolsToServer(ctx, serverID, toolNames)
}

// RemoveToolFromServer 从 Server 移除工具
func (s *MCPServerService) RemoveToolFromServer(ctx context.Context, serverID uint, toolName string) error {
	// 验证服务器存在
	_, err := s.serverStore.GetByID(ctx, serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMCPServerNotFound
		}
		return err
	}

	return s.toolStore.RemoveToolFromServer(ctx, serverID, toolName)
}
