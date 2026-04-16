package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// CommonDAO 通用DAO接口
type CommonDAO interface {
	// Exists 检查记录是否存在
	Exists(ctx context.Context, id uint) (bool, error)
	// Count 获取记录总数
	Count(ctx context.Context) (int64, error)
}

// BaseDAO 基础DAO结构，提供通用方法
type BaseDAO struct {
	db    *gorm.DB
	table string
}

// NewBaseDAO 创建基础DAO
func NewBaseDAO(db *gorm.DB, table string) *BaseDAO {
	return &BaseDAO{
		db:    db,
		table: table,
	}
}

// Exists 检查记录是否存在
func (b *BaseDAO) Exists(ctx context.Context, id uint) (bool, error) {
	var count int64
	result := b.db.WithContext(ctx).Table(b.table).Where("id = ?", id).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// Count 获取记录总数
func (b *BaseDAO) Count(ctx context.Context) (int64, error) {
	var count int64
	result := b.db.WithContext(ctx).Table(b.table).Count(&count)
	return count, result.Error
}

// SoftDelete 软删除记录
func (b *BaseDAO) SoftDelete(ctx context.Context, id uint) error {
	result := b.db.WithContext(ctx).Table(b.table).Where("id = ?", id).Updates(map[string]interface{}{
		"enabled":    false,
		"updated_at": gorm.Expr("NOW()"),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found: %d", id)
	}
	return nil
}

// HardDelete 硬删除记录
func (b *BaseDAO) HardDelete(ctx context.Context, id uint) error {
	result := b.db.WithContext(ctx).Table(b.table).Where("id = ?", id).Delete(nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found: %d", id)
	}
	return nil
}
