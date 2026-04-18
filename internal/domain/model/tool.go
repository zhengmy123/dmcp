package model

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bytedance/sonic"
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
	Enabled       bool      `json:"enabled" gorm:"default:true"`
	OutputMapping []byte    `json:"output_mapping" gorm:"type:text"`
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

// toolDefinitionJSON 用于严格解析 JSON，要求 enabled 显式为 true
type toolDefinitionJSON struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Parameters  []ParameterDefinition `json:"parameters"`
	Enabled     *bool                 `json:"enabled"`
}

// ParseToolDefinitions 解析工具定义数组
func ParseToolDefinitions(raw []byte) ([]ToolDefinition, error) {
	var inputs []toolDefinitionJSON
	if err := sonic.Unmarshal(raw, &inputs); err != nil {
		return nil, err
	}
	defs := make([]ToolDefinition, 0, len(inputs))
	for i, in := range inputs {
		ref := fmt.Sprintf("index %d", i)
		name := strings.TrimSpace(in.Name)
		if name == "" {
			return nil, fmt.Errorf("tool definition %s: missing or empty required field name", ref)
		}
		ref = fmt.Sprintf("%q", name)
		if in.Enabled == nil {
			return nil, fmt.Errorf("tool %s: missing required field enabled", ref)
		}
		if !*in.Enabled {
			return nil, fmt.Errorf("tool %s: enabled must be true (got false); omit the tool from the JSON array to remove it", ref)
		}
		paramsJSON, err := sonic.Marshal(in.Parameters)
		if err != nil {
			return nil, fmt.Errorf("tool %s: marshal parameters: %w", ref, err)
		}
		defs = append(defs, ToolDefinition{
			Name:        name,
			Description: in.Description,
			Parameters:  paramsJSON,
			Enabled:     true,
		})
	}
	return defs, nil
}

const (
	ToolNameMinLength = 1
	ToolNameMaxLength = 64
	ToolNamePattern   = `^[a-zA-Z0-9_.-]+$`
)

var toolNameRegex = regexp.MustCompile(ToolNamePattern)

func ValidateToolName(name string) error {
	if name == "" {
		return errors.New("tool name cannot be empty")
	}
	if len(name) > ToolNameMaxLength {
		return errors.New("tool name cannot exceed 64 characters")
	}
	if len(name) < ToolNameMinLength {
		return errors.New("tool name must be at least 1 character")
	}
	if !toolNameRegex.MatchString(name) {
		return errors.New("tool name can only contain letters, numbers, underscore (_), hyphen (-), and dot (.)")
	}
	return nil
}
