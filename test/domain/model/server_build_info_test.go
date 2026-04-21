package model_test

import (
	"encoding/json"
	"testing"

	"dynamic_mcp_go_server/internal/domain/model"

	"github.com/bytedance/sonic"
)

func TestServerBuildInfo_TableName(t *testing.T) {
	info := model.ServerBuildInfo{}
	if info.TableName() != "server_build_info" {
		t.Errorf("expected table name 'server_build_info', got '%s'", info.TableName())
	}
}

func TestBuildData_ParametersFormatted(t *testing.T) {
	params := []model.ParameterDefinition{
		{Name: "arg1", Type: model.ParameterTypeString, Required: true},
		{Name: "arg2", Type: model.ParameterTypeInteger, Required: false},
	}

	inputMapping := []model.InputMappingField{
		{Source: "arg1", Target: "query.keyword"},
	}

	outputMapping := &model.OutputMappingConfig{
		Fields: []model.OutputMappingField{
			{SourceField: "result", TargetField: "output", ValueType: "string", DefaultValue: ""},
		},
	}

	data := model.BuildData{
		Tools: []model.ToolSnapshot{
			{
				ID:            1,
				Name:          "test_tool",
				Description:   "test tool",
				Parameters:    params,
				InputMapping:  inputMapping,
				OutputMapping: outputMapping,
				State:         1,
			},
		},
		HTTPServices: []model.HTTPServiceSnapshot{
			{ID: 1, Name: "test_service", TargetURL: "http://test.com"},
		},
	}

	jsonData, err := sonic.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal BuildData: %v", err)
	}

	var parsed model.BuildData
	if err := sonic.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("failed to unmarshal BuildData: %v", err)
	}

	if len(parsed.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(parsed.Tools))
	}

	tool := parsed.Tools[0]
	if len(tool.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(tool.Parameters))
	}
	if tool.Parameters[0].Name != "arg1" {
		t.Errorf("expected first param name 'arg1', got '%s'", tool.Parameters[0].Name)
	}
	if tool.Parameters[0].Type != model.ParameterTypeString {
		t.Errorf("expected first param type 'string', got '%s'", tool.Parameters[0].Type)
	}

	if len(tool.InputMapping) != 1 {
		t.Errorf("expected 1 input mapping, got %d", len(tool.InputMapping))
	}
	if tool.InputMapping[0].Source != "arg1" {
		t.Errorf("expected input mapping source 'arg1', got '%s'", tool.InputMapping[0].Source)
	}
	if tool.InputMapping[0].Target != "query.keyword" {
		t.Errorf("expected input mapping target 'query.keyword', got '%s'", tool.InputMapping[0].Target)
	}

	if tool.OutputMapping == nil {
		t.Fatal("expected output mapping, got nil")
	}
	if len(tool.OutputMapping.Fields) != 1 {
		t.Errorf("expected 1 output mapping field, got %d", len(tool.OutputMapping.Fields))
	}
}

func TestBuildData_JSONNotBase64(t *testing.T) {
	params := []model.ParameterDefinition{
		{Name: "test_param", Type: model.ParameterTypeString},
	}

	data := model.BuildData{
		Tools: []model.ToolSnapshot{
			{
				ID:         1,
				Name:       "test_tool",
				Parameters: params,
				State:      1,
			},
		},
	}

	jsonData, err := sonic.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal BuildData: %v", err)
	}

	jsonStr := string(jsonData)

	var rawJSON map[string]any
	if err := json.Unmarshal(jsonData, &rawJSON); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	tools, ok := rawJSON["tools"].([]any)
	if !ok {
		t.Fatal("expected tools to be an array")
	}
	if len(tools) == 0 {
		t.Fatal("expected at least 1 tool")
	}

	tool, ok := tools[0].(map[string]any)
	if !ok {
		t.Fatal("expected tool to be an object")
	}

	paramsField, ok := tool["parameters"]
	if !ok {
		t.Fatal("expected parameters field")
	}

	paramsArr, ok := paramsField.([]any)
	if !ok {
		t.Fatalf("parameters should be an array, got %T", paramsField)
	}
	if len(paramsArr) == 0 {
		t.Fatal("parameters array should not be empty")
	}

	firstParam, ok := paramsArr[0].(map[string]any)
	if !ok {
		t.Fatal("first param should be an object")
	}
	if firstParam["name"] != "test_param" {
		t.Errorf("expected param name 'test_param', got '%v'", firstParam["name"])
	}

	_ = jsonStr
}
