package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestProxyServerCacheService_BuildKey(t *testing.T) {
	svc := NewProxyServerCacheService(nil, 0)
	key := svc.buildKey("test-key")
	if key != "mcp:proxy:test-key" {
		t.Errorf("expected 'mcp:proxy:test-key', got '%s'", key)
	}
}

func TestProxyServerCacheService_GetProxyServer_NilRedis(t *testing.T) {
	svc := NewProxyServerCacheService(nil, 0)
	result, err := svc.GetProxyServer(context.Background(), "test")
	if err != nil {
		t.Errorf("expected no error with nil redis, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result with nil redis, got %v", result)
	}
}

func TestProxyServerCacheService_SetProxyServer_NilRedis(t *testing.T) {
	svc := NewProxyServerCacheService(nil, 0)
	err := svc.SetProxyServer(context.Background(), "test", &CachedProxyServer{
		ID:             1,
		VAuthKey:       "test-key",
		Name:           "Test Server",
		HTTPServerURL:  "https://example.com/mcp",
		AuthHeader:     "Bearer token",
		TimeoutSeconds: 30,
		ExtraHeaders:   map[string]string{"X-Custom": "value"},
		State:          1,
	})
	if err != nil {
		t.Errorf("expected no error with nil redis, got %v", err)
	}
}

func TestProxyServerCacheService_DeleteProxyServer_NilRedis(t *testing.T) {
	svc := NewProxyServerCacheService(nil, 0)
	err := svc.DeleteProxyServer(context.Background(), "test")
	if err != nil {
		t.Errorf("expected no error with nil redis, got %v", err)
	}
}

func TestProxyServerCacheService_CachedProxyServer_JSON(t *testing.T) {
	cached := CachedProxyServer{
		ID:             1,
		VAuthKey:       "test-key",
		Name:           "Test Server",
		HTTPServerURL:  "https://example.com/mcp",
		AuthHeader:     "Bearer token",
		TimeoutSeconds: 30,
		ExtraHeaders:   map[string]string{"X-Custom": "value"},
		State:          1,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}

	var decoded CachedProxyServer
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}

	if decoded.ID != cached.ID {
		t.Errorf("expected ID %d, got %d", cached.ID, decoded.ID)
	}
	if decoded.VAuthKey != cached.VAuthKey {
		t.Errorf("expected VAuthKey '%s', got '%s'", cached.VAuthKey, decoded.VAuthKey)
	}
	if decoded.HTTPServerURL != cached.HTTPServerURL {
		t.Errorf("expected HTTPServerURL '%s', got '%s'", cached.HTTPServerURL, decoded.HTTPServerURL)
	}
	if decoded.State != cached.State {
		t.Errorf("expected State %d, got %d", cached.State, decoded.State)
	}
}

func TestProxyServerCacheService_DefaultTTL(t *testing.T) {
	svc := NewProxyServerCacheService(nil, 0)
	if svc.ttl != 5*time.Minute {
		t.Errorf("expected default TTL 5min, got %v", svc.ttl)
	}
}

func TestProxyServerCacheService_CustomTTL(t *testing.T) {
	ttl := 10 * time.Minute
	svc := NewProxyServerCacheService(nil, ttl)
	if svc.ttl != ttl {
		t.Errorf("expected TTL %v, got %v", ttl, svc.ttl)
	}
}

func TestProxyServerCacheService_UpdateProxyServer_NilRedis(t *testing.T) {
	svc := NewProxyServerCacheService(nil, 0)
	err := svc.UpdateProxyServer(context.Background(), "test", &CachedProxyServer{
		ID:             1,
		VAuthKey:       "test-key",
		Name:           "Test Server Updated",
		HTTPServerURL:  "https://example.com/mcp",
		AuthHeader:     "Bearer token",
		TimeoutSeconds: 60,
		State:          1,
	})
	if err != nil {
		t.Errorf("expected no error with nil redis, got %v", err)
	}
}