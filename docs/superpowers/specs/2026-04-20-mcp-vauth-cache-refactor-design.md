# MCP VAuth Key 调用逻辑重构设计

**日期**: 2026-04-20

## 1. 背景与目标

当前项目中的 `/mcp/:vauth_key` 调用逻辑存在以下问题：

- MCPGroupManager 只是简单存储 handler 映射，不按需构建
- 缺少高效的缓存策略
- VAuthKey 到 BuildInfo 的映射缺乏缓存层

**本次重构目标**：

1. MCPGroupManager 通过 vauth\_key 获取 mcp-server 的 build\_info，按需构建 handler
2. 实现三级缓存策略：内存双 LRU → Redis → MySQL
3. 泛型双 LRU 缓存库，支持 L1/L2 分层，访问频率准入 L2
4. Redis 缓存采用 Lua 脚本，携带版本号保证原子性
5. build\_info 变更时，事务更新版本号并删除 Redis 缓存

***

## 2. 架构设计

### 2.1 整体架构

```
请求 /mcp/:vauth_key
    │
    ▼
┌─────────────────────────────────────────────────────────────┐
│                    MCPGroupManager                          │
│  ┌─────────┐    ┌─────────┐                                │
│  │   L2    │───▶│   L1    │  (build_uuid → handler)        │
│  │ 热点缓存 │    │ 普通缓存 │   L2: 1秒内≥2次访问准入        │
│  └─────────┘    └─────────┘                                │
└─────────────────────────────────────────────────────────────┘
    │ miss              │ miss
    ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│              BuildInfoCacheService (Redis)                  │
│  vauth_key → {build_uuid, version}                          │
│  Lua 脚本写入，携带版本号                                   │
└─────────────────────────────────────────────────────────────┘
    │ miss (或无缓存)
    ▼
┌─────────────────────────────────────────────────────────────┐
│                       MySQL                                │
│  mcp_servers (vauth_key 关联)                               │
│  server_build_info (build_uuid, build_data)                 │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 调用流程（序列图）

```
┌─────────┐     ┌──────────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│ Client  │     │ScopedMCPHandler│     │ Manager │     │  L1/L2  │     │  Redis  │     │  MySQL  │
└────┬────┘     └──────┬───────┘     └────┬────┘     └────┬────┘     └────┬────┘     └────┬────┘
     │                 │                 │               │               │               │
     │ GET /mcp/:vkey  │                 │               │               │               │
     │────────────────▶│                 │               │               │               │
     │                 │                 │               │               │               │
     │                 │ GetHandler(vkey)               │               │               │
     │                 │───────────────▶│               │               │               │
     │                 │                 │               │               │               │
     │                 │                 │ L2.Get(uuid)  │               │               │
     │                 │                 │──────────────▶│               │               │
     │                 │                 │◀──────────────│ (miss)        │               │
     │                 │                 │               │               │               │
     │                 │                 │ L1.Get(uuid)  │               │               │
     │                 │                 │──────────────▶│               │               │
     │                 │                 │◀──────────────│ (miss)        │               │
     │                 │                 │               │               │               │
     │                 │                 │ GetBuildUUID(vkey)             │               │
     │                 │                 │──────────────────────────────▶│               │
     │                 │                 │◀──────────────────────────────│ (miss)        │
     │                 │                 │               │               │               │
     │                 │                 │ GetByVAuthKey(vkey)           │               │
     │                 │                 │──────────────────────────────────────────────▶│
     │                 │                 │◀──────────────────────────────────────────────│
     │                 │                 │               │               │               │
     │                 │                 │ [构建 handler]│               │               │
     │                 │                 │               │               │               │
     │                 │                 │ L1.Set(uuid, handler)        │               │
     │                 │                 │──────────────▶│               │               │
     │                 │                 │               │               │               │
     │                 │                 │◀──────────────│ (回填成功)    │               │
     │                 │                 │               │               │               │
     │                 │◀───────────────│ (handler)     │               │               │
     │                 │                 │               │               │               │
     │◀────────────────│ (200 OK)       │               │               │               │
