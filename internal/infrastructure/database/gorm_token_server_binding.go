package database

import (
	"context"
	"fmt"

	"dynamic_mcp_go_server/internal/domain/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GORMTokenServerBindingDAO GORM实现的令牌服务器绑定DAO
type GORMTokenServerBindingDAO struct {
	db *gorm.DB
}

// NewGORMTokenServerBindingDAO 创建GORM令牌服务器绑定DAO
func NewGORMTokenServerBindingDAO(db *gorm.DB) *GORMTokenServerBindingDAO {
	return &GORMTokenServerBindingDAO{db: db}
}

// ListByTokenID 根据令牌ID获取所有绑定
func (d *GORMTokenServerBindingDAO) ListByTokenID(ctx context.Context, tokenID uint) ([]*model.TokenServerBinding, error) {
	var bindings []*model.TokenServerBinding

	result := d.db.WithContext(ctx).Where("token_id = ?", tokenID).Find(&bindings)
	if result.Error != nil {
		return nil, result.Error
	}

	return bindings, nil
}

// ListByServerID 根据服务器ID获取所有绑定
func (d *GORMTokenServerBindingDAO) ListByServerID(ctx context.Context, serverID uint) ([]*model.TokenServerBinding, error) {
	var bindings []*model.TokenServerBinding

	result := d.db.WithContext(ctx).Where("server_id = ?", serverID).Find(&bindings)
	if result.Error != nil {
		return nil, result.Error
	}

	return bindings, nil
}

// Save 保存绑定（创建或更新）
func (d *GORMTokenServerBindingDAO) Save(ctx context.Context, binding *model.TokenServerBinding) error {
	return d.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "token_id"}, {Name: "server_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"created_at"}),
	}).Create(binding).Error
}

// DeleteByTokenID 根据令牌ID删除所有绑定
func (d *GORMTokenServerBindingDAO) DeleteByTokenID(ctx context.Context, tokenID uint) error {
	result := d.db.WithContext(ctx).Where("token_id = ?", tokenID).Delete(&model.TokenServerBinding{})
	if result.Error != nil {
		return fmt.Errorf("delete token bindings failed: %w", result.Error)
	}
	return nil
}

// ReplaceByTokenID 替换令牌的所有服务器绑定
func (d *GORMTokenServerBindingDAO) ReplaceByTokenID(ctx context.Context, tokenID uint, serverIDs []uint) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除现有绑定
		if err := tx.Where("token_id = ?", tokenID).Delete(&model.TokenServerBinding{}).Error; err != nil {
			return fmt.Errorf("delete existing bindings failed: %w", err)
		}

		// 如果没有新绑定，直接返回
		if len(serverIDs) == 0 {
			return nil
		}

		// 创建新绑定
		bindings := make([]*model.TokenServerBinding, len(serverIDs))
		for i, serverID := range serverIDs {
			bindings[i] = &model.TokenServerBinding{
				TokenID:  tokenID,
				ServerID: serverID,
			}
		}

		return tx.Create(&bindings).Error
	})
}
