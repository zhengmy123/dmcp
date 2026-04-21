package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"

	"gorm.io/gorm"
)

type ServerBuildInfoStore interface {
	DB() *gorm.DB
	GetByServerID(ctx context.Context, serverID uint) ([]*model.ServerBuildInfo, error)
	GetActiveByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error)
	GetByBuildUUID(ctx context.Context, buildUUID string) (*model.ServerBuildInfo, error)
	Save(ctx context.Context, info *model.ServerBuildInfo) error
	SaveWithTx(ctx context.Context, tx *gorm.DB, info *model.ServerBuildInfo) error
	UpdateState(ctx context.Context, id uint, state int) error
	GetMaxVersionByServerID(ctx context.Context, serverID uint) (int, error)
}
