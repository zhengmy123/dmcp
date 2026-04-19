package model_test

import (
	"testing"

	"dynamic_mcp_go_server/internal/domain/model"
)

func TestServerBuildInfo_TableName(t *testing.T) {
	info := model.ServerBuildInfo{}
	if info.TableName() != "server_build_info" {
		t.Errorf("expected table name 'server_build_info', got '%s'", info.TableName())
	}
}

func TestBuildData_Structure(t *testing.T) {
	data := model.BuildData{
		Tools: []model.ToolSnapshot{
			{ID: 1, Name: "test_tool", Description: "test", Enabled: true},
		},
		HTTPServices: []model.HTTPServiceSnapshot{
			{ID: 1, Name: "test_service", TargetURL: "http://test.com"},
		},
	}
	if len(data.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(data.Tools))
	}
	if len(data.HTTPServices) != 1 {
		t.Errorf("expected 1 http service, got %d", len(data.HTTPServices))
	}
}
