package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type BuildInfoCacheService struct {
	redis *redis.Client
	ttl   time.Duration
}

type CachedBuildUUID struct {
	BuildUUID string `json:"build_uuid"`
	Version   int    `json:"version"`
}

func NewBuildInfoCacheService(redisClient *redis.Client, ttl time.Duration) *BuildInfoCacheService {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &BuildInfoCacheService{
		redis: redisClient,
		ttl:   ttl,
	}
}

func (s *BuildInfoCacheService) buildKey(vauthKey string) string {
	return fmt.Sprintf("mcp:vauth:%s", vauthKey)
}

var setWithVersionScript = redis.NewScript(`
local key = KEYS[1]
local value = ARGV[1]
local ttl = tonumber(ARGV[2])

local existing = redis.call('GET', key)
if existing then
    local existingValue = cjson.decode(existing)
    local newValue = cjson.decode(value)
    if existingValue.version > newValue.version then
        return 0
    end
end

redis.call('SET', key, value, 'EX', ttl)
return 1
`)

func (s *BuildInfoCacheService) GetBuildUUID(ctx context.Context, vauthKey string) (*CachedBuildUUID, error) {
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

	var cached CachedBuildUUID
	if err := json.Unmarshal([]byte(data), &cached); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	return &cached, nil
}

func (s *BuildInfoCacheService) SetBuildUUID(ctx context.Context, vauthKey string, uuid *CachedBuildUUID) error {
	if s.redis == nil {
		return nil
	}

	data, err := json.Marshal(uuid)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	_, err = setWithVersionScript.Run(ctx, s.redis, []string{s.buildKey(vauthKey)}, string(data), int(s.ttl.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("redis set with version: %w", err)
	}

	return nil
}

func (s *BuildInfoCacheService) DeleteBuildUUID(ctx context.Context, vauthKey string) error {
	if s.redis == nil {
		return nil
	}

	err := s.redis.Del(ctx, s.buildKey(vauthKey)).Err()
	if err != nil {
		return fmt.Errorf("redis del: %w", err)
	}

	return nil
}
