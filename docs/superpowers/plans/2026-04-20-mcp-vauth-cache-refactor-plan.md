# MCP VAuth Key 缓存重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构 MCPGroupManager，实现三级缓存策略（内存双LRU → Redis → MySQL），按需构建 handler

**Architecture:**
1. 新增泛型 LRU 缓存库（基础 LRU + 双 LRU）
2. 重构 MCPGroupManager，按需从 Redis/MySQL 获取 build_info 并构建 handler
3. 新增 BuildInfoCacheService，管理 Redis 缓存（带版本号的 Lua 脚本）
4. 修改 ToolBindingService，触发缓存失效

**Tech Stack:** Go, GORM, MySQL, go-redis/v9, Sonic JSON

---

## 文件结构

```
internal/common/cache/
├── lru.go              # 新增：基础泛型 LRU
└── two_level_lru.go    # 新增：双 LRU

internal/service/
├── build_info_cache.go     # 新增：BuildInfo 缓存服务
└── group_manager.go        # 重构：MCPGroupManager
```

---

## Task 1: 创建基础泛型 LRU

**Files:**
- Create: `internal/common/cache/lru.go`
- Test: `test/common/cache/lru_test.go`

- [ ] **Step 1: 创建目录**

```bash
mkdir -p internal/common/cache
mkdir -p test/common/cache
```

- [ ] **Step 2: 创建 lru.go**

```go
package cache

import (
    "container/list"
    "sync"
)

type entry[K comparable, V any] struct {
    key K
    val V
}

type LRU[K comparable, V any] struct {
    capacity int
    items    map[K]*list.Element
    order    *list.List
    mu       sync.RWMutex
}

func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
    return &LRU[K, V]{
        capacity: capacity,
        items:    make(map[K]*list.Element),
        order:    list.New(),
    }
}

func (l *LRU[K, V]) Get(key K) (V, bool) {
    l.mu.Lock()
    defer l.mu.Unlock()

    elem, ok := l.items[key]
    if !ok {
        return *new(V), false
    }
    l.order.MoveToFront(elem)
    return elem.Value.(*entry[K, V]).val, true
}

func (l *LRU[K, V]) Set(key K, value V) {
    l.mu.Lock()
    defer l.mu.Unlock()

    if elem, ok := l.items[key]; ok {
        l.order.MoveToFront(elem)
        elem.Value.(*entry[K, V]).val = value
        return
    }

    if l.order.Len() >= l.capacity {
        l.evict()
    }

    elem := l.order.PushFront(&entry[K, V]{key: key, val: value})
    l.items[key] = elem
}

func (l *LRU[K, V]) Delete(key K) {
    l.mu.Lock()
    defer l.mu.Unlock()

    if elem, ok := l.items[key]; ok {
        l.order.Remove(elem)
        delete(l.items, key)
    }
}

func (l *LRU[K, V]) Contains(key K) bool {
    l.mu.RLock()
    defer l.mu.RUnlock()
    _, ok := l.items[key]
    return ok
}

func (l *LRU[K, V]) Len() int {
    l.mu.RLock()
    defer l.mu.RUnlock()
    return l.order.Len()
}

func (l *LRU[K, V]) Clear() {
    l.mu.Lock()
    defer l.mu.Unlock()
    l.items = make(map[K]*list.Element)
    l.order = list.New()
}

func (l *LRU[K, V]) evict() {
    elem := l.order.Back()
    if elem != nil {
        e := elem.Value.(*entry[K, V])
        delete(l.items, e.key)
        l.order.Remove(elem)
    }
}
```

- [ ] **Step 3: 创建 lru_test.go**

