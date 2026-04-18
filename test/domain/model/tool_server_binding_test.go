package model_test

import (
	"testing"

	"dynamic_mcp_go_server/internal/domain/model"
)

func TestToolServerBinding_TableName(t *testing.T) {
	binding := model.ToolServerBinding{}
	if binding.TableName() != "tool_mcp_server_bindings" {
		t.Errorf("expected table name 'tool_mcp_server_bindings', got '%s'", binding.TableName())
	}
}