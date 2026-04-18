package service

import (
	"context"
	"errors"
	"fmt"

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

	existingIncludeDeleted, err := s.bindingStore.GetByToolAndServerIncludeDeleted(ctx, req.ToolID, req.ServerID)
	if err != nil {
		return nil, err
	}
	if existingIncludeDeleted != nil {
		if err := s.bindingStore.Restore(ctx, existingIncludeDeleted.ID); err != nil {
			return nil, fmt.Errorf("failed to restore binding: %w", err)
		}
		existingIncludeDeleted.State = 1
		return existingIncludeDeleted, nil
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

	allBindings, err := s.bindingStore.ListAllIncludeDeleted(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list all bindings: %w", err)
	}

	existingMap := make(map[uint]map[uint]int)
	for _, b := range allBindings {
		if existingMap[b.ToolID] == nil {
			existingMap[b.ToolID] = make(map[uint]int)
		}
		existingMap[b.ToolID][b.ServerID] = b.State
	}

	var toRestore []uint
	var toCreate []*model.ToolServerBinding

	for _, toolID := range req.ToolIDs {
		for _, serverID := range req.ServerIDs {
			state, exists := existingMap[toolID][serverID]
			if !exists {
				toCreate = append(toCreate, &model.ToolServerBinding{
					ToolID:   toolID,
					ServerID: serverID,
				})
			} else if state == 0 {
				binding, err := s.bindingStore.GetByToolAndServerIncludeDeleted(ctx, toolID, serverID)
				if err != nil {
					continue
				}
				if binding != nil {
					toRestore = append(toRestore, binding.ID)
				}
			}
		}
	}

	if len(toRestore) > 0 {
		if err := s.bindingStore.BatchRestore(ctx, toRestore); err != nil {
			return 0, fmt.Errorf("failed to restore bindings: %w", err)
		}
	}

	if len(toCreate) > 0 {
		if err := s.bindingStore.BatchSave(ctx, toCreate); err != nil {
			return 0, fmt.Errorf("failed to create bindings: %w", err)
		}
	}

	return len(toRestore) + len(toCreate), nil
}

func (s *ToolBindingService) BatchUnbindTools(ctx context.Context, bindingIDs []uint) (int, error) {
	if len(bindingIDs) == 0 {
		return 0, nil
	}

	existingBindings, err := s.bindingStore.ListAllIncludeDeleted(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list all bindings: %w", err)
	}

	existingMap := make(map[uint]int)
	for _, b := range existingBindings {
		existingMap[b.ID] = b.State
	}

	var toInvalidate []uint
	for _, id := range bindingIDs {
		state, exists := existingMap[id]
		if exists && state == 1 {
			toInvalidate = append(toInvalidate, id)
		}
	}

	if len(toInvalidate) > 0 {
		if err := s.bindingStore.BatchDelete(ctx, toInvalidate); err != nil {
			return 0, fmt.Errorf("failed to invalidate bindings: %w", err)
		}
	}

	return len(toInvalidate), nil
}

func (s *ToolBindingService) GetToolBindings(ctx context.Context, toolID uint) ([]*model.ToolServerBinding, error) {
	return s.bindingStore.ListByToolID(ctx, toolID)
}

func (s *ToolBindingService) GetServerBindings(ctx context.Context, serverID uint) ([]*model.ToolServerBinding, error) {
	return s.bindingStore.ListByServerID(ctx, serverID)
}