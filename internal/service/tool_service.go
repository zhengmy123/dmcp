package service

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
	domainService "dynamic_mcp_go_server/internal/domain/service"
)

type ToolService struct {
	toolDomainService *domainService.ToolDomainService
	buildSvc         *ServerBuildService
}

func NewToolService(
	toolDomainService *domainService.ToolDomainService,
	buildSvc *ServerBuildService,
) *ToolService {
	return &ToolService{
		toolDomainService: toolDomainService,
		buildSvc:         buildSvc,
	}
}

func (s *ToolService) CreateFromHTTPService(ctx context.Context, cmd domainService.CreateToolFromHTTPServiceCommand) (*model.ToolDefinition, error) {
	tool, err := s.toolDomainService.CreateToolFromHTTPService(ctx, cmd)
	if err != nil {
		return nil, err
	}

	if s.buildSvc != nil {
		_ = s.buildSvc.BuildOrUpdate(ctx, cmd.ServerID)
	}

	return tool, nil
}
