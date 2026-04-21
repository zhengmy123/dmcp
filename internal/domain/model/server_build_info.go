package model

import (
	"time"
)

type ServerBuildInfo struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ServerID  uint      `json:"server_id" gorm:"not null;index:idx_server_state"`
	Version   int       `json:"version" gorm:"not null;default:1"`
	BuildUUID string    `json:"build_uuid" gorm:"size:36;not null;uniqueIndex"`
	Hash      string    `json:"hash" gorm:"size:64;not null;index"`
	BuildData string    `json:"build_data" gorm:"type:text"`
	State     int       `json:"state" gorm:"not null;default:1;index:idx_server_state"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ServerBuildInfo) TableName() string {
	return "server_build_info"
}

type BuildData struct {
	Tools        []ToolSnapshot        `json:"tools"`
	HTTPServices []HTTPServiceSnapshot `json:"http_services"`
}

type ToolSnapshot struct {
	ID            uint                   `json:"id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	ServiceID     uint                   `json:"service_id"`
	Parameters    []ParameterDefinition  `json:"parameters"`
	InputMapping  []InputMappingField    `json:"input_mapping"`
	OutputMapping *OutputMappingConfig    `json:"output_mapping"`
	State         int                    `json:"state"`
}

type HTTPServiceSnapshot struct {
	ID            uint            `json:"id"`
	Name          string          `json:"name"`
	TargetURL     string          `json:"target_url"`
	Method        string          `json:"method"`
	Headers       JSONBytes       `json:"headers"`
	BodyType      string          `json:"body_type"`
	Timeout       int             `json:"timeout_seconds"`
	InputSchema   JSONBytes       `json:"input_schema"`
	OutputSchema  JSONBytes       `json:"output_schema"`
}
