package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

// ToolStore 定义工具定义存储接口
type ToolStore interface {
	List(ctx context.Context) ([]*model.ToolDefinition, error)
	ListByServiceID(ctx context.Context, serviceID uint) ([]*model.ToolDefinition, error)
	ListByVAuthKey(ctx context.Context, vauthKey string) ([]*model.ToolDefinition, error)
	AddToolsToServer(ctx context.Context, serverID uint, toolNames []string) error
	RemoveToolFromServer(ctx context.Context, serverID uint, toolName string) error
	SaveTool(ctx context.Context, tool *model.ToolDefinition) error
	DeleteTool(ctx context.Context, id uint) error
	GetByNameAndServer(ctx context.Context, name, vauthKey string) (*model.ToolDefinition, error)
	Save(ctx context.Context, tool *model.ToolDefinition) error
	AddToolWithVAuthKey(ctx context.Context, tool *model.ToolDefinition) error
	RemoveToolByNameAndVAuthKey(ctx context.Context, name, vauthKey string) error
}
