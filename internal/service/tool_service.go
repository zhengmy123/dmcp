package service

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
	domainService "dynamic_mcp_go_server/internal/domain/service"
)

// ToolService 工具应用服务
type ToolService struct {
	toolDomainService *domainService.ToolDomainService
}

func NewToolService(
	toolDomainService *domainService.ToolDomainService,
) *ToolService {
	return &ToolService{
		toolDomainService: toolDomainService,
	}
}

// CreateFromHTTPService 从 HTTPService 创建工具
func (s *ToolService) CreateFromHTTPService(ctx context.Context, cmd domainService.CreateToolFromHTTPServiceCommand) (*model.ToolDefinition, error) {
	return s.toolDomainService.CreateToolFromHTTPService(ctx, cmd)
}
