package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

type ToolQuery struct {
	ID        *uint
	Name      *string
	ServiceID *uint
	State     *int
	Keyword   *string
}

func (q *ToolQuery) HasCondition() bool {
	return q.ID != nil || q.Name != nil || q.ServiceID != nil || q.State != nil || q.Keyword != nil
}

// ToolStore 定义工具定义存储接口
type ToolStore interface {
	GetByID(ctx context.Context, id uint) (*model.ToolDefinition, error)
	GetByName(ctx context.Context, name string) (*model.ToolDefinition, error)
	List(ctx context.Context, query *ToolQuery, page, pageSize int) ([]*model.ToolDefinition, int64, error)
	SaveTool(ctx context.Context, tool *model.ToolDefinition) error
	DeleteTool(ctx context.Context, id uint) error
	Delete(ctx context.Context, id uint) error
	Create(ctx context.Context, tool *model.ToolDefinition) error
	Update(ctx context.Context, tool *model.ToolDefinition) error
}
