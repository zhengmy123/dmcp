# MCP VAuth Key 调用逻辑和工具绑定逻辑重构设计

**日期**: 2026-04-19

## 1. 背景与目标

当前项目中的 MCP 服务按 vauthKey 聚合工具，但存在以下问题：
- `ListDefinitionsByVAuthKey` 和 `GetDefinitionByVAuthKey` 是空实现
- 工具绑定后没有构建信息记录
- 工具变动时没有触发构建更新
- 工具删除时没有校验绑定关系

**本次重构目标**：
1. 工具绑定 mcp-server 时，不允许删除工具（必须先解绑）
2. 新增构建信息表，记录 mcp-go server 和 tool 的快照信息
3. 一个 mcp-server 同一时刻只有一个有效的构建信息
4. 工具变动时，更新构建信息并生成新版本

---

## 2. 数据模型

### 2.1 新增表：server_build_info

```sql
CREATE TABLE server_build_info (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    server_id BIGINT NOT NULL COMMENT '关联 mcp_servers.id',
    version INT NOT NULL DEFAULT 1 COMMENT '版本号，每次构建+1',
    build_uuid VARCHAR(36) NOT NULL COMMENT '构建UUID，用于唯一标识',
    hash VARCHAR(64) NOT NULL COMMENT '工具列表的 SHA256 哈希值，用于变更检测',
    build_data TEXT COMMENT 'JSON: 工具定义和HTTP服务完整定义的快照合并',
    state INT NOT NULL DEFAULT 1 COMMENT '状态 1-有效 0-失效',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_build_uuid (build_uuid),
    INDEX idx_hash (hash),
    INDEX idx_server_state (server_id, state)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='MCP Server 构建信息表';
```

**约束**：
- `build_uuid` 全局唯一，用于唯一标识每次构建
- `hash` 用于判断是否有变更，配合索引快速查询
- `build_data` 合并存储 mcp_tools 和 http_services 的 JSON 快照
- 同一 server_id 同时只有一条 state=1 的记录
- version 在同一 server_id 内自增

### 2.2 表结构（Go Model）

```go
type ServerBuildInfo struct {
    ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
    ServerID  uint      `json:"server_id" gorm:"not null;index:idx_server_state"`
    Version   int       `json:"version" gorm:"not null;default:1"`
    BuildUUID string    `json:"build_uuid" gorm:"size:36;not null;uniqueIndex"`
    Hash      string    `json:"hash" gorm:"size:64;not null;index"`
    BuildData string    `json:"build_data" gorm:"type:text"`
    State     int       `json:"state" gorm:"not null;default:1;index:idx_server_state"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ServerBuildInfo) TableName() string {
    return "server_build_info"
}
```

**BuildData JSON 结构**：
```json
{
  "tools": [...],
  "http_services": [...]
}
```

---

## 3. 规范约束

### 3.1 MySQL 表规范

**所有 MySQL 表必须包含**：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | BIGINT | 主键，自增 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 3.2 Hash 计算规范

**Hash 计算必须是有序的**：
- 序列化前对数据进行排序（如按 tool.name 排序）
- 使用稳定的序列化方式（sonic.Marshal）
- 示例：
  ```go
  sort.Slice(tools, func(i, j int) bool {
      return tools[i].Name < tools[j].Name
  })
  data, _ := sonic.Marshal(tools)
  sum := sha256.Sum256(data)
  hash := hex.EncodeToString(sum[:])
  ```

---

## 4. 功能设计

### 4.1 工具删除校验

**位置**: `ToolDomainService` 或 `ToolService`

**逻辑**：
1. 删除工具前，检查 `tool_server_bindings` 是否存在有效绑定（state=1）
2. 如有绑定，返回错误 `ErrToolHasActiveBinding`
3. 只有无绑定或绑定已软删除时，才允许删除工具

**错误定义**：
```go
var ErrToolHasActiveBinding = errors.New("tool has active binding, unbind first")
```

---

### 4.2 工具绑定时的构建更新

**位置**: `ToolBindingService`

**修改方法**：
- `BindTool()`
- `BatchBindTools()`

**逻辑**：
1. 绑定成功后，调用 `ServerBuildService.BuildOrUpdate(serverID)`
2. 如果绑定操作涉及多个 server，需分别调用

---

### 4.3 构建更新服务

**新建**: `ServerBuildService`

```go
type ServerBuildService struct {
    serverStore    repository.MCPServerStore
    toolStore      repository.ToolStore
    bindingStore   repository.ToolServerBindingStore
    buildInfoStore repository.ServerBuildInfoStore
}

