package database

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"

	"gorm.io/gorm"
)

type GORMToolStore struct {
	db *gorm.DB
}

func NewGORMToolStore(db *gorm.DB) *GORMToolStore {
	return &GORMToolStore{db: db}
}

func (d *GORMToolStore) GetByID(ctx context.Context, id uint) (*model.ToolDefinition, error) {
	var tool model.ToolDefinition
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}

func (d *GORMToolStore) GetByName(ctx context.Context, name string) (*model.ToolDefinition, error) {
	var tool model.ToolDefinition
	err := d.db.WithContext(ctx).Where("name = ?", name).First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}

func (d *GORMToolStore) List(ctx context.Context, query *repository.ToolQuery, page, pageSize int) ([]*model.ToolDefinition, int64, error) {
	var tools []*model.ToolDefinition
	var total int64

	db := d.db.WithContext(ctx).Model(&model.ToolDefinition{})

	if query != nil {
		if query.ID != nil {
			db = db.Where("id = ?", *query.ID)
		}
		if query.Name != nil {
			db = db.Where("name = ?", *query.Name)
		}
		if query.ServiceID != nil {
			db = db.Where("service_id = ?", *query.ServiceID)
		}
		if query.Enabled != nil {
			db = db.Where("enabled = ?", *query.Enabled)
		}
		if query.Keyword != nil && *query.Keyword != "" {
			db = db.Where("name LIKE ?", "%"+*query.Keyword+"%")
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&tools).Error; err != nil {
		return nil, 0, err
	}

	return tools, total, nil
}

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

func (d *GORMToolStore) DeleteTool(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Delete(&model.ToolDefinition{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *GORMToolStore) Delete(ctx context.Context, id uint) error {
	result := d.db.WithContext(ctx).Model(&model.ToolDefinition{}).Where("id = ?", id).Update("enabled", false)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *GORMToolStore) Create(ctx context.Context, tool *model.ToolDefinition) error {
	return d.db.WithContext(ctx).Create(tool).Error
}

func (d *GORMToolStore) Update(ctx context.Context, tool *model.ToolDefinition) error {
	return d.db.WithContext(ctx).Save(tool).Error
}
