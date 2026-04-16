package database

import (
	"context"
	"fmt"

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

// ListByServiceID 获取某服务下的所有工具
func (d *GORMToolStore) ListByServiceID(ctx context.Context, serviceID uint) ([]*model.ToolDefinition, error) {
	var tools []*model.ToolDefinition

	result := d.db.WithContext(ctx).Where("service_id = ? AND enabled = ?", serviceID, true).Find(&tools)
	if result.Error != nil {
		return nil, result.Error
	}

	return tools, nil
}

// ListByVAuthKey 获取某服务器下的所有工具
func (d *GORMToolStore) ListByVAuthKey(ctx context.Context, vauthKey string) ([]*model.ToolDefinition, error) {
	var tools []*model.ToolDefinition

	result := d.db.WithContext(ctx).
		Where("vauth_key = ? AND enabled = ?", vauthKey, true).
		Order("created_at DESC").
		Find(&tools)
	if result.Error != nil {
		return nil, result.Error
	}

	return tools, nil
}

// AddToolWithVAuthKey 添加带 vauth_key 的工具
func (d *GORMToolStore) AddToolWithVAuthKey(ctx context.Context, tool *model.ToolDefinition) error {
	return d.db.WithContext(ctx).Create(tool).Error
}

// RemoveToolByNameAndVAuthKey 按名称和 vauth_key 移除工具
func (d *GORMToolStore) RemoveToolByNameAndVAuthKey(ctx context.Context, name, vauthKey string) error {
	result := d.db.WithContext(ctx).
		Model(&model.ToolDefinition{}).
		Where("vauth_key = ? AND name = ?", vauthKey, name).
		Update("enabled", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("tool not found: %s", name)
	}
	return nil
}

// AddToolsToServer 向服务器添加工具（通过 vauth_key 关联）
func (d *GORMToolStore) AddToolsToServer(ctx context.Context, serverID uint, toolNames []string) error {
	if len(toolNames) == 0 {
		return nil
	}

	result := d.db.WithContext(ctx).Model(&model.ToolDefinition{}).
		Where("name IN ? AND enabled = ?", toolNames, true).
		Update("service_id", serverID)
	if result.Error != nil {
		return fmt.Errorf("add tools to server failed: %w", result.Error)
	}

	return nil
}

// RemoveToolFromServer 从服务器移除工具
func (d *GORMToolStore) RemoveToolFromServer(ctx context.Context, serverID uint, toolName string) error {
	result := d.db.WithContext(ctx).Model(&model.ToolDefinition{}).
		Where("service_id = ? AND name = ?", serverID, toolName).
		Update("service_id", 0)
	if result.Error != nil {
		return fmt.Errorf("remove tool from server failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("tool not found on server: %s", toolName)
	}

	return nil
}

// SaveTool 保存工具定义（创建或更新）
func (d *GORMToolStore) SaveTool(ctx context.Context, tool *model.ToolDefinition) error {
	// 如果没有ID，创建新记录
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
		return fmt.Errorf("delete tool failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("tool not found: %d", id)
	}
	return nil
}

// GetByNameAndServer 根据名称和服务器获取工具
func (d *GORMToolStore) GetByNameAndServer(ctx context.Context, name, vauthKey string) (*model.ToolDefinition, error) {
	var tool model.ToolDefinition
	err := d.db.WithContext(ctx).
		Where("name = ? AND vauth_key = ? AND enabled = ?", name, vauthKey, true).
		First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}

// Save 保存工具
func (d *GORMToolStore) Save(ctx context.Context, tool *model.ToolDefinition) error {
	return d.db.WithContext(ctx).Save(tool).Error
}