```

### 2.3 缓存命中流程（序列图）

```
┌─────────┐     ┌──────────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐
│ Client  │     │ScopedMCPHandler│     │ Manager │     │  L1/L2  │     │  Redis  │
└────┬────┘     └──────┬───────┘     └────┬────┘     └────┬────┘     └────┬────┘
     │                 │                 │               │               │
     │ GET /mcp/:vkey  │                 │               │               │
     │────────────────▶│                 │               │               │
     │                 │                 │               │               │
     │                 │ GetHandler(vkey)               │               │
     │                 │───────────────▶│               │               │
     │                 │                 │               │               │
     │                 │                 │ L2.Get(uuid)  │ (热点缓存)   │
     │                 │                 │──────────────▶│               │
     │                 │                 │◀──────────────│ (hit! handler)│
     │                 │                 │               │               │
     │                 │◀───────────────│ (handler)     │               │
     │                 │                 │               │               │
     │◀────────────────│ (200 OK)       │               │               │
```

### 2.4 缓存失效流程（序列图）

```
┌────────────┐     ┌──────────────┐     ┌─────────┐     ┌─────────┐
│ToolBinding │     │BuildService  │     │  Redis  │     │  MySQL  │
│   Service  │     │              │     │         │     │         │
└─────┬──────┘     └──────┬───────┘     └────┬────┘     └────┬────┘
      │                  │                   │               │
      │ BindTool(...)   │                   │               │
      │────────────────▶│                   │               │
      │                  │                   │               │
      │                  │ BuildOrUpdate()  │               │
      │                  │──┐                │               │
      │                  │  │ Start tx       │               │
      │                  │────────────────▶│               │
      │                  │◀────────────────│               │
      │                  │  │               │               │
      │                  │ GetActiveBuild  │               │
      │                  │────────────────────────────────▶│
      │                  │◀────────────────────────────────│
      │                  │  │               │               │
      │                  │ CreateNewBuild │               │
      │                  │────────────────────────────────▶│
      │                  │◀────────────────────────────────│
      │                  │  │               │               │
      │                  │ Commit tx      │               │
      │                  │───────────────────────────────▶│
      │                  │  │               │               │
      │                  │ InvalidateVAuthKey(vkey)      │
      │                  │────────────────▶│               │
      │                  │◀────────────────│ (DEL)        │
      │                  │  │               │               │
      │◀─────────────────│ (成功)          │               │
```

### 2.5 LRU 内部流程

```
┌─────────────────────────────────────────────────────────────────┐
│                    TwoLevelLRU[K, V]                           │
│                                                                 │
│  Get(key) {                                                     │
│    1. 检查 L2  ──────────────────────────────────────────┐     │
│       │ hit: 返回 value                                    │     │
│       ▼                                                      │     │
│    2. 检查 L1  ──────────────────────────────────────────┐│     │
│       │ hit: 返回 value                                   ││     │
│       ▼                                                     ││     │
│    3. 返回 not found                                       ││     │
│                                                               ││     │
│    4. [miss时外部加载数据后调用 Set(key, value)]            ││     │
│       ├─▶ L1.Set(key, value)  ──▶ L1 容量满? 淘汰尾部     ││     │
│       │                                                         ││
│       └─▶ AccessTracker.RecordAccess(key)                      │
│               │ 1秒内≥2次?                                     │
│               ▼                                                │
│           L2.Set(key, value)  ──▶ L2 容量满? 淘汰尾部          │
│  }                                                              │
└─────────────────────────────────────────────────────────────────┘
```

***

## 3. 泛型 LRU 缓存库设计

### 3.1 文件位置

```
internal/common/cache/
├── lru.go           # 基础泛型 LRU
└── two_level_lru.go # 双 LRU（依赖基础 LRU）
```

### 3.2 基础泛型 LRU

单级 LRU，支持泛型 K/V。

```go
package cache

type LRU[K comparable, V any] struct {
    capacity int
    items    map[K]V
    order    []K // 维护顺序
    mu       sync.RWMutex
}

func NewLRU[K comparable, V any](capacity int) *LRU[K, V]

func (l *LRU[K, V]) Get(key K) (V, bool)
func (l *LRU[K, V]) Set(key K, value V)
func (l *LRU[K, V]) Delete(key K)
func (l *LRU[K, V]) Contains(key K) bool
func (l *LRU[K, V]) Len() int
func (l *LRU[K, V]) Clear()
```

### 3.3 双 LRU

基于两个基础 LRU 实现，支持访问频率准入 L2。

```go
package cache

