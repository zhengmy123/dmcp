package model

import (
	"time"
)

// ParameterType 参数类型
type ParameterType string

const (
	ParameterTypeString  ParameterType = "string"
	ParameterTypeInteger ParameterType = "integer"
	ParameterTypeNumber  ParameterType = "number"
	ParameterTypeBoolean ParameterType = "boolean"
)

// ParameterDefinition 参数定义
type ParameterDefinition struct {
	Name        string        `json:"name"`
	Type        ParameterType `json:"type"`
	Required    bool          `json:"required"`
	Description string        `json:"description,omitempty"`
	Default     any           `json:"default,omitempty"`
	Enum        []any         `json:"enum,omitempty"`
	Minimum     *float64      `json:"minimum,omitempty"`
	Maximum     *float64      `json:"maximum,omitempty"`
}

// ToolDefinition 动态工具定义
// VAuthKey: 运行时通过 tool_mcp_server_bindings + mcp_servers 关联获取，不存储在DB
type ToolDefinition struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name          string    `json:"name" gorm:"size:128;uniqueIndex"`
	Description   string    `json:"description" gorm:"type:text"`
	ServiceID     uint      `json:"service_id" gorm:"index"`
	Parameters    []byte    `json:"parameters" gorm:"type:text"`
	InputMapping  []byte    `json:"input_mapping" gorm:"type:text"`
	OutputMapping []byte    `json:"output_mapping" gorm:"type:text"`
	State         int       `json:"state" gorm:"default:1"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// InputExtraFields 入参扩展字段定义
type InputExtraField struct {
	Name        string        `json:"name"`
	Type        ParameterType `json:"type"`
	Description string        `json:"description,omitempty"`
	Required    bool          `json:"required"`
}

// InputExtraConfig 入参扩展配置
type InputExtraConfig struct {
	ExtraFields []InputExtraField `json:"extra_fields"`
}

// OutputMappingField 出参映射字段
type OutputMappingField struct {
	SourceField  string `json:"source_field"` // 源字段（来自HTTP服务OutputSchema）
	TargetField  string `json:"target_field"` // 目标字段名（MCP工具返回）
	ValueType    string `json:"value_type"`
	DefaultValue string `json:"default_value"`
	Description  string `json:"description,omitempty"`
}

// InputMappingField 入参映射字段
type InputMappingField struct {
	Source      string `json:"source"` // 源字段（来自MCP工具入参）
	Target      string `json:"target"` // 目标字段名（HTTP服务InputSchema）
	Description string `json:"description,omitempty"`
}

// OutputMappingConfig 出参映射配置
type OutputMappingConfig struct {
	Fields []OutputMappingField `json:"fields"`
}

func (ToolDefinition) TableName() string {
	return "mcp_tool_definitions"
}
