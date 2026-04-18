package service

import (
	"context"
	"errors"

	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"
)

var (
	ErrToolNotFound   = errors.New("tool not found")
	ErrServerNotFound = errors.New("server not found")
	ErrBindingExists  = errors.New("binding already exists")
)

type ToolBindingService struct {
	bindingStore repository.ToolServerBindingStore
	toolStore    repository.ToolStore
	serverStore  repository.MCPServerStore
}

func NewToolBindingService(
	bindingStore repository.ToolServerBindingStore,
	toolStore repository.ToolStore,
	serverStore repository.MCPServerStore,
) *ToolBindingService {
	return &ToolBindingService{
		bindingStore: bindingStore,
		toolStore:    toolStore,
		serverStore:  serverStore,
	}
}

type BindToolRequest struct {
	ToolID   uint
	ServerID uint
}

type BatchBindRequest struct {
	ToolIDs   []uint
	ServerIDs []uint
}

func (s *ToolBindingService) BindTool(ctx context.Context, req BindToolRequest) (*model.ToolServerBinding, error) {
	tool, err := s.toolStore.GetByID(ctx, req.ToolID)
	if err != nil {
		return nil, ErrToolNotFound
	}
	if tool == nil {
		return nil, ErrToolNotFound
	}

	server, err := s.serverStore.GetByID(ctx, req.ServerID)
	if err != nil {
		return nil, ErrServerNotFound
	}
	if server == nil {
		return nil, ErrServerNotFound
	}

	existing, err := s.bindingStore.GetByToolAndServer(ctx, req.ToolID, req.ServerID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrBindingExists
	}

	binding := &model.ToolServerBinding{
		ToolID:   req.ToolID,
		ServerID: req.ServerID,
	}

	if err := s.bindingStore.Save(ctx, binding); err != nil {
		return nil, err
	}

	return binding, nil
}

func (s *ToolBindingService) UnbindTool(ctx context.Context, req BindToolRequest) error {
	binding, err := s.bindingStore.GetByToolAndServer(ctx, req.ToolID, req.ServerID)
	if err != nil {
		return err
	}
	if binding == nil {
		return nil
	}

	return s.bindingStore.Delete(ctx, binding.ID)
}

func (s *ToolBindingService) BatchBindTools(ctx context.Context, req BatchBindRequest) (int, error) {
	if len(req.ToolIDs) == 0 || len(req.ServerIDs) == 0 {
		return 0, nil
	}

	count := 0
	for _, toolID := range req.ToolIDs {
		for _, serverID := range req.ServerIDs {
			binding := &model.ToolServerBinding{
				ToolID:   toolID,
				ServerID: serverID,
			}
			if err := s.bindingStore.Save(ctx, binding); err == nil {
				count++
			}
		}
	}

	return count, nil
}

func (s *ToolBindingService) BatchUnbindTools(ctx context.Context, bindingIDs []uint) (int, error) {
	if len(bindingIDs) == 0 {
		return 0, nil
	}

	deleted := 0
	for _, id := range bindingIDs {
		if err := s.bindingStore.Delete(ctx, id); err == nil {
			deleted++
		}
	}

	return deleted, nil
}

func (s *ToolBindingService) GetToolBindings(ctx context.Context, toolID uint) ([]*model.ToolServerBinding, error) {
	return s.bindingStore.ListByToolID(ctx, toolID)
}

func (s *ToolBindingService) GetServerBindings(ctx context.Context, serverID uint) ([]*model.ToolServerBinding, error) {
	return s.bindingStore.ListByServerID(ctx, serverID)
}