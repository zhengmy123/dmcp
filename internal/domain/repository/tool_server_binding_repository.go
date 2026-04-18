package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

type ToolServerBindingStore interface {
	ListByToolID(ctx context.Context, toolID uint) ([]*model.ToolServerBinding, error)
	ListByServerID(ctx context.Context, serverID uint) ([]*model.ToolServerBinding, error)
	ListByToolIDs(ctx context.Context, toolIDs []uint) ([]*model.ToolServerBinding, error)
	ListByServerIDs(ctx context.Context, serverIDs []uint) ([]*model.ToolServerBinding, error)
	ListAllIncludeDeleted(ctx context.Context) ([]*model.ToolServerBinding, error)
	GetByToolAndServer(ctx context.Context, toolID, serverID uint) (*model.ToolServerBinding, error)
	ExistByToolAndServer(ctx context.Context, toolID, serverID uint) (bool, error)
	GetByToolAndServerIncludeDeleted(ctx context.Context, toolID, serverID uint) (*model.ToolServerBinding, error)
	Save(ctx context.Context, binding *model.ToolServerBinding) error
	BatchSave(ctx context.Context, bindings []*model.ToolServerBinding) error
	Delete(ctx context.Context, id uint) error
	DeleteByToolID(ctx context.Context, toolID uint) error
	DeleteByServerID(ctx context.Context, serverID uint) error
	BatchDelete(ctx context.Context, ids []uint) error
	ReplaceByToolID(ctx context.Context, toolID uint, serverIDs []uint) error
	Restore(ctx context.Context, id uint) error
	BatchRestore(ctx context.Context, ids []uint) error
}