```go
package cache

import (
    "testing"
)

func TestLRU_BasicOperations(t *testing.T) {
    lru := NewLRU[string, int](3)

    if lru.Len() != 0 {
        t.Errorf("expected empty LRU, got %d", lru.Len())
    }

    lru.Set("a", 1)
    lru.Set("b", 2)
    lru.Set("c", 3)

    if lru.Len() != 3 {
        t.Errorf("expected length 3, got %d", lru.Len())
    }

    val, ok := lru.Get("a")
    if !ok || val != 1 {
        t.Errorf("expected a=1, got ok=%v val=%d", ok, val)
    }

    val, ok = lru.Get("nonexistent")
    if ok {
        t.Errorf("expected not found for nonexistent key")
    }
}

func TestLRU_Eviction(t *testing.T) {
    lru := NewLRU[string, int](3)

    lru.Set("a", 1)
    lru.Set("b", 2)
    lru.Set("c", 3)
    lru.Set("d", 4)

    _, ok := lru.Get("a")
    if ok {
        t.Errorf("expected 'a' to be evicted")
    }

    _, ok = lru.Get("d")
    if !ok {
        t.Errorf("expected 'd' to exist")
    }
}

func TestLRU_Update(t *testing.T) {
    lru := NewLRU[string, int](3)

    lru.Set("a", 1)
    lru.Set("a", 10)

    val, _ := lru.Get("a")
    if val != 10 {
        t.Errorf("expected a=10, got %d", val)
    }

    if lru.Len() != 1 {
        t.Errorf("expected length 1, got %d", lru.Len())
    }
}

func TestLRU_Delete(t *testing.T) {
    lru := NewLRU[string, int](3)

    lru.Set("a", 1)
    lru.Set("b", 2)

    lru.Delete("a")

    _, ok := lru.Get("a")
    if ok {
        t.Errorf("expected 'a' to be deleted")
    }

    val, _ := lru.Get("b")
    if val != 2 {
        t.Errorf("expected b=2, got %d", val)
    }
}

func TestLRU_Clear(t *testing.T) {
    lru := NewLRU[string, int](3)

    lru.Set("a", 1)
    lru.Set("b", 2)

    lru.Clear()

    if lru.Len() != 0 {
        t.Errorf("expected empty LRU after clear, got %d", lru.Len())
    }
}
```

- [ ] **Step 4: 运行测试**

Run: `go test ./test/common/cache/lru_test.go -v`
Expected: PASS

---

## Task 2: 创建双 LRU

**Files:**
- Create: `internal/common/cache/two_level_lru.go`
- Test: `test/common/cache/two_level_lru_test.go`

- [ ] **Step 1: 创建 two_level_lru.go**

```go
package cache

import (
    "sync"
    "time"
)

type Config struct {
    L1Size     int
    L2Size     int
    L2Window   time.Duration
    L2Threshold int
}

type AccessTracker[K comparable] struct {
    mu         sync.RWMutex
    timestamps map[K][]time.Time
    threshold int
    window    time.Duration
}

func NewAccessTracker[K comparable](threshold int, window time.Duration) *AccessTracker[K] {
    return &AccessTracker[K]{
        timestamps: make(map[K][]time.Time),
        threshold:  threshold,
        window:    window,
    }
}

func (t *AccessTracker[K]) RecordAccess(key K) bool {
    now := time.Now()
    t.mu.Lock()
    defer t.mu.Unlock()

    cutoff := now.Add(-t.window)
    var recent []time.Time
    for _, ts := range t.timestamps[key] {
        if ts.After(cutoff) {
            recent = append(recent, ts)
        }
    }
    t.timestamps[key] = recent
    t.timestamps[key] = append(t.timestamps[key], now)

    return len(t.timestamps[key]) >= t.threshold
}

func (t *AccessTracker[K]) Clear() {
    t.mu.Lock()
    defer t.mu.Unlock()
    t.timestamps = make(map[K][]time.Time)
}

type TwoLevelLRU[K comparable, V any] struct {
    l1     *LRU[K, V]
    l2     *LRU[K, V]
    access *AccessTracker[K]
    config Config
}

func NewTwoLevelLRU[K comparable, V any](config Config) *TwoLevelLRU[K, V] {
    if config.L2Window == 0 {
        config.L2Window = time.Second
    }
    if config.L2Threshold == 0 {
        config.L2Threshold = 2
    }
    if config.L1Size == 0 {
        config.L1Size = 2000
    }
    if config.L2Size == 0 {
        config.L2Size = 2000
    }

    return &TwoLevelLRU[K, V]{
        l1:     NewLRU[K, V](config.L1Size),
        l2:     NewLRU[K, V](config.L2Size),
        access: NewAccessTracker[K](config.L2Threshold, config.L2Window),
        config: config,
    }
}

func (t *TwoLevelLRU[K, V]) Get(key K) (V, bool) {
    if val, ok := t.l2.Get(key); ok {
        return val, true
    }
    if val, ok := t.l1.Get(key); ok {
        t.access.RecordAccess(key)
        return val, true
    }
    return *new(V), false
}

func (t *TwoLevelLRU[K, V]) Set(key K, value V) {
    t.l1.Set(key, value)
    if t.access.RecordAccess(key) {
        t.l2.Set(key, value)
    }
}

func (t *TwoLevelLRU[K, V]) Delete(key K) {
    t.l1.Delete(key)
    t.l2.Delete(key)
}

func (t *TwoLevelLRU[K, V]) Len() (l1Len, l2Len int) {
    return t.l1.Len(), t.l2.Len()
}

func (t *TwoLevelLRU[K, V]) Clear() {
    t.l1.Clear()
    t.l2.Clear()
    t.access.Clear()
}
```

- [ ] **Step 2: 创建 two_level_lru_test.go**

```go
package cache

import (
    "testing"
    "time"
)

func TestTwoLevelLRU_BasicOperations(t *testing.T) {
    config := Config{
        L1Size:      3,
        L2Size:      3,
        L2Window:    time.Second,
        L2Threshold: 2,
    }
    lru := NewTwoLevelLRU[string, int](config)

    l1, l2 := lru.Len()
    if l1 != 0 || l2 != 0 {
        t.Errorf("expected empty LRU, got L1=%d L2=%d", l1, l2)
    }

    lru.Set("a", 1)
    l1, _ = lru.Len()
    if l1 != 1 {
        t.Errorf("expected L1 length 1, got %d", l1)
    }
}

func TestTwoLevelLRU_L2Admission(t *testing.T) {
    config := Config{
        L1Size:      10,
        L2Size:      10,
        L2Window:    time.Second,
        L2Threshold: 2,
    }
    lru := NewTwoLevelLRU[string, int](config)

    lru.Set("a", 1)
    time.Sleep(10 * time.Millisecond)
    lru.Set("a", 2)
    time.Sleep(10 * time.Millisecond)
    lru.Set("a", 3)

    _, l2 := lru.Len()
    if l2 != 1 {
        t.Errorf("expected 'a' to be in L2 after 3 accesses, got L2=%d", l2)
    }

    val, ok := lru.Get("a")
    if !ok || val != 3 {
        t.Errorf("expected a=3, got ok=%v val=%d", ok, val)
    }
}

func TestTwoLevelLRU_L1Only(t *testing.T) {
    config := Config{
        L1Size:      10,
        L2Size:      10,
        L2Window:    time.Second,
        L2Threshold: 2,
    }
    lru := NewTwoLevelLRU[string, int](config)

    lru.Set("a", 1)

    _, l2 := lru.Len()
    if l2 != 0 {
        t.Errorf("expected 'a' not in L2 after single access, got L2=%d", l2)
    }
}

func TestTwoLevelLRU_L2Eviction(t *testing.T) {
    config := Config{
        L1Size:      2,
        L2Size:      2,
        L2Window:    time.Second,
        L2Threshold: 2,
    }
    lru := NewTwoLevelLRU[string, int](config)

    for i := 0; i < 3; i++ {
        lru.Set(string(rune('a'+i)), i)
        time.Sleep(5 * time.Millisecond)
    }

    for i := 0; i < 3; i++ {
        lru.Set(string(rune('a'+i)), i+10)
        time.Sleep(5 * time.Millisecond)
    }

    _, l2 := lru.Len()
    if l2 > 2 {
        t.Errorf("expected L2 to respect capacity 2, got L2=%d", l2)
    }
}

func TestTwoLevelLRU_Delete(t *testing.T) {
    config := Config{
        L1Size:      10,
        L2Size:      10,
        L2Window:    time.Second,
        L2Threshold: 2,
    }
    lru := NewTwoLevelLRU[string, int](config)

    lru.Set("a", 1)
    lru.Set("a", 2)
    lru.Set("a", 3)

    lru.Delete("a")

    _, l2 := lru.Len()
    if l2 != 0 {
        t.Errorf("expected 'a' deleted from L2, got L2=%d", l2)
    }
}

func TestTwoLevelLRU_Clear(t *testing.T) {
    config := Config{
        L1Size:      10,
        L2Size:      10,
        L2Window:    time.Second,
        L2Threshold: 2,
    }
    lru := NewTwoLevelLRU[string, int](config)

    lru.Set("a", 1)
    lru.Set("b", 2)
    lru.Set("c", 3)

    lru.Clear()

    l1, l2 := lru.Len()
    if l1 != 0 || l2 != 0 {
        t.Errorf("expected empty LRU after clear, got L1=%d L2=%d", l1, l2)
    }
}
```

