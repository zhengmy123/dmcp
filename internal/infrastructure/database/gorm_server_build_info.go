package database

import (
	"context"
	"dynamic_mcp_go_server/internal/domain/model"
	"errors"

	"gorm.io/gorm"
)

type GORMServerBuildInfoDAO struct {
	db *gorm.DB
}

func NewGORMServerBuildInfoDAO(db *gorm.DB) *GORMServerBuildInfoDAO {
	return &GORMServerBuildInfoDAO{db: db}
}

// DB 获取数据库连接
func (d *GORMServerBuildInfoDAO) DB() *gorm.DB {
	return d.db
}

func (d *GORMServerBuildInfoDAO) GetByServerID(ctx context.Context, serverID uint) ([]*model.ServerBuildInfo, error) {
	var infos []*model.ServerBuildInfo
	err := d.db.WithContext(ctx).Where("server_id = ?", serverID).Order("version DESC").Find(&infos).Error
	if err != nil {
		return nil, err
	}
	return infos, nil
}

func (d *GORMServerBuildInfoDAO) GetActiveByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
	var info model.ServerBuildInfo
	err := d.db.WithContext(ctx).Where("server_id = ? AND state = ?", serverID, 1).First(&info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (d *GORMServerBuildInfoDAO) GetByBuildUUID(ctx context.Context, buildUUID string) (*model.ServerBuildInfo, error) {
	var info model.ServerBuildInfo
	err := d.db.WithContext(ctx).Where("build_uuid = ?", buildUUID).First(&info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (d *GORMServerBuildInfoDAO) Save(ctx context.Context, info *model.ServerBuildInfo) error {
	return d.db.WithContext(ctx).Create(info).Error
}

func (d *GORMServerBuildInfoDAO) SaveWithTx(ctx context.Context, tx *gorm.DB, info *model.ServerBuildInfo) error {
	return tx.Create(info).Error
}

func (d *GORMServerBuildInfoDAO) UpdateState(ctx context.Context, id uint, state int) error {
	return d.db.WithContext(ctx).Model(&model.ServerBuildInfo{}).Where("id = ?", id).Update("state", state).Error
}

func (d *GORMServerBuildInfoDAO) GetMaxVersionByServerID(ctx context.Context, serverID uint) (int, error) {
	var maxVersion int
	err := d.db.WithContext(ctx).Model(&model.ServerBuildInfo{}).
		Where("server_id = ?", serverID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	if err != nil {
		return 0, err
	}
	return maxVersion, nil
}
