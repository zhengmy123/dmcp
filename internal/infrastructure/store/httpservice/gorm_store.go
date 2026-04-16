package httpservice

import (
	"context"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"
	"dynamic_mcp_go_server/internal/infrastructure/database"

	"gorm.io/gorm"
)

var _ repository.ServiceStore = (*GORMServiceStore)(nil)

// GORMServiceStore GORM实现的MySQL存储
type GORMServiceStore struct {
	dao *database.GORMServiceDAO
}

// NewGORMServiceStore 创建GORM MySQL存储
func NewGORMServiceStore(db *gorm.DB, log logger.Logger) *GORMServiceStore {
	return &GORMServiceStore{
		dao: database.NewGORMServiceDAO(db, log),
	}
}

func (s *GORMServiceStore) List(ctx context.Context) ([]*model.HTTPService, error) {
	return s.dao.List(ctx)
}

func (s *GORMServiceStore) Get(ctx context.Context, id uint) (*model.HTTPService, error) {
	return s.dao.Get(ctx, id)
}

func (s *GORMServiceStore) Save(ctx context.Context, service *model.HTTPService) error {
	return s.dao.Save(ctx, service)
}

func (s *GORMServiceStore) Delete(ctx context.Context, id uint) error {
	return s.dao.Delete(ctx, id)
}