type TwoLevelLRU[K comparable, V any] struct {
    l1     *LRU[K, V]      // 依赖基础 LRU
    l2     *LRU[K, V]      // 依赖基础 LRU
    access *AccessTracker[K]
    config Config
}

func NewTwoLevelLRU[K comparable, V any](config Config) *TwoLevelLRU[K, V]
```

### 3.4 LRU 准入规则

| 层级     | 准入条件     | 说明          |
| ------ | -------- | ----------- |
| **L1** | 无条件      | 任何数据首次进入 L1 |
| **L2** | 1秒内访问≥2次 | 只有热点数据进入 L2 |

### 3.5 访问计数器

```go
type AccessTracker[K comparable] struct {
    mu         sync.RWMutex
    timestamps map[K][]time.Time
    threshold int           // 准入阈值：1秒内访问次数
    window    time.Duration // 时间窗口
}

func (t *AccessTracker[K]) RecordAccess(key K) bool {
    now := time.Now()
    t.mu.Lock()
    defer t.mu.Unlock()

    cutoff := now.Add(-t.window)
    t.timestamps[key] = filterRecent(t.timestamps[key], cutoff)
    t.timestamps[key] = append(t.timestamps[key], now)

    return len(t.timestamps[key]) >= t.threshold
}
```

***

## 4. BuildInfo 缓存服务设计

### 4.1 文件位置

```
internal/service/build_info_cache_service.go
```

### 4.2 Redis 数据结构

**Key**: `mcp:vauth:{vauth_key}`
**Value**: `{"build_uuid": "xxx", "version": 1}`
**TTL**: 5分钟

### 4.3 Lua 脚本（原子性写入带版本号）

```lua
-- SETNX with version check
local key = KEYS[1]
local value = cjson.decode(ARGV[1])
local expectedVersion = tonumber(ARGV[2])

local existing = redis.call('GET', key)
if existing then
    local existingValue = cjson.decode(existing)
    if existingValue.version > expectedVersion then
        -- 发现更高版本，不覆盖
        return 0
    end
end

redis.call('SET', key, ARGV[1], 'EX', 300)
return 1
```

### 4.4 接口设计

```go
type BuildInfoCacheService struct {
    redis    *redis.Client
    buildDAO repository.ServerBuildInfoStore
}

type CachedBuildUUID struct {
    BuildUUID string `json:"build_uuid"`
    Version   int    `json:"version"`
}

