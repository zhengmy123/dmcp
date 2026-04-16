package httpservice

import (
	"context"
	"fmt"
	"time"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"
)

var _ repository.ServiceStore = (*MemoryServiceStore)(nil)

// MemoryServiceStore 内存存储实现
type MemoryServiceStore struct {
	services map[uint]*model.HTTPService
	logger   logger.Logger
}

// NewMemoryServiceStore 创建内存存储
func NewMemoryServiceStore(log logger.Logger) *MemoryServiceStore {
	return &MemoryServiceStore{
		services: make(map[uint]*model.HTTPService),
		logger:   log,
	}
}

func (s *MemoryServiceStore) List(ctx context.Context) ([]*model.HTTPService, error) {
	services := make([]*model.HTTPService, 0, len(s.services))
	for _, service := range s.services {
		services = append(services, service)
	}
	return services, nil
}

func (s *MemoryServiceStore) Get(ctx context.Context, id uint) (*model.HTTPService, error) {
	service, exists := s.services[id]
	if !exists {
		return nil, fmt.Errorf("service not found: %d", id)
	}
	return service, nil
}

func (s *MemoryServiceStore) Save(ctx context.Context, service *model.HTTPService) error {
	if service.ID == 0 {
		return fmt.Errorf("service ID is required")
	}

	now := time.Now()
	if service.CreatedAt.IsZero() {
		service.CreatedAt = now
	}
	service.UpdatedAt = now

	s.services[service.ID] = service
	return nil
}

func (s *MemoryServiceStore) Delete(ctx context.Context, id uint) error {
	_, exists := s.services[id]
	if !exists {
		return fmt.Errorf("service not found: %d", id)
	}
	delete(s.services, id)
	return nil
}
