package tooldef

import (
	"encoding/json"
	"testing"

	"dynamic_mcp_go_server/internal/domain/model"
)

func TestParseToolDefinitionsFromRedisData(t *testing.T) {
	raw := []byte(`[
  {
    "name": "search_users",
    "description": "Search users",
    "parameters": [
      {"name":"query","type":"string","required":true}
    ],
    "enabled": true
  }
]`)

	defs, err := model.ParseToolDefinitions(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected one definition, got %d", len(defs))
	}

	def := defs[0]
	if def.Name != "search_users" {
		t.Fatalf("unexpected name: %s", def.Name)
	}
	if def.Description != "Search users" {
		t.Fatalf("unexpected description: %s", def.Description)
	}
	if len(def.Parameters) == 0 {
		t.Fatal("expected parameters to be parsed")
	}
}

func TestParseToolDefinitionsRequiresName(t *testing.T) {
	raw := []byte(`[
  {
    "description": "x",
    "parameters": [],
    "enabled": true
  }
]`)

	if _, err := model.ParseToolDefinitions(raw); err == nil {
		t.Fatalf("expected error when name is missing")
	}
}

func TestParseToolDefinitionsRequiresEnabledField(t *testing.T) {
	raw := []byte(`[
  {
    "name": "search_users",
    "description": "Search users",
    "parameters": []
  }
]`)

	if _, err := model.ParseToolDefinitions(raw); err == nil {
		t.Fatalf("expected error when enabled is missing")
	}
}

func TestParseToolDefinitionsRejectsEnabledFalse(t *testing.T) {
	raw := []byte(`[
  {
    "name": "search_users",
    "description": "Search users",
    "parameters": [],
    "enabled": false
  }
]`)

	if _, err := model.ParseToolDefinitions(raw); err == nil {
		t.Fatalf("expected error when enabled is false")
	}
}

func TestParameterDefinitionSerialization(t *testing.T) {
	params := []model.ParameterDefinition{
		{
			Name:        "query",
			Type:        model.ParameterTypeString,
			Required:    true,
			Description: "Search keyword",
		},
		{
			Name:        "limit",
			Type:        model.ParameterTypeInteger,
			Required:    false,
			Default:     10,
		},
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed []model.ParameterDefinition
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(parsed) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(parsed))
	}
	if parsed[0].Name != "query" || parsed[0].Required != true {
		t.Fatalf("unexpected first parameter: %+v", parsed[0])
	}
	if parsed[1].Name != "limit" || parsed[1].Required != false {
		t.Fatalf("unexpected second parameter: %+v", parsed[1])
	}
}