- [ ] **Step 3: 运行测试**

Run: `go test ./test/common/cache/two_level_lru_test.go -v`
Expected: PASS

---

## Task 3: 创建 BuildInfoCacheService

**Files:**
- Create: `internal/service/build_info_cache.go`
- Test: `test/service/build_info_cache_test.go`

- [ ] **Step 1: 创建 build_info_cache.go**

```go
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
```

- [ ] **Step 2: 创建 build_info_cache_test.go**

```go
package service

import (
    "context"
    "testing"
    "time"

    "github.com/redis/go-redis/v9"
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
```

- [ ] **Step 3: 运行测试**

Run: `go test ./test/service/build_info_cache_test.go -v`
Expected: PASS

---

## Task 4: 重构 MCPGroupManager

**Files:**
- Modify: `internal/service/group_manager.go`

- [ ] **Step 1: 读取现有 group_manager.go**

```go
package service

import (
    "net/http"
    "sync"

    "github.com/mark3labs/mcp-go/server"
)

type MCPGroupManager struct {
    serverName    string
    serverVersion string
    authService   *AuthService

    mu       sync.RWMutex
    handlers map[string]http.Handler // vauthKey -> StreamableHTTPServer
}
```

- [ ] **Step 2: 重构 group_manager.go**

```go
package service

import (
    "context"
    "fmt"
    "net/http"
    "sync"
    "time"

    "dynamic_mcp_go_server/internal/common/cache"
    "dynamic_mcp_go_server/internal/domain/model"
    "dynamic_mcp_go_server/internal/domain/repository"

    "github.com/bytedance/sonic"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

var ErrBuildInfoNotFound = fmt.Errorf("build info not found")

type MCPGroupManagerConfig struct {
    Cache cache.Config
    Redis time.Duration
}

type MCPGroupManager struct {
    serverName    string
    serverVersion string
    authService   *AuthService

    buildInfoCache *BuildInfoCacheService
    buildInfoStore repository.ServerBuildInfoStore
    serverStore    repository.MCPServerStore

    cache *cache.TwoLevelLRU[string, http.Handler]

    mu sync.RWMutex
}

func NewMCPGroupManager(
    serverName, serverVersion string,
    authService *AuthService,
    buildInfoCache *BuildInfoCacheService,
    buildInfoStore repository.ServerBuildInfoStore,
    serverStore repository.MCPServerStore,
    config MCPGroupManagerConfig,
) *MCPGroupManager {
    return &MCPGroupManager{
        serverName:     serverName,
        serverVersion:  serverVersion,
        authService:    authService,
        buildInfoCache: buildInfoCache,
        buildInfoStore: buildInfoStore,
        serverStore:    serverStore,
        cache:          cache.NewTwoLevelLRU[string, http.Handler](config.Cache),
    }
}

func (m *MCPGroupManager) GetHandler(vauthKey string) (http.Handler, error) {
    cached, ok := m.cache.Get(vauthKey)
    if ok {
        return cached, nil
    }

    buildInfo, err := m.loadBuildInfo(context.Background(), vauthKey)
    if err != nil {
        return nil, err
    }
    if buildInfo == nil {
        return nil, ErrBuildInfoNotFound
    }

    handler, err := m.buildHandler(buildInfo)
    if err != nil {
        return nil, err
    }

    m.cache.Set(vauthKey, handler)

    return handler, nil
}

func (m *MCPGroupManager) loadBuildInfo(ctx context.Context, vauthKey string) (*model.ServerBuildInfo, error) {
    if m.buildInfoCache != nil {
        cached, err := m.buildInfoCache.GetBuildUUID(ctx, vauthKey)
        if err == nil && cached != nil {
            if m.buildInfoStore != nil {
                return m.buildInfoStore.GetByBuildUUID(ctx, cached.BuildUUID)
            }
        }
    }

    if m.serverStore == nil || m.buildInfoStore == nil {
        return nil, nil
    }

    server, err := m.serverStore.GetByVAuthKey(ctx, vauthKey)
    if err != nil {
        return nil, nil
    }

    buildInfo, err := m.buildInfoStore.GetActiveByServerID(ctx, server.ID)
    if err != nil {
        return nil, nil
    }

    if buildInfo != nil && m.buildInfoCache != nil {
        _ = m.buildInfoCache.SetBuildUUID(ctx, vauthKey, &CachedBuildUUID{
            BuildUUID: buildInfo.BuildUUID,
            Version:   buildInfo.Version,
        })
    }

    return buildInfo, nil
}

func (m *MCPGroupManager) buildHandler(info *model.ServerBuildInfo) (http.Handler, error) {
    if info == nil || info.BuildData == "" {
        return nil, ErrBuildInfoNotFound
    }

    var buildData model.BuildData
    if err := sonic.Unmarshal([]byte(info.BuildData), &buildData); err != nil {
        return nil, fmt.Errorf("unmarshal build_data: %w", err)
    }

    var tools []server.ServerTool
    for _, t := range buildData.Tools {
        if !t.Enabled {
            continue
        }
        tool, err := m.convertToServerTool(t)
        if err != nil {
            return nil, err
        }
        tools = append(tools, tool)
    }

    groupMCP := server.NewMCPServer(
        m.serverName+"::"+info.BuildUUID[:8],
        m.serverVersion,
        server.WithToolCapabilities(true),
        server.WithRecovery(),
    )
    groupMCP.SetTools(tools...)

    return server.NewStreamableHTTPServer(
        groupMCP,
        server.WithStateLess(true),
    ), nil
}

func (m *MCPGroupManager) convertToServerTool(t model.ToolSnapshot) (server.ServerTool, error) {
    params, err := parseToolParams(t.Parameters)
    if err != nil {
        return server.ServerTool{}, fmt.Errorf("parse parameters for tool %q: %w", t.Name, err)
    }
    rawSchema, err := toRawInputSchema(params)
    if err != nil {
        return server.ServerTool{}, err
    }

    tool := mcp.NewToolWithRawSchema(t.Name, t.Description, rawSchema)
    return server.ServerTool{
        Tool:    tool,
        Handler: defaultHandler(t.Name),
    }, nil
}

func parseToolParams(data []byte) ([]ParameterDefinition, error) {
    if len(data) == 0 {
        return nil, nil
    }
    var params []ParameterDefinition
    if err := sonic.Unmarshal(data, &params); err != nil {
        return nil, err
    }
    return params, nil
}

func toRawInputSchema(params []ParameterDefinition) ([]byte, error) {
    properties := make(map[string]any, len(params))
    required := make([]string, 0)

    for _, p := range params {
        prop := map[string]any{
            "type": string(p.Type),
        }
        if p.Description != "" {
            prop["description"] = p.Description
        }
        if p.Default != nil {
            prop["default"] = p.Default
        }
        if len(p.Enum) > 0 {
            prop["enum"] = p.Enum
        }
        if p.Minimum != nil {
            prop["minimum"] = *p.Minimum
        }
        if p.Maximum != nil {
            prop["maximum"] = *p.Maximum
        }
        properties[p.Name] = prop
        if p.Required {
            required = append(required, p.Name)
        }
    }

    required = append(required, "")
    required = required[:len(required)-1]

    schema := map[string]any{
        "type":       "object",
        "properties": properties,
        "required":   required,
    }
    raw, err := sonic.Marshal(schema)
    if err != nil {
        return nil, err
    }
    return raw, nil
}

func defaultHandler(toolName string) server.ToolHandlerFunc {
    return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        payload := map[string]any{
            "tool":      toolName,
            "arguments": request.GetArguments(),
            "note":      "Replace defaultHandler with business logic.",
        }
        result, err := mcp.NewToolResultJSON(payload)
        if err != nil {
            return mcp.NewToolResultError(fmt.Sprintf("marshal result: %v", err)), nil
        }
        return result, nil
    }
}

func (m *MCPGroupManager) Handler(vauthKey string) (http.Handler, bool) {
    h, err := m.GetHandler(vauthKey)
    if err != nil {
        return nil, false
    }
    return h, true
}

func (m *MCPGroupManager) ListGroups() []string {
    return nil
}

type ParameterDefinition struct {
    Name        string   `json:"name"`
    Type        string   `json:"type"`
    Description string   `json:"description,omitempty"`
    Default     any      `json:"default,omitempty"`
    Enum        []string `json:"enum,omitempty"`
    Minimum     *float64 `json:"minimum,omitempty"`
    Maximum     *float64 `json:"maximum,omitempty"`
    Required    bool     `json:"required"`
}
```

- [ ] **Step 3: 编译检查**

Run: `go build ./internal/service/group_manager.go`
Expected: 无错误

---

## Task 5: 添加 repository 接口（如缺失）

**Files:**
- Check: `internal/domain/repository/server_build_info_repository.go`

- [ ] **Step 1: 检查 repository 接口是否存在**

如果不存在，创建：

```go
package repository

import (
    "context"

    "dynamic_mcp_go_server/internal/domain/model"
)

type ServerBuildInfoStore interface {
    GetByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error)
    GetActiveByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error)
    GetByBuildUUID(ctx context.Context, buildUUID string) (*model.ServerBuildInfo, error)
    Save(ctx context.Context, info *model.ServerBuildInfo) error
    UpdateState(ctx context.Context, id uint, state int) error
    GetMaxVersionByServerID(ctx context.Context, serverID uint) (int, error)
}
```

- [ ] **Step 2: 检查 GORM 实现是否存在**

如果不存在，创建：

```go
package database

import (
    "context"
    "dynamic_mcp_go_server/internal/domain/model"
    "errors"

    "gorm.io/gorm"
)

type GORMServerBuildInfoDAO struct {
    db *gorm.DB
}

func NewGORMServerBuildInfoDAO(db *gorm.DB) *GORMServerBuildInfoDAO {
    return &GORMServerBuildInfoDAO{db: db}
}

func (d *GORMServerBuildInfoDAO) GetByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
    var info model.ServerBuildInfo
    err := d.db.WithContext(ctx).Where("server_id = ?", serverID).First(&info).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &info, nil
}

func (d *GORMServerBuildInfoDAO) GetActiveByServerID(ctx context.Context, serverID uint) (*model.ServerBuildInfo, error) {
    var info model.ServerBuildInfo
    err := d.db.WithContext(ctx).Where("server_id = ? AND state = ?", serverID, 1).First(&info).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &info, nil
}

func (d *GORMServerBuildInfoDAO) GetByBuildUUID(ctx context.Context, buildUUID string) (*model.ServerBuildInfo, error) {
    var info model.ServerBuildInfo
    err := d.db.WithContext(ctx).Where("build_uuid = ?", buildUUID).First(&info).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &info, nil
}

func (d *GORMServerBuildInfoDAO) Save(ctx context.Context, info *model.ServerBuildInfo) error {
    return d.db.WithContext(ctx).Create(info).Error
}

func (d *GORMServerBuildInfoDAO) UpdateState(ctx context.Context, id uint, state int) error {
    return d.db.WithContext(ctx).Model(&model.ServerBuildInfo{}).Where("id = ?", id).Update("state", state).Error
}

func (d *GORMServerBuildInfoDAO) GetMaxVersionByServerID(ctx context.Context, serverID uint) (int, error) {
    var maxVersion int
    err := d.db.WithContext(ctx).Model(&model.ServerBuildInfo{}).
        Where("server_id = ?", serverID).
        Select("COALESCE(MAX(version), 0)").
        Scan(&maxVersion).Error
    if err != nil {
        return 0, err
    }
    return maxVersion, nil
}
```

---

## Task 6: 检查 model 定义

**Files:**
- Check: `internal/domain/model/server_build_info.go`

- [ ] **Step 1: 检查 model 定义**

```go
package model

import (
    "time"
)

type ServerBuildInfo struct {
    ID        uint      `gorm:"primaryKey;autoIncrement"`
    ServerID  uint      `gorm:"not null;index:idx_server_state"`
    Version   int       `gorm:"not null;default:1"`
    BuildUUID string    `gorm:"size:36;not null;uniqueIndex"`
    Hash      string    `gorm:"size:64;not null;index"`
    BuildData string    `gorm:"type:text"`
    State     int       `gorm:"not null;default:1;index:idx_server_state"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (ServerBuildInfo) TableName() string {
    return "server_build_info"
}

type BuildData struct {
    Tools       []ToolSnapshot        `json:"tools"`
    HTTPServices []HTTPServiceSnapshot `json:"http_services"`
}

type ToolSnapshot struct {
    ID          uint   `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Parameters  []byte `json:"parameters"`
    Enabled     bool   `json:"enabled"`
}

type HTTPServiceSnapshot struct {
    ID          uint                 `json:"id"`
    Name        string               `json:"name"`
    TargetURL   string               `json:"target_url"`
    Method      string               `json:"method"`
    Headers     map[string]string    `json:"headers"`
    BodyType    string               `json:"body_type"`
    Timeout     int                  `json:"timeout_seconds"`
    InputSchema []byte               `json:"input_schema"`
    OutputSchema []byte              `json:"output_schema"`
}
```

---

## Task 7: 修改 ScopedMCPHandler

**Files:**
- Modify: `internal/service/scoped_handler.go`

- [ ] **Step 1: 修改 ScopedMCPHandler 使用新的 GetHandler**

```go
package service

import (
    "net/http"
    "strings"
)

func NewScopedMCPHandler(manager *MCPGroupManager) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        groupKey := strings.Trim(strings.TrimPrefix(r.URL.Path, mcpPathPrefix), "/")
        if groupKey == "" || strings.Contains(groupKey, "/") {
            writeJSON(w, http.StatusNotFound, map[string]any{
                "error": "path must be /mcp/{vauth_key}",
                "path":  r.URL.Path,
            })
            return
        }

        h, err := manager.GetHandler(groupKey)
        if err != nil {
            writeJSON(w, http.StatusNotFound, map[string]any{
                "error":     "mcp server not found",
                "vauth_key": groupKey,
            })
            return
        }

        h.ServeHTTP(w, r)
    })
}
```

---

## Task 8: 更新 router.go 初始化

**Files:**
- Modify: `cmd/server/main.go` 或 `internal/api/http/router.go`

- [ ] **Step 1: 添加 BuildInfoCacheService 初始化**

在创建 MCPGroupManager 的地方，添加：

```go
buildInfoCache := service.NewBuildInfoCacheService(redisClient, 5*time.Minute)

config := service.MCPGroupManagerConfig{
    Cache: cache.Config{
        L1Size:      2000,
        L2Size:      2000,
        L2Window:    time.Second,
        L2Threshold: 2,
    },
    Redis: 5 * time.Minute,
}

groupManager := service.NewMCPGroupManager(
    serverName, serverVersion,
    authService,
    buildInfoCache,
    serverBuildInfoStore, // repository.ServerBuildInfoStore
    serverStore,          // repository.MCPServerStore
    config,
)
```

---

## Task 9: 修改 ToolBindingService 触发缓存失效

**Files:**
- Modify: `internal/service/tool_binding_service.go`

- [ ] **Step 1: 添加缓存失效调用**

在 BindTool、BatchBind、Unbind 等方法中，构建成功后调用：

```go
if s.buildInfoCache != nil {
    _ = s.buildInfoCache.DeleteBuildUUID(ctx, server.VAuthKey)
}
```

---

## 自检清单

- [ ] 所有 Task 完成
- [ ] 所有测试通过
- [ ] 符合 AGENTS.md 规范
- [ ] 错误处理完善
- [ ] 代码无 TODO/TBD 占位符

---

**Plan complete.** Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
