package model

import (
	"encoding/json"
	"time"
)

// ServiceMapping 服务映射定义
type ServiceMapping struct {
	ID            uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	ServiceID     uint            `json:"service_id" gorm:"not null;index:idx_service_vauth,unique"`
	ServiceName   string          `json:"service_name" gorm:"size:128;not null"`
	VAuthKey      string          `json:"vauth_key" gorm:"size:128;not null;index:idx_service_vauth,unique"`
	JSONSchema    json.RawMessage `json:"json_schema" gorm:"type:text"`
	SchemaHash    string          `json:"schema_hash" gorm:"size:64;index"`
	MappingConfig json.RawMessage `json:"mapping_config" gorm:"type:text"`
	Enabled       bool            `json:"enabled" gorm:"default:true;index"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime;index"`
}

func (ServiceMapping) TableName() string {
	return "mcp_service_mappings"
}

// NewServiceMapping 创建新的服务映射
func NewServiceMapping(serviceID uint, serviceName, vauthKey string, jsonSchema []byte) *ServiceMapping {
	now := time.Now()
	return &ServiceMapping{
		ServiceID:   serviceID,
		ServiceName: serviceName,
		VAuthKey:    vauthKey,
		JSONSchema:  jsonSchema,
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsValid 检查服务映射是否有效
func (m *ServiceMapping) IsValid() bool {
	return m.ServiceID > 0 && m.VAuthKey != "" && len(m.JSONSchema) > 0
}

// HasMappingConfig 检查是否有映射配置
func (m *ServiceMapping) HasMappingConfig() bool {
	return len(m.MappingConfig) > 0
}