func (s *BuildInfoCacheService) GetBuildUUID(ctx context.Context, vauthKey string) (*CachedBuildUUID, error)
func (s *BuildInfoCacheService) SetBuildUUID(ctx context.Context, vauthKey string, uuid *CachedBuildUUID) error
func (s *BuildInfoCacheService) DeleteBuildUUID(ctx context.Context, vauthKey string) error
func (s *BuildInfoCacheService) InvalidateVAuthKey(ctx context.Context, vauthKey string) error
```

### 4.5 缓存失效策略

- **变更触发**：build\_info 变更时，调用 `InvalidateVAuthKey` 删除 Redis 缓存
- **TTL兜底**：5分钟自然过期
- **L1/L2**：只靠 LRU 自然淘汰，不主动失效

***

## 5. MCPGroupManager 重构设计

### 5.1 文件位置

```
internal/service/group_manager.go
```

### 5.2 结构设计

```go
type MCPGroupManager struct {
    serverName    string
    serverVersion string
    authService  *AuthService

    // 依赖
    buildInfoCache *BuildInfoCacheService
    buildInfoStore repository.ServerBuildInfoStore
    serverStore    repository.MCPServerStore

    // 双 LRU 缓存 (build_uuid → handler)
    cache *cache.TwoLevelLRU[string, http.Handler]

    mu sync.RWMutex
}
```

### 5.3 Handler 获取流程

```go
func (m *MCPGroupManager) GetHandler(vauthKey string) (http.Handler, error) {
    // 1. 尝试从缓存获取
    cached, ok := m.cache.Get(vauthKey)
    if ok {
        return cached, nil
    }

    // 2. 从 Redis/MySQL 获取 build_info
    buildInfo, err := m.loadBuildInfo(vauthKey)
    if err != nil {
        return nil, err
    }
    if buildInfo == nil {
        return nil, ErrBuildInfoNotFound
    }

    // 3. 构建 handler
    handler, err := m.buildHandler(buildInfo)
    if err != nil {
        return nil, err
    }

    // 4. 回填缓存
    m.cache.Set(buildInfo.BuildUUID, handler)

    return handler, nil
}
```

### 5.4 Handler 构建逻辑

```go
func (m *MCPGroupManager) buildHandler(info *model.ServerBuildInfo) (http.Handler, error) {
    var buildData model.BuildData
    if err := sonic.Unmarshal([]byte(info.BuildData), &buildData); err != nil {
        return nil, fmt.Errorf("unmarshal build_data: %w", err)
    }

    // 转换为 server.Tool
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

    // 创建 MCP Server
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
```

***

## 6. 缓存失效设计

### 6.1 触发场景

- 工具绑定到 server
- 工具从 server 解绑
- 工具定义更新
- MCP Server 配置变更

### 6.2 失效流程

```go
func (s *ServerBuildService) BuildOrUpdate(ctx context.Context, serverID uint) error {
    // 1. 获取 server 信息
    server, err := s.serverStore.GetByID(ctx, serverID)
    if err != nil {
        return err
    }

    // 2. 构建新版本（事务内）
    newBuild, err := s.createNewBuild(ctx, serverID)
    if err != nil {
        return err
    }

    // 3. 失效缓存
    if err := s.buildInfoCache.InvalidateVAuthKey(ctx, server.VAuthKey); err != nil {
        // 日志记录，不阻塞主流程
        log.Printf("invalidate cache failed: %v", err)
    }

    return nil
}
```

### 6.3 版本号管理

- 每次 `BuildOrUpdate` 生成新版本，version + 1
- Redis Lua 脚本检查版本号，高版本不覆盖低版本
- 防止短时并发更新导致缓存覆盖

***

## 7. 数据模型

### 7.1 ServerBuildInfo

```go
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
```

### 7.2 BuildData (存储在 BuildData JSON 字段中)

```go
type BuildData struct {
    Tools       []ToolSnapshot       `json:"tools"`
    HTTPServices []HTTPServiceSnapshot `json:"http_services"`
}

type ToolSnapshot struct {
    ID          uint   `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Parameters  []byte `json:"parameters"`
    Enabled     bool   `json:"enabled"`
}
```

***

## 8. 文件结构

```
internal/
├── common/
│   └── cache/
│       ├── lru.go              # 新增：基础泛型 LRU
│       └── two_level_lru.go    # 新增：双 LRU（依赖基础 LRU）
├── service/
│   ├── group_manager.go        # 重构：MCPGroupManager
│   ├── build_info_cache.go     # 新增：BuildInfo 缓存服务
│   └── server_build_service.go # 新增/已有：构建更新服务
└── domain/
    └── model/
        └── server_build_info.go  # 已有：构建信息模型

docs/superpowers/specs/
└── 2026-04-20-mcp-vauth-cache-refactor-design.md  # 本文档
```

***

## 9. 测试要点

1. **泛型 LRU 测试**
   - L1 容量满后正确淘汰
   - L2 只有满足频率条件才准入
   - 并发访问计数准确
2. **缓存服务测试**
   - Redis Lua 脚本版本号检查
   - 缓存 miss 后正确回填
   - 缓存失效后正确删除
3. **MCPGroupManager 测试**
   - 404 场景：vauth\_key 不存在
   - 缓存命中场景：直接返回 handler
   - 缓存 miss 场景：构建 handler 并回填
   - 并发场景：相同 vauth\_key 只构建一次
4. **缓存失效测试**
   - build\_info 变更后 Redis 缓存被删除
   - L1/L2 自然淘汰，不影响新请求

***

## 10. 配置项

### 10.1 LRU 内存缓存配置

```go
type LRUCacheConfig struct {
    L1Size     int           // L1 缓存大小，默认 2000
    L2Size     int           // L2 缓存大小，默认 2000
    L2Window   time.Duration // L2 准入时间窗口，默认 1 秒
    L2Threshold int           // L2 准入访问次数，默认 2
}
```

### 10.2 Redis 缓存配置

```go
type RedisCacheConfig struct {
    TTL time.Duration // 缓存 TTL，默认 5 分钟
}
```

### 10.3 MCPGroupManager 配置

```go
type MCPGroupManagerConfig struct {
    Cache LRUCacheConfig  // 内存 LRU 配置
    Redis RedisCacheConfig // Redis 配置
}
```

