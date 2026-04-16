package tooldef

import "testing"

func TestParseToolDefinitionsFromRedisData(t *testing.T) {
	raw := []byte(`[
  {
    "vauth_key": "user-service",
    "server_desc": "User tools",
    "name": "search_users",
    "description": "Search users",
    "parameters": [
      {"name":"query","type":"string","required":true}
    ],
    "enabled": true
  }
]`)

	defs, err := ParseToolDefinitions(raw)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected one definition, got %d", len(defs))
	}

	def := defs[0]
	if def.VAuthKey != "user-service" {
		t.Fatalf("unexpected vauth_key: %s", def.VAuthKey)
	}
	if def.ServerDesc != "User tools" {
		t.Fatalf("unexpected server_desc: %s", def.ServerDesc)
	}
}

func TestParseToolDefinitionsRequiresVAuthKey(t *testing.T) {
	raw := []byte(`[
  {
    "name": "search_users",
    "description": "Search users",
    "parameters": [],
    "enabled": true
  }
]`)

	if _, err := ParseToolDefinitions(raw); err == nil {
		t.Fatalf("expected error when vauth_key is missing")
	}
}

func TestParseToolDefinitionsRequiresName(t *testing.T) {
	raw := []byte(`[
  {
    "vauth_key": "user-service",
    "description": "x",
    "parameters": [],
    "enabled": true
  }
]`)

	if _, err := ParseToolDefinitions(raw); err == nil {
		t.Fatalf("expected error when name is missing")
	}
}

func TestParseToolDefinitionsRequiresEnabledField(t *testing.T) {
	raw := []byte(`[
  {
    "vauth_key": "user-service",
    "name": "search_users",
    "description": "Search users",
    "parameters": []
  }
]`)

	if _, err := ParseToolDefinitions(raw); err == nil {
		t.Fatalf("expected error when enabled is missing")
	}
}

func TestParseToolDefinitionsRejectsEnabledFalse(t *testing.T) {
	raw := []byte(`[
  {
    "vauth_key": "user-service",
    "name": "search_users",
    "description": "Search users",
    "parameters": [],
    "enabled": false
  }
]`)

	if _, err := ParseToolDefinitions(raw); err == nil {
		t.Fatalf("expected error when enabled is false")
	}
}
