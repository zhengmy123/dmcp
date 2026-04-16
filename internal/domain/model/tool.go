package model

import (
	"encoding/json"
	"fmt"
	"strings"
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
// VAuthKey: 将多个不同的 HTTP 接口打包成一个 MCP Server 的聚合键
// ServerDesc: MCP Server 的描述信息
type ToolDefinition struct {
	Name        string                `json:"name" gorm:"size:128"`
	Description string                `json:"description" gorm:"type:text"`
	Parameters  []ParameterDefinition `json:"parameters" gorm:"-"`
	Enabled     bool                  `json:"enabled" gorm:"default:true"`
	VAuthKey    string                `json:"vauth_key" gorm:"size:128;index"`
	ServerDesc  string                `json:"server_desc" gorm:"size:512"`
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
	VAuthKey    string                `json:"vauth_key"`
	ServerDesc  string                `json:"server_desc,omitempty"`
}

// ParseToolDefinitions 解析工具定义数组
func ParseToolDefinitions(raw []byte) ([]ToolDefinition, error) {
	var inputs []toolDefinitionJSON
	if err := json.Unmarshal(raw, &inputs); err != nil {
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
		if strings.TrimSpace(in.VAuthKey) == "" {
			return nil, fmt.Errorf("tool %s: missing required field vauth_key", ref)
		}
		defs = append(defs, ToolDefinition{
			Name:        name,
			Description: in.Description,
			Parameters:  in.Parameters,
			Enabled:     true,
			VAuthKey:    strings.TrimSpace(in.VAuthKey),
			ServerDesc:  in.ServerDesc,
		})
	}
	return defs, nil
}
