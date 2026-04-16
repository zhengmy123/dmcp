package tooldef

import (
	"dynamic_mcp_go_server/internal/domain/model"
)

// ParameterType 是 model.ParameterType 的别名
type ParameterType = model.ParameterType

const (
	ParameterTypeString  = model.ParameterTypeString
	ParameterTypeInteger = model.ParameterTypeInteger
	ParameterTypeNumber  = model.ParameterTypeNumber
	ParameterTypeBoolean = model.ParameterTypeBoolean
)

// ParameterDefinition 是 model.ParameterDefinition 的别名
type ParameterDefinition = model.ParameterDefinition

// ToolDefinition 是 model.ToolDefinition 的别名
type ToolDefinition = model.ToolDefinition

// toolDefinitionJSON 工具定义的JSON结构（用于解析）
type toolDefinitionJSON struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Parameters  []ParameterDefinition `json:"parameters"`
	Enabled     *bool                 `json:"enabled"`
	VAuthKey    string                `json:"vauth_key"`
	ServerDesc  string                `json:"server_desc,omitempty"`
}

// ParseToolDefinitions 解析工具定义数组
var ParseToolDefinitions = model.ParseToolDefinitions
