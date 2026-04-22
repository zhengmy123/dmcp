package repository

import (
	"context"

	"dynamic_mcp_go_server/internal/domain/model"
)

type SystemConfigQuery struct {
	ConfigKey *string
}

func (q *SystemConfigQuery) HasCondition() bool {
	return q.ConfigKey != nil
}

// SystemConfigStore 系统配置存储接口
type SystemConfigStore interface {
	GetByKey(ctx context.Context, key string) (*model.SystemConfig, error)
	Upsert(ctx context.Context, config *model.SystemConfig) error
}
