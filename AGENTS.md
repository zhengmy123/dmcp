# AGENTS 规范与约束

本文档定义了 Dynamic MCP-Go Server 项目中 AI Agent 开发和协作的规范与约束。

## 目录
- [项目架构](#项目架构)
- [代码规范](#代码规范)
- [开发流程](#开发流程)
- [测试规范](#测试规范)
- [文档规范](#文档规范)

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
    Enabled   *bool
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

---

## 联系方式

如有问题，请参考项目现有代码和文档，或与项目维护者联系。
