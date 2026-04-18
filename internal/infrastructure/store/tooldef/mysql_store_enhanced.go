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
	toolDAO *database.GORMToolDAO
}

// NewEnhancedMySQLStore 创建增强的MySQL存储
func NewEnhancedMySQLStore(db *gorm.DB, table string, log logger.Logger) *EnhancedMySQLStore {
	return &EnhancedMySQLStore{
		toolDAO: database.NewGORMToolDAO(db, table),
	}
}

// List 实现Store接口，列出所有工具定义
func (m *EnhancedMySQLStore) List(ctx context.Context) ([]model.ToolDefinition, error) {
	return m.toolDAO.List(ctx)
}
