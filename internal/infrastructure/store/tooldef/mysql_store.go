package tooldef

import (
	"context"
	"fmt"
	"regexp"

	"dynamic_mcp_go_server/internal/domain/model"

	"gorm.io/gorm"
)

var mysqlTableNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// GORMToolStore GORM存储实现（旧版兼容）
type GORMToolStore struct {
	db    *gorm.DB
	table string
}

// NewGORMToolStore 创建GORM工具存储
func NewGORMToolStore(db *gorm.DB, table string) *GORMToolStore {
	return &GORMToolStore{
		db:    db,
		table: table,
	}
}

// List 获取所有启用的工具定义
func (m *GORMToolStore) List(ctx context.Context) ([]model.ToolDefinition, error) {
	if !mysqlTableNamePattern.MatchString(m.table) {
		return nil, fmt.Errorf("invalid table name %q", m.table)
	}

	type toolRow struct {
		Name        string `gorm:"column:name"`
		Description string `gorm:"column:description"`
		Parameters  []byte `gorm:"column:parameters"`
	}

	var rows []toolRow
	result := m.db.WithContext(ctx).Table(m.table).
		Select("name, description, parameters").
		Where("enabled = ?", true).
		Order("updated_at DESC").
		Find(&rows)
	if result.Error != nil {
		return nil, fmt.Errorf("query gorm tool definitions: %w", result.Error)
	}

	defs := make([]model.ToolDefinition, 0, len(rows))
	for _, r := range rows {
		defs = append(defs, model.ToolDefinition{
			Name:        r.Name,
			Description: r.Description,
			Parameters:  nil,
			Enabled:     true,
		})
	}
	return defs, nil
}
