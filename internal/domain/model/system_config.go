package model

import (
	"time"
)

type SystemConfig struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	ConfigKey   string    `json:"config_key" gorm:"column:config_key;type:varchar(64);not null;uniqueIndex"`
	ConfigValue string    `json:"config_value" gorm:"column:config_value;type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
