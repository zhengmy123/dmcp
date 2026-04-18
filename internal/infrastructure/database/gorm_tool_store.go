package database

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"

	"gorm.io/gorm"
)

// GORMToolStore GORM工具存储实现（实现repository.ToolStore接口）
type GORMToolStore struct {
	db *gorm.DB
}

// NewGORMToolStore 创建GORM工具存储
func NewGORMToolStore(db *gorm.DB) *GORMToolStore {
	return &GORMToolStore{db: db}
}

// List 获取所有启用的工具定义
func (d *GORMToolStore) List(ctx context.Context) ([]*model.ToolDefinition, error) {
	var tools []*model.ToolDefinition

	result := d.db.WithContext(ctx).Where("enabled = ?", true).Find(&tools)
	if result.Error != nil {
		return nil, result.Error
	}

	return tools, nil
}

// SaveTool 保存工具定义（创建或更新）
func (d *GORMToolStore) SaveTool(ctx context.Context, tool *model.ToolDefinition) error {
	if tool.ID == 0 {
		return d.db.WithContext(ctx).Create(tool).Error
	}

	var existing model.ToolDefinition
	result := d.db.WithContext(ctx).Where("id = ?", tool.ID).First(&existing)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return d.db.WithContext(ctx).Create(tool).Error
		}
		return result.Error
	}

	return d.db.WithContext(ctx).Model(&model.ToolDefinition{}).Where("id = ?", tool.ID).Updates(tool).Error
}

// DeleteTool 删除工具定义
func (d *GORMToolStore) DeleteTool(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Delete(&model.ToolDefinition{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return nil
	}
	return nil
}

// GetByID 根据ID获取工具
func (d *GORMToolStore) GetByID(ctx context.Context, id uint) (*model.ToolDefinition, error) {
	var tool model.ToolDefinition
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}

// Save 保存工具
func (d *GORMToolStore) Save(ctx context.Context, tool *model.ToolDefinition) error {
	return d.db.WithContext(ctx).Save(tool).Error
}
