package systemconfig

import (
	"context"
	"fmt"

	"dynamic_mcp_go_server/internal/domain/model"

	"gorm.io/gorm"
)

type GORMSystemConfigStore struct {
	db *gorm.DB
}

func NewGORMSystemConfigStore(db *gorm.DB) *GORMSystemConfigStore {
	return &GORMSystemConfigStore{db: db}
}

func (s *GORMSystemConfigStore) GetByKey(ctx context.Context, key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	result := s.db.WithContext(ctx).Where("config_key = ?", key).First(&config)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("query system config: %w", result.Error)
	}
	return &config, nil
}

func (s *GORMSystemConfigStore) Upsert(ctx context.Context, config *model.SystemConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.ConfigKey == "" {
		return fmt.Errorf("config_key cannot be empty")
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var existing model.SystemConfig
	err := tx.Where("config_key = ?", config.ConfigKey).First(&existing).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("query existing config: %w", err)
	}

	if err == gorm.ErrRecordNotFound {
		if err := tx.Create(config).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("create config: %w", err)
		}
		return tx.Commit().Error
	}

	existing.ConfigValue = config.ConfigValue
	if err := tx.Save(&existing).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("update config: %w", err)
	}

	return tx.Commit().Error
}