func (s *ServerBuildService) BuildOrUpdate(ctx context.Context, serverID uint) error
```

**BuildOrUpdate 流程**：
1. 获取 server 关联的所有有效绑定
2. 获取每个绑定关联的工具及其 HTTP 服务
3. 生成 mcp_tools 和 http_services 快照
4. 计算 hash（有序序列化后 SHA256）
5. 如 hash 与当前 active 版本相同，跳过
6. 否则：
   - 将旧 active 版本 state 设为 0
   - 创建新版本记录（version+1），state=1

---

### 4.4 工具变动时的构建更新

**位置**: `ToolService`

**触发场景**：
- 创建工具（CreateToolFromHTTPService）
- 更新工具参数
- 更新工具的 HTTP 服务配置（service_id 变更）
- 删除工具（解绑后）

**逻辑**：
1. 上述操作成功后，查找该工具关联的所有 server
2. 对每个 server 调用 `ServerBuildService.BuildOrUpdate`

---

### 4.5 VAuth Key 调用逻辑重构

**位置**: `DynamicRegistry`

**修改方法**：
- `ListDefinitionsByVAuthKey(vauthKey string)`
- `GetDefinitionByVAuthKey(vauthKey, name string)`

**新逻辑**：
1. 通过 vauthKey 查找 mcp_server
2. 查找 mcp_server 对应的 active build info（state=1）
3. 从 build info 的 mcp_tools 解析返回工具定义

**不再实时查询** bindings + servers，改为从构建信息表获取。

---

## 5. 服务分层

### 5.1 Domain 层

**新增**：
- `domain/model/server_build_info.go`
- `domain/repository/server_build_info_repository.go`（接口）

### 5.2 Infrastructure 层

**新增**：
- `infrastructure/database/gorm_server_build_info.go`
- 实现 `ServerBuildInfoStore` 接口

### 5.3 Service 层

**新增**：
- `service/server_build_service.go`
- 实现 `BuildOrUpdate` 方法

**修改**：
- `service/tool_binding_service.go` - 绑定后触发构建更新
- `service/tool_service.go` - 工具变动后触发构建更新
- `service/registry.go` - 使用构建信息表查询工具

---

## 6. 影响范围

| 文件 | 操作 | 说明 |
|------|------|------|
| `domain/model/server_build_info.go` | 新增 | 构建信息模型 |
| `domain/repository/server_build_info_repository.go` | 新增 | 仓储接口 |
| `infrastructure/database/gorm_server_build_info.go` | 新增 | GORM 实现 |
| `service/server_build_service.go` | 新增 | 构建更新服务 |
| `service/tool_binding_service.go` | 修改 | 绑定后触发构建 |
| `service/tool_service.go` | 修改 | 工具变动触发构建 |
| `service/registry.go` | 修改 | 使用构建信息表查询 |
| `domain/service/tool_domain_service.go` | 修改 | 增加删除校验 |
| `docs/mysql_migration.sql` | 修改 | 添加建表语句 |

---

## 7. 迁移脚本

```sql
-- 新增 server_build_info 表
CREATE TABLE IF NOT EXISTS server_build_info (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    server_id BIGINT NOT NULL COMMENT '关联 mcp_servers.id',
    version INT NOT NULL DEFAULT 1 COMMENT '版本号',
    build_uuid VARCHAR(36) NOT NULL COMMENT '构建UUID',
    hash VARCHAR(64) NOT NULL COMMENT 'SHA256',
    build_data TEXT COMMENT 'JSON: 工具和HTTP服务快照合并',
    state INT NOT NULL DEFAULT 1 COMMENT '1-有效 0-失效',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_build_uuid (build_uuid),
    INDEX idx_hash (hash),
    INDEX idx_server_state (server_id, state)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='MCP Server 构建信息表';
```

---

## 8. 测试要点

1. **工具删除校验**：有绑定时删除应返回错误
2. **绑定后构建**：绑定工具后应生成新构建版本
3. **Hash 有序性**：相同工具列表不同顺序应生成相同 hash
4. **单版本约束**：同一 server 同时只有一条 active 构建
5. **VAuth 查询**：通过 vauthKey 能正确返回工具列表
