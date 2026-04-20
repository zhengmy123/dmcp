package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestBuildInfoCacheService_BuildKey(t *testing.T) {
	svc := NewBuildInfoCacheService(nil, 0)
	key := svc.buildKey("test-key")
	if key != "mcp:vauth:test-key" {
		t.Errorf("expected 'mcp:vauth:test-key', got '%s'", key)
	}
}

func TestBuildInfoCacheService_GetBuildUUID_NilRedis(t *testing.T) {
	svc := NewBuildInfoCacheService(nil, 0)
	result, err := svc.GetBuildUUID(context.Background(), "test")
	if err != nil {
		t.Errorf("expected no error with nil redis, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result with nil redis, got %v", result)
	}
}

func TestBuildInfoCacheService_SetBuildUUID_NilRedis(t *testing.T) {
	svc := NewBuildInfoCacheService(nil, 0)
	err := svc.SetBuildUUID(context.Background(), "test", &CachedBuildUUID{
		BuildUUID: "uuid-123",
		Version:   1,
	})
	if err != nil {
		t.Errorf("expected no error with nil redis, got %v", err)
	}
}

func TestBuildInfoCacheService_DeleteBuildUUID_NilRedis(t *testing.T) {
	svc := NewBuildInfoCacheService(nil, 0)
	err := svc.DeleteBuildUUID(context.Background(), "test")
	if err != nil {
		t.Errorf("expected no error with nil redis, got %v", err)
	}
}

func TestBuildInfoCacheService_CachedBuildUUID_JSON(t *testing.T) {
	cached := CachedBuildUUID{
		BuildUUID: "uuid-123",
		Version:   5,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}

	var decoded CachedBuildUUID
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}

	if decoded.BuildUUID != cached.BuildUUID {
		t.Errorf("expected BuildUUID '%s', got '%s'", cached.BuildUUID, decoded.BuildUUID)
	}
	if decoded.Version != cached.Version {
		t.Errorf("expected Version %d, got %d", cached.Version, decoded.Version)
	}
}

func TestBuildInfoCacheService_DefaultTTL(t *testing.T) {
	svc := NewBuildInfoCacheService(nil, 0)
	if svc.ttl != 5*time.Minute {
		t.Errorf("expected default TTL 5min, got %v", svc.ttl)
	}
}

func TestBuildInfoCacheService_CustomTTL(t *testing.T) {
	ttl := 10 * time.Minute
	svc := NewBuildInfoCacheService(nil, ttl)
	if svc.ttl != ttl {
		t.Errorf("expected TTL %v, got %v", ttl, svc.ttl)
	}
}
