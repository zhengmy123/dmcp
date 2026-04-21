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

func (d *GORMToolServerBindingDAO) DB() *gorm.DB {
	return d.db
}

func (d *GORMToolServerBindingDAO) ListByToolID(ctx context.Context, toolID uint) ([]*model.ToolServerBinding, error) {
	var bindings []*model.ToolServerBinding
	err := d.db.WithContext(ctx).Where("tool_id = ? AND state = ?", toolID, 1).Find(&bindings).Error
	return bindings, err
}

func (d *GORMToolServerBindingDAO) ListByServerID(ctx context.Context, serverID uint) ([]*model.ToolServerBinding, error) {
	var bindings []*model.ToolServerBinding
	err := d.db.WithContext(ctx).Where("server_id = ? AND state = ?", serverID, 1).Find(&bindings).Error
	return bindings, err
}

func (d *GORMToolServerBindingDAO) ListByToolIDs(ctx context.Context, toolIDs []uint) ([]*model.ToolServerBinding, error) {
	if len(toolIDs) == 0 {
		return nil, nil
	}
	var bindings []*model.ToolServerBinding
	err := d.db.WithContext(ctx).Where("tool_id IN ? AND state = ?", toolIDs, 1).Find(&bindings).Error
	return bindings, err
}

func (d *GORMToolServerBindingDAO) ListByServerIDs(ctx context.Context, serverIDs []uint) ([]*model.ToolServerBinding, error) {
	if len(serverIDs) == 0 {
		return nil, nil
	}
	var bindings []*model.ToolServerBinding
	err := d.db.WithContext(ctx).Where("server_id IN ? AND state = ?", serverIDs, 1).Find(&bindings).Error
	return bindings, err
}

func (d *GORMToolServerBindingDAO) ListAllIncludeDeleted(ctx context.Context) ([]*model.ToolServerBinding, error) {
	var bindings []*model.ToolServerBinding
	err := d.db.WithContext(ctx).Unscoped().Find(&bindings).Error
	return bindings, err
}

func (d *GORMToolServerBindingDAO) GetByToolAndServer(ctx context.Context, toolID, serverID uint) (*model.ToolServerBinding, error) {
	var binding model.ToolServerBinding
	err := d.db.WithContext(ctx).Where("tool_id = ? AND server_id = ? AND state = ?", toolID, serverID, 1).First(&binding).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

func (d *GORMToolServerBindingDAO) ExistByToolAndServer(ctx context.Context, toolID, serverID uint) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&model.ToolServerBinding{}).Where("tool_id = ? AND server_id = ?", toolID, serverID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (d *GORMToolServerBindingDAO) GetByToolAndServerIncludeDeleted(ctx context.Context, toolID, serverID uint) (*model.ToolServerBinding, error) {
	var binding model.ToolServerBinding
	err := d.db.WithContext(ctx).Unscoped().Where("tool_id = ? AND server_id = ?", toolID, serverID).First(&binding).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

func (d *GORMToolServerBindingDAO) Restore(ctx context.Context, id uint) error {
	return d.db.WithContext(ctx).Model(&model.ToolServerBinding{}).Where("id = ?", id).Update("state", 1).Error
}

func (d *GORMToolServerBindingDAO) BatchRestore(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Model(&model.ToolServerBinding{}).Where("id IN ?", ids).Update("state", 1).Error
}

func (d *GORMToolServerBindingDAO) Save(ctx context.Context, binding *model.ToolServerBinding) error {
	return d.db.WithContext(ctx).Create(binding).Error
}

func (d *GORMToolServerBindingDAO) Delete(ctx context.Context, id uint) error {
	return d.db.WithContext(ctx).Model(&model.ToolServerBinding{}).Where("id = ?", id).Update("state", 0).Error
}

func (d *GORMToolServerBindingDAO) DeleteByToolID(ctx context.Context, toolID uint) error {
	return d.db.WithContext(ctx).Model(&model.ToolServerBinding{}).Where("tool_id = ? AND state = ?", toolID, 1).Update("state", 0).Error
}

func (d *GORMToolServerBindingDAO) DeleteByServerID(ctx context.Context, serverID uint) error {
	return d.db.WithContext(ctx).Model(&model.ToolServerBinding{}).Where("server_id = ? AND state = ?", serverID, 1).Update("state", 0).Error
}

func (d *GORMToolServerBindingDAO) ReplaceByToolID(ctx context.Context, toolID uint, serverIDs []uint) error {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Model(&model.ToolServerBinding{}).Where("tool_id = ? AND state = ?", toolID, 1).Update("state", 0).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, serverID := range serverIDs {
		binding := &model.ToolServerBinding{
			ToolID:   toolID,
			ServerID: serverID,
		}
		if err := tx.Create(binding).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
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
	return d.db.WithContext(ctx).Model(&model.ToolServerBinding{}).Where("id IN ? AND state = ?", ids, 1).Update("state", 0).Error
}