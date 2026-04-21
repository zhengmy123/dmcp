package tooldef

import (
	"fmt"
	"regexp"
	"strings"

	"dynamic_mcp_go_server/internal/domain/model"

	"github.com/bytedance/sonic"
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
	State       *int                 `json:"state"`
	ServerDesc  string                `json:"server_desc,omitempty"`
}

const (
	ToolNameMinLength = 1
	ToolNameMaxLength = 64
	ToolNamePattern   = `^[a-zA-Z0-9_.-]+$`
)

var toolNameRegex = regexp.MustCompile(ToolNamePattern)

func ValidateToolName(name string) error {
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if len(name) > ToolNameMaxLength {
		return fmt.Errorf("tool name cannot exceed 64 characters")
	}
	if len(name) < ToolNameMinLength {
		return fmt.Errorf("tool name must be at least 1 character")
	}
	if !toolNameRegex.MatchString(name) {
		return fmt.Errorf("tool name can only contain letters, numbers, underscore (_), hyphen (-), and dot (.)")
	}
	return nil
}

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
		if in.State == nil {
			return nil, fmt.Errorf("tool %s: missing required field state", ref)
		}
		if *in.State != 1 {
			return nil, fmt.Errorf("tool %s: state must be 1 (got %d); omit the tool from the JSON array to remove it", ref, *in.State)
		}
		paramsJSON, err := sonic.Marshal(in.Parameters)
		if err != nil {
			return nil, fmt.Errorf("tool %s: marshal parameters: %w", ref, err)
		}
		defs = append(defs, ToolDefinition{
			Name:        name,
			Description: in.Description,
			Parameters:  paramsJSON,
			State:       1,
		})
	}
	return defs, nil
}
