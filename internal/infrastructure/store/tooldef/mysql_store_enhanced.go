package tooldef

import (
	"context"

	"dynamic_mcp_go_server/internal/common/logger"
	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/infrastructure/database"

	"gorm.io/gorm"
)

// EnhancedMySQLStore 增强的MySQL存储，支持JSON Schema校验
type EnhancedMySQLStore struct {
	dao     *database.GORMMappingDAO
	toolDAO *database.GORMToolDAO
}

// NewEnhancedMySQLStore 创建增强的MySQL存储
func NewEnhancedMySQLStore(db *gorm.DB, table string, log logger.Logger) *EnhancedMySQLStore {
	return &EnhancedMySQLStore{
		dao:     database.NewGORMMappingDAO(db, log),
		toolDAO: database.NewGORMToolDAO(db, table),
	}
}

// List 实现Store接口，列出所有工具定义
func (m *EnhancedMySQLStore) List(ctx context.Context) ([]model.ToolDefinition, error) {
	return m.toolDAO.List(ctx)
}

// StoreServiceMapping 存储服务映射
func (m *EnhancedMySQLStore) StoreServiceMapping(ctx context.Context, mapping *model.ServiceMapping) error {
	existing, err := m.dao.Get(ctx, mapping.ServiceID, mapping.VAuthKey)
	if err != nil {
		return err
	}

	if existing != nil {
		mapping.ID = existing.ID
		return m.dao.Update(ctx, mapping)
	}

	return m.dao.Create(ctx, mapping)
}

// GetServiceMapping 获取服务映射
func (m *EnhancedMySQLStore) GetServiceMapping(ctx context.Context, serviceID uint, vauthKey string) (*model.ServiceMapping, error) {
	return m.dao.Get(ctx, serviceID, vauthKey)
}

// ListServiceMappings 列出所有服务映射
func (m *EnhancedMySQLStore) ListServiceMappings(ctx context.Context) ([]*model.ServiceMapping, error) {
	return m.dao.List(ctx)
}

// ValidateServiceMapping 验证服务映射是否有效
func (m *EnhancedMySQLStore) ValidateServiceMapping(ctx context.Context, serviceID uint, vauthKey string) (bool, string, error) {
	return m.dao.Validate(ctx, serviceID, vauthKey)
}
