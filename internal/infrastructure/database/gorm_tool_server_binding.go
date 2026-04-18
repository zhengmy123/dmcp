package database

import (
	"context"
	"dynamic_mcp_go_server/internal/domain/model"
	"errors"

	"gorm.io/gorm"
)

type GORMToolServerBindingDAO struct {
	db *gorm.DB
}

func NewGORMToolServerBindingDAO(db *gorm.DB) *GORMToolServerBindingDAO {
	return &GORMToolServerBindingDAO{db: db}
}

func (d *GORMToolServerBindingDAO) ListByToolID(ctx context.Context, toolID uint) ([]*model.ToolServerBinding, error) {
	var bindings []*model.ToolServerBinding
	err := d.db.WithContext(ctx).Where("tool_id = ?", toolID).Find(&bindings).Error
	return bindings, err
}

func (d *GORMToolServerBindingDAO) ListByServerID(ctx context.Context, serverID uint) ([]*model.ToolServerBinding, error) {
	var bindings []*model.ToolServerBinding
	err := d.db.WithContext(ctx).Where("server_id = ?", serverID).Find(&bindings).Error
	return bindings, err
}

func (d *GORMToolServerBindingDAO) GetByToolAndServer(ctx context.Context, toolID, serverID uint) (*model.ToolServerBinding, error) {
	var binding model.ToolServerBinding
	err := d.db.WithContext(ctx).Where("tool_id = ? AND server_id = ?", toolID, serverID).First(&binding).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

func (d *GORMToolServerBindingDAO) Save(ctx context.Context, binding *model.ToolServerBinding) error {
	return d.db.WithContext(ctx).Create(binding).Error
}

func (d *GORMToolServerBindingDAO) Delete(ctx context.Context, id uint) error {
	return d.db.WithContext(ctx).Delete(&model.ToolServerBinding{}, "id = ?", id).Error
}

func (d *GORMToolServerBindingDAO) DeleteByToolID(ctx context.Context, toolID uint) error {
	return d.db.WithContext(ctx).Where("tool_id = ?", toolID).Delete(&model.ToolServerBinding{}).Error
}

func (d *GORMToolServerBindingDAO) DeleteByServerID(ctx context.Context, serverID uint) error {
	return d.db.WithContext(ctx).Where("server_id = ?", serverID).Delete(&model.ToolServerBinding{}).Error
}

func (d *GORMToolServerBindingDAO) ReplaceByToolID(ctx context.Context, toolID uint, serverIDs []uint) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tool_id = ?", toolID).Delete(&model.ToolServerBinding{}).Error; err != nil {
			return err
		}
		for _, serverID := range serverIDs {
			binding := &model.ToolServerBinding{
				ToolID:   toolID,
				ServerID: serverID,
			}
			if err := tx.Create(binding).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *GORMToolServerBindingDAO) BatchSave(ctx context.Context, bindings []*model.ToolServerBinding) error {
	if len(bindings) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Create(bindings).Error
}

func (d *GORMToolServerBindingDAO) BatchDelete(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Delete(&model.ToolServerBinding{}, "id IN ?", ids).Error
}