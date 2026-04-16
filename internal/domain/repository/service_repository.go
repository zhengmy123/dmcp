package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

// ServiceStore 定义HTTP服务存储接口
type ServiceStore interface {
	List(ctx context.Context) ([]*model.HTTPService, error)
	Get(ctx context.Context, id uint) (*model.HTTPService, error)
	Save(ctx context.Context, service *model.HTTPService) error
	Delete(ctx context.Context, id uint) error
}
