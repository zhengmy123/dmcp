package service

import (
	"context"
	"fmt"

	"dynamic_mcp_go_server/internal/domain/model"
	"dynamic_mcp_go_server/internal/domain/repository"
)

type SystemConfigService struct {
	store repository.SystemConfigStore
}

func NewSystemConfigService(store repository.SystemConfigStore) *SystemConfigService {
	return &SystemConfigService{store: store}
}

func (s *SystemConfigService) GetConfig(ctx context.Context, key string) (*model.SystemConfig, error) {
	if key == "" {
		return nil, fmt.Errorf("config key cannot be empty")
	}
	config, err := s.store.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	return config, nil
}

func (s *SystemConfigService) UpdateConfig(ctx context.Context, key, value string) (*model.SystemConfig, error) {
	if key == "" {
		return nil, fmt.Errorf("config key cannot be empty")
	}
	config := &model.SystemConfig{
		ConfigKey:   key,
		ConfigValue: value,
	}
	if err := s.store.Upsert(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update config: %w", err)
	}
	return config, nil
}
