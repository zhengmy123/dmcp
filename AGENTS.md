# AGENTS 规范与约束

本文档定义了 Dynamic MCP-Go Server 项目中 AI Agent 开发和协作的规范与约束。

## 目录
- [项目架构](#项目架构)
  - [DDD 分层架构](#ddd-分层架构)
  - [关键技术栈](#关键技术栈)
- [代码规范](#代码规范)
  - [Go 代码规范](#go-代码规范)
    - [1. 错误处理](#1-错误处理)
    - [2. JSON 处理](#2-json-处理)
    - [3. 数据库操作](#3-数据库操作)
    - [4. 软删除](#4-软删除)
    - [5. 批量软删除操作规范](#5-批量软删除操作规范)
    - [6. 函数/方法参数](#6-函数方法参数)
    - [7. Infrastructure 层查询](#7-infrastructure-层查询)
    - [8. 事务处理](#8-事务处理)
    - [9. 统一返回值规范](#9-统一返回值规范)
    - [10. 指针参数校验](#10-指针参数校验)
    - [11. MCP 调试规范](#11-mcp-调试规范)
  - [前端代码规范](#前端代码规范)
    - [1. 项目结构](#1-项目结构)
    - [2. 接口文档](#2-接口文档)
    - [3. UI 规范](#3-ui-规范)
- [开发流程](#开发流程)
  - [Git 工作流](#git-工作流)
  - [开发步骤](#开发步骤)
- [测试规范](#测试规范)
  - [测试目录结构](#测试目录结构)
  - [E2E 测试配置](#e2e-测试配置)
  - [测试编写规范](#测试编写规范)
- [文档规范](#文档规范)
  - [文档目录](#文档目录)
  - [接口文档更新](#接口文档更新)

---

## 项目架构

### DDD 分层架构

项目遵循领域驱动设计 (DDD) 模式，主要分层如下：

```
cmd/server/          # 应用入口
├── main.go          # 服务启动入口

internal/
├── api/http/        # API 层
│   ├── handler/     # HTTP 处理器
│   └── router.go    # 路由配置
│
├── domain/          # 领域层
│   ├── model/       # 领域模型
│   ├── repository/  # 仓储接口
│   └── service/     # 领域服务
│
├── infrastructure/  # 基础设施层
│   ├── auth/        # 认证实现
│   ├── database/    # 数据库实现
│   └── store/       # 存储实现
│
├── service/         # 应用服务层
│
├── common/          # 公共组件
│   ├── logger/      # 日志
│   ├── middleware/  # 中间件
│   └── response/    # 响应处理
│
└── config/          # 配置管理

web/admin/           # 前端管理后台
docs/                # 文档目录
test/                # 测试目录
```

### 关键技术栈

| 技术 | 用途 | 版本 |
|------|------|------|
| Go | 后端开发 | 1.25.5 |
| Gin | Web 框架 | 1.12.0 |
| GORM | ORM | 1.25.7 |
| MySQL | 数据库 | 8.0 |
| Sonic JSON | JSON 处理 | 1.15.0 |
| JWT | 认证 | 5.3.1 |
| Vue 3 | 前端框架 | - |

---

## 代码规范

### Go 代码规范

#### 1. 错误处理
- **必须**处理所有 error，不能忽略
- 使用 `fmt.Errorf("context: %w", err)` 包装错误
- 示例：

```go
result, err := someFunction()
if err != nil {
    return nil, fmt.Errorf("failed to do something: %w", err)
}
```

#### 2. JSON 处理
- **必须**使用 `github.com/bytedance/sonic` 进行 JSON 序列化/反序列化
- 不要使用标准库 `encoding/json`

```go
import "github.com/bytedance/sonic"

// 序列化
data, err := sonic.Marshal(obj)
if err != nil {
    return err
}

// 反序列化
err = sonic.Unmarshal(data, &obj)
if err != nil {
    return err
}
```

#### 3. 数据库操作
- **必须**使用 GORM 进行数据库操作
- 模型定义使用 gorm 标签

```go
type User struct {
    ID        uint      `gorm:"primaryKey;autoIncrement"`
    Name      string    `gorm:"size:128;not null"`
    State     int       `gorm:"default:1;comment:状态 1-正常 0-删除"`
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (User) TableName() string {
    return "mcp_users"
}
```

#### 4. 软删除
- **所有删除操作都是软删除**，通过 `state` 字段来区分
- `state` 字段约定：`1` 表示正常状态，`0` 表示已删除
- 查询时默认过滤 `state = 1` 的记录
- 删除操作时将 `state` 设为 `0`，而不是真正删除记录
- 需要查询包括软删除的记录时，使用 `Unscoped()` 或单独的查询方法

#### 5. 批量软删除操作规范
- **必须**采用「先批量查询，再分三种情况处理」的模式：
  1. **已有效**（state=1）：保持不变
  2. **已失效**（state=0）：恢复或跳过
  3. **不存在**：新增

- 示例流程（批量绑定）：

```go
func (s *Service) BatchBind(ctx context.Context, req BatchRequest) (int, error) {
    // 1. 批量查询所有记录（包括软删除）
    allBindings, err := s.store.ListAllIncludeDeleted(ctx)
    if err != nil {
        return 0, err
    }

    // 2. 构建存在性Map
    existingMap := make(map[uint]map[uint]int)
    for _, b := range allBindings {
        if existingMap[b.ToolID] == nil {
            existingMap[b.ToolID] = make(map[uint]int)
        }
        existingMap[b.ToolID][b.ServerID] = b.State
    }

    // 3. 分三种情况处理
    var toRestore []uint    // 失效 -> 有效
    var toCreate []*Model   // 不存在 -> 新增

    for _, toolID := range req.ToolIDs {
        for _, serverID := range req.ServerIDs {
            state, exists := existingMap[toolID][serverID]
            if !exists {
                // 不存在：新增
                toCreate = append(toCreate, &Model{ToolID: toolID, ServerID: serverID})
            } else if state == 0 {
                // 已失效：恢复
                binding, _ := s.store.GetByToolAndServerIncludeDeleted(ctx, toolID, serverID)
                if binding != nil {
                    toRestore = append(toRestore, binding.ID)
                }
            }
            // 已有效：跳过
        }
    }

    // 4. 批量执行
    if len(toRestore) > 0 {
        s.store.BatchRestore(ctx, toRestore)
    }
    if len(toCreate) > 0 {
        s.store.BatchSave(ctx, toCreate)
    }

    return len(toRestore) + len(toCreate), nil
}
```

#### 6. 函数/方法参数
- 优先使用结构体作为函数和方法的入参和出参
- 避免使用过多的独立参数

```go
// 推荐
type CreateToolRequest struct {
    Name        string
    Description string
    Parameters  []byte
}

func CreateTool(req *CreateToolRequest) (*Tool, error) {
    // ...
}

// 不推荐
func CreateTool(name string, description string, parameters []byte) (*Tool, error) {
    // ...
}
```

#### 7. Infrastructure 层查询
- Infrastructure 层的查询是泛化的
- 传入的 query 参数是结构体指针
- 根据结构体的字段进行查询

```go
type ToolQuery struct {
    ID        *uint
    Name      *string
    ServiceID *uint
    State     *int
}

func (s *GORMToolStore) List(ctx context.Context, query *ToolQuery) ([]*model.Tool, error) {
    db := s.db.WithContext(ctx)

    if query.ID != nil {
        db = db.Where("id = ?", *query.ID)
    }
    if query.Name != nil {
        db = db.Where("name = ?", *query.Name)
    }
    // ... 其他条件

    var tools []*model.Tool
    err := db.Find(&tools).Error
    return tools, err
}
```

#### 8. 事务处理
- **多表联合变更必须在一个事务中执行**，确保数据一致性
- 使用 GORM 的 `Transaction` 方法管理事务
- 示例：

```go
func (s *Service) MultiTableOperation(ctx context.Context, req *Request) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 操作表1
        if err := tx.Create(&record1).Error; err != nil {
            return fmt.Errorf("failed to create record1: %w", err)
        }

        // 操作表2
        if err := tx.Where("id = ?", req.ID).Update("status", req.Status).Error; err != nil {
            return fmt.Errorf("failed to update record2: %w", err)
        }

        // 操作表3
        if err := tx.Delete(&record3).Error; err != nil {
            return fmt.Errorf("failed to delete record3: %w", err)
        }

        return nil
    })
}
```

- 事务回滚规则：
  - 任何一步操作失败，整个事务自动回滚
  - 不要在事务中进行不必要的长时间操作
  - 确保事务范围最小化，避免锁表时间过长
  - **推荐使用 `defer` 方式控制回滚**，避免事务处理函数中漏掉错误返回时的回滚逻辑

```go
func (s *Service) MultiTableOperation(ctx context.Context, req *Request) error {
    tx := s.db.WithContext(ctx).Begin()
    if tx.Error != nil {
        return fmt.Errorf("failed to begin transaction: %w", tx.Error)
    }

    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    if err := tx.Create(&record1).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to create record1: %w", err)
    }

    if err := tx.Where("id = ?", req.ID).Update("status", req.Status).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to update record2: %w", err)
    }

    if err := tx.Delete(&record3).Error; err != nil {
        tx.Rollback()
        return fmt.Errorf("failed to delete record3: %w", err)
    }

    return tx.Commit().Error
}
```

- `defer` 回滚方式优势：
  - 代码更清晰，事务边界明确
  - 避免在每个错误分支手动调用 `Rollback()`
  - `defer` 中的 `recover()` 可以捕获 panic，确保事务不会卡住
  - 与 GORM 的 `Transaction` 方法相比，**更推荐使用手动事务 + defer 方式**，因为 GORM 的 `Transaction` 方法在返回错误时内部已自动调用回滚，但如果在外部有额外清理逻辑时不如手动事务灵活

#### 9. 统一返回值规范
- MCP 接口遵守 [MCP 协议规范](../docs/api/mcp-protocol.md)
- 格式定义在 `internal/common/response/response.go` 中
- 响应结构：

```json
{
    "code": 0,
    "message": "success",
    "detail": "...",  // 仅在错误时存在
    "data": {...}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 状态码，0 表示成功，非 0 表示错误 |
| message | string | 状态信息，成功时为 "success" |
| detail | string | 错误详情，仅在错误时返回 |
| data | object | 响应数据，成功时返回 |

- 常用响应方法：
  - `response.Success(c, data)` - 成功响应
  - `response.SuccessWithMessage(c, message, data)` - 带自定义消息的成功响应
  - `response.Created(c, data)` - 创建成功响应
  - `response.BadRequest(c, message, detail...)` - 400 错误
  - `response.Unauthorized(c, message)` - 401 错误
  - `response.Forbidden(c, message)` - 403 错误
  - `response.NotFound(c, message)` - 404 错误
  - `response.Conflict(c, message)` - 409 错误
  - `response.InternalError(c, message)` - 500 错误

- 错误码定义：

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 |

#### 10. 指针参数校验
- **必须**校验所有指针类型参数是否为 `nil`
- 在使用指针参数前，必须先判断是否为 `nil`，避免空指针 panic
- 示例：

```go
func ProcessData(req *ProcessRequest) (*Response, error) {
    if req == nil {
        return nil, fmt.Errorf("request cannot be nil")
    }

    // 对指针类型的字段进行 nil 检查
    if req.Filter != nil && req.Filter.ID == nil {
        return nil, fmt.Errorf("filter.id cannot be nil when filter is provided")
    }

    // ... 业务逻辑
}
```

- 特殊场景：
  - 接口入口层（handler）必须对指针参数进行校验
  - 领域服务层在调用下层方法时，可根据业务逻辑判断是否需要校验
  - 基础设施层（store）对于泛化查询结构体指针，内部已有 nil 判断逻辑，但仍需对关键字段进行校验

#### 11. MCP 调试规范
- **MCP 调试时统一使用 POST 方法**
- 禁止使用 GET 方法进行 MCP 接口调试
- 所有 MCP 协议相关的请求都必须通过 POST 方式发送 JSON 请求体
- 示例：

```bash
# 正确：使用 POST
curl -X POST http://localhost:17050/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'

# 错误：禁止使用 GET
curl -X GET http://localhost:17050/mcp?method=tools/list
```

- 原因：MCP 协议基于 JSON-RPC 2.0 规范，请求参数通常包含复杂的 JSON 对象，GET 方法无法正确传递请求体

### 前端代码规范

#### 1. 项目结构
```
web/admin/src/
├── api/           # API 接口
├── components/    # 组件
├── layouts/       # 布局
├── pages/         # 页面
├── router/        # 路由
├── stores/        # 状态管理 (Pinia)
├── styles/        # 样式
├── types/         # 类型定义
└── utils/         # 工具函数
```

#### 2. 接口文档
- 编写页面功能时参考 `docs/api/` 下的接口文档
- **修改接口后必须立即同步更新接口文档**
- 接口文档应包含：请求参数、响应格式、字段说明
- 保持文档格式一致

#### 3. UI 规范
- **错误提示要在最上层显眼的位置**：错误消息应显示在页面或组件的顶部，使用醒目的颜色（如红色）并带有明显的视觉样式，确保用户第一时间能看到错误信息

---

## 开发流程

### Git 工作流
- Git 不需要每次修改都自动提交，由开发者自行判断
- 提交前确保代码符合规范
- 提交信息清晰描述变更内容

### 开发步骤
1. 阅读相关文档（`docs/api/`、`docs/architecture.md`）
2. 遵循 DDD 分层架构设计
3. 编写代码（遵循代码规范）
4. 编写测试（放在 `test/` 目录）
5. 更新文档（如需要）
6. 运行测试确保通过

---
**禁止规则：**

```
❌ 不得在 models/ 中写业务逻辑
❌ 不得在 handler 中写业务逻辑
❌ 不得跨模块直接调用（auth → org 直接引用）
❌ 不得返回裸 err（必须 wrapping）
❌ db.AutoMigrate 不得提交到代码库（schema 由 docs/sql/schema.sql 管理）
```
---
## 测试规范

### 测试目录结构
- **所有测试**必须放在项目 `test/` 目录下
- 测试文件以 `_test.go` 结尾
- 保持与源代码相同的包结构

```
test/
├── domain/
│   └── model/
│       └── tool_server_binding_test.go
├── infrastructure/
│   └── store/
│       └── tooldef/
│           └── types_test.go
└── script_validator_test.go
```

### E2E 测试配置
- **前端服务端口**：`17000`（本地开发环境，不经过 nginx）
- **后端服务端口**：`17050`（本地开发环境）

### 测试编写规范
- 使用 Go 标准测试框架 `testing`
- 测试函数命名：`TestXxx`
- 每个测试函数独立运行

---

## 文档规范

### 文档目录
- `docs/api/` - API 接口文档
- `docs/architecture.md` - 架构文档
- `docs/plans/` - 开发计划
- `docs/superpowers/` - 功能规格和计划

### 接口文档更新
- 修改接口后**必须立即同步更新** `docs/api/` 下的对应文档
- 文档格式保持一致
- 文档应包含完整的请求参数、响应格式、字段说明

---

## 项目规则总结

1. ✅ 所有测试都要在项目 test 目录下
2. ✅ 项目遵循 DDD 模式
3. ✅ 项目中查询数据库 golang 的都使用 gorm
4. ✅ 函数、方法出入参优先使用结构体
5. ✅ infrastructure 层的查询是泛化的，传入的 query 参数是结构体，根据结构体的字段进行查询，参数类型是结构体指针
6. ✅ git 不需要每次修改都自动提交，自己判断是否需要提交
7. ✅ 编写 golang 代码时，都不能忽略和不处理 error
8. ✅ golang json 库使用 sonic json
9. ✅ 编写页面功能时可以参考下项目中 docs/api/ 的接口文档，接口修改后要及时更新接口文档
10. ✅ 错误提示要在最上层显眼的位置
11. ✅ 所有删除都是软删除，通过 state 字段来区分（1-正常，0-删除）
12. ✅ 批量软删除操作采用「先批量查询，再分三种情况处理」的模式
13. ✅ 所有 MySQL 表必须包含 id（BIGINT 主键）、created_at（创建时间）、updated_at（更新时间）
14. ✅ Hash 计算必须是有序的，序列化前需对数据进行排序，确保同样数据计算结果一致
15. ✅ 数据表字段变更规范：每次修改功能时，如果涉及数据表字段的变更，必须保证 model 和数据表映射完整，**必须立即更新 `mysql_migration.sql`**，同时生成的 SQL 必须保证 MySQL 8.0 可以正常运行
16. ✅ 统一返回值规范：所有 API 响应必须使用 `{code, message, detail, data}` 格式，detail 仅在错误时返回（MCP 接口除外，MCP 接口遵守 MCP 协议）
17. ✅ 多表联合变更必须在一个事务中执行，确保数据一致性
18. ✅ 指针参数校验：所有指针类型参数在使用前必须校验是否为 nil，避免空指针 panic
19. ✅ MCP 调试规范：MCP 调试时统一使用 POST 方法，禁止使用 GET 方法

---

## 联系方式

如有问题，请参考项目现有代码和文档，或与项目维护者联系。
