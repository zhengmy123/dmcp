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
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (User) TableName() string {
    return "mcp_users"
}
```

#### 4. 函数/方法参数
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

#### 5. Infrastructure 层查询
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
- 接口修改后要及时更新接口文档

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
- 修改接口后**必须**及时更新 `docs/api/` 下的对应文档
- 文档格式保持一致

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

---

## 联系方式

如有问题，请参考项目现有代码和文档，或与项目维护者联系。
