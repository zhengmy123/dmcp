package httpservice

import (
	"context"
	"fmt"
	"strings"
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

func (s *MemoryServiceStore) ListWithQuery(ctx context.Context, query *model.ServiceQuery) ([]*model.HTTPService, int64, error) {
	services := make([]*model.HTTPService, 0)
	for _, service := range s.services {
		// 名称筛选
		if query.Name != nil && *query.Name != "" {
			if !contains(service.Name, *query.Name) {
				continue
			}
		}
		// 状态筛选
		if query.State != nil {
			if service.State != *query.State {
				continue
			}
		}
		services = append(services, service)
	}

	// 统计总数
	total := int64(len(services))

	// 分页
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(services) {
		return []*model.HTTPService{}, total, nil
	}
	if end > len(services) {
		end = len(services)
	}

	return services[start:end], total, nil
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
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
