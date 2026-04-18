package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

// ToolStore 定义工具定义存储接口
type ToolStore interface {
	GetByID(ctx context.Context, id uint) (*model.ToolDefinition, error)
	List(ctx context.Context) ([]*model.ToolDefinition, error)
	SaveTool(ctx context.Context, tool *model.ToolDefinition) error
	DeleteTool(ctx context.Context, id uint) error
	Save(ctx context.Context, tool *model.ToolDefinition) error
}
