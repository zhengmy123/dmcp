package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

type ProxyServerCacheService struct {
	redis *redis.Client
	ttl   time.Duration
}

type CachedProxyServer struct {
	ID             uint              `json:"id"`
	VAuthKey       string            `json:"vauth_key"`
	Name           string            `json:"name"`
	HTTPServerURL  string            `json:"http_server_url"`
	AuthHeader     string            `json:"auth_header"`
	TimeoutSeconds int              `json:"timeout_seconds"`
	ExtraHeaders   map[string]string `json:"extra_headers"`
	State          int              `json:"state"`
}

func NewProxyServerCacheService(redisClient *redis.Client, ttl time.Duration) *ProxyServerCacheService {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &ProxyServerCacheService{
		redis: redisClient,
		ttl:   ttl,
	}
}

func (s *ProxyServerCacheService) buildKey(vauthKey string) string {
	return fmt.Sprintf("mcp:proxy:%s", vauthKey)
}

func (s *ProxyServerCacheService) GetProxyServer(ctx context.Context, vauthKey string) (*CachedProxyServer, error) {
	if s.redis == nil {
		return nil, nil
	}

	data, err := s.redis.Get(ctx, s.buildKey(vauthKey)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var cached CachedProxyServer
	if err := sonic.Unmarshal([]byte(data), &cached); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	return &cached, nil
}

func (s *ProxyServerCacheService) SetProxyServer(ctx context.Context, vauthKey string, server *CachedProxyServer) error {
	if s.redis == nil {
		return nil
	}

	data, err := sonic.Marshal(server)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	err = s.redis.Set(ctx, s.buildKey(vauthKey), string(data), s.ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (s *ProxyServerCacheService) DeleteProxyServer(ctx context.Context, vauthKey string) error {
	if s.redis == nil {
		return nil
	}

	err := s.redis.Del(ctx, s.buildKey(vauthKey)).Err()
	if err != nil {
		return fmt.Errorf("redis del: %w", err)
	}

	return nil
}

func (s *ProxyServerCacheService) UpdateProxyServer(ctx context.Context, vauthKey string, server *CachedProxyServer) error {
	if s.redis == nil {
		return nil
	}

	data, err := sonic.Marshal(server)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	err = s.redis.Set(ctx, s.buildKey(vauthKey), string(data), s.ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

var _ json.Marshaler = (*CachedProxyServer)(nil)
var _ json.Unmarshaler = (*CachedProxyServer)(nil)

func (c *CachedProxyServer) MarshalJSON() ([]byte, error) {
	type Alias CachedProxyServer
	return sonic.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	})
}

func (c *CachedProxyServer) UnmarshalJSON(data []byte) error {
	type Alias CachedProxyServer
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	return sonic.Unmarshal(data, aux)
}