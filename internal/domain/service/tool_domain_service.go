package service

import (
	"context"
	"errors"

	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"
)

var (
	ErrOnlyHTTPServiceServerCanHaveTools = errors.New("only http_service server can have tools")
	ErrToolNameAlreadyExists              = errors.New("tool with same name already exists in this server")
	ErrHTTPServiceNotFound                = errors.New("http service not found")
)

type CreateToolFromHTTPServiceCommand struct {
	Name          string
	Description   string
	ServerID      uint
	ServiceID     uint
	InputExtra    []byte
	OutputMapping []byte
}

// ToolDomainService 工具领域服务
type ToolDomainService struct {
	toolStore    repository.ToolStore
	serverStore  repository.MCPServerStore
	serviceStore repository.ServiceStore
}

func NewToolDomainService(
	toolStore repository.ToolStore,
	serverStore repository.MCPServerStore,
	serviceStore repository.ServiceStore,
) *ToolDomainService {
	return &ToolDomainService{
		toolStore:    toolStore,
		serverStore:  serverStore,
		serviceStore: serviceStore,
	}
}

// CreateToolFromHTTPService 从 HTTP Service 创建工具
func (s *ToolDomainService) CreateToolFromHTTPService(ctx context.Context, cmd CreateToolFromHTTPServiceCommand) (*model.ToolDefinition, error) {
	// 1. 校验 MCPServer 存在且类型为 http_service
	server, err := s.serverStore.GetByID(ctx, cmd.ServerID)
	if err != nil {
		return nil, errors.New("mcp server not found")
	}
	if server.Type != "http_service" {
		return nil, ErrOnlyHTTPServiceServerCanHaveTools
	}

	// 2. 校验 HTTPService 存在
	_, err = s.serviceStore.Get(ctx, cmd.ServiceID)
	if err != nil {
		return nil, ErrHTTPServiceNotFound
	}

	// 3. 校验工具名称不重复
	existing, _ := s.toolStore.GetByNameAndServer(ctx, cmd.Name, server.VAuthKey)
	if existing != nil {
		return nil, ErrToolNameAlreadyExists
	}

	// 4. 创建工具
	tool := &model.ToolDefinition{
		Name:          cmd.Name,
		Description:   cmd.Description,
		VAuthKey:      server.VAuthKey,
		ServiceID:     cmd.ServiceID,
		InputExtra:    cmd.InputExtra,
		OutputMapping: cmd.OutputMapping,
		Enabled:       true,
	}

	if err := s.toolStore.Save(ctx, tool); err != nil {
		return nil, err
	}

	return tool, nil
}
