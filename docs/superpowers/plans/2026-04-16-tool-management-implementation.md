# MCP Server 工具管理功能实现计划

&gt; **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 MCP Server 工具管理功能，包括类型不可修改、从 HTTP Service 创建工具等。

**Architecture:** 遵循 DDD 分层架构，Handler → Service → Repository，业务规则在 Service/Domain 层。

**Tech Stack:** Go, Gin, GORM

---

## 任务分解

### Task 1: 添加服务端类型不可修改约束

**Files:**
- Modify: `internal/service/mcp_server_service.go`
- Modify: `internal/api/http/handler/mcp/mcp_server_handler.go`

- [ ] **Step 1.1: 在 service 层添加错误定义和校验**
在 `internal/service/mcp_server_service.go` 添加错误定义，然后修改 `UpdateServer` 函数校验类型不可修改。

```go
var (
    // ... 已有错误
    ErrServerTypeCannotBeChanged = errors.New("server type cannot be changed after creation")
)

// UpdateServer 更新 MCPServer
func (s *MCPServerService) UpdateServer(ctx context.Context, server *model.MCPServer) error {
    // 检查是否存在
    existing, err := s.serverStore.GetByID(ctx, server.ID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrMCPServerNotFound
        }
        return err
    }

    // 类型不可修改校验
    if existing.Type != server.Type {
        return ErrServerTypeCannotBeChanged
    }

    // ... 已有逻辑
}
```

- [ ] **Step 1.2: 在 handler 层添加错误处理**
在 `internal/api/http/handler/mcp/mcp_server_handler.go` 的 `UpdateServer` 函数中添加类型错误处理：

```go
if err := h.service.UpdateServer(ctx.Request.Context(), server); err != nil {
    if err == service.ErrMCPServerNotFound {
        ctx.JSON(http.StatusNotFound, gin.H{
            "error": "mcp server not found",
            "id":    idParam,
        })
        return
    }
    if err == service.ErrServerTypeCannotBeChanged {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": "server type cannot be changed after creation",
        })
        return
    }
    // ... 已有错误处理
}
```

- [ ] **Step 1.3: 运行编译确保通过**

```bash
go build ./...
```

Expected: 编译通过

---

### Task 2: 扩展 ToolStore 接口和实现

**Files:**
- Modify: `internal/domain/repository/tool_repository.go`
- Modify: `internal/infrastructure/database/gorm_tool_store.go`

- [ ] **Step 2.1: 更新 ToolStore 接口**
在 `internal/domain/repository/tool_repository.go` 添加：

```go
// ToolStore 定义工具定义存储接口
type ToolStore interface {
    // ... 已有方法
    GetByNameAndServer(ctx context.Context, name, vauthKey string) (*model.ToolDefinition, error)
    Save(ctx context.Context, tool *model.ToolDefinition) error
}
```

- [ ] **Step 2.2: 更新 GORM 实现**
在 `internal/infrastructure/database/gorm_tool_store.go` 添加：

```go
func (s *GORMToolStore) GetByNameAndServer(ctx context.Context, name, vauthKey string) (*model.ToolDefinition, error) {
    var tool model.ToolDefinition
    err := s.db.WithContext(ctx).
        Where("name = ? AND vauth_key = ? AND enabled = ?", name, vauthKey, true).
        First(&tool).Error
    if err != nil {
        return nil, err
    }
    return &tool, nil
}

func (s *GORMToolStore) Save(ctx context.Context, tool *model.ToolDefinition) error {
    return s.db.WithContext(ctx).Save(tool).Error
}
```

注意：需要先查看现有 `gorm_tool_store.go` 的实现后再正确修改。

- [ ] **Step 2.3: 运行编译确保通过**

```bash
go build ./...
```

Expected: 编译通过

---

### Task 3: 创建 ToolDomainService 领域服务

**Files:**
- Create: `internal/domain/service/tool_domain_service.go`
- Modify: `internal/service/tool_service.go`

- [ ] **Step 3.1: 创建领域服务**
创建 `internal/domain/service/tool_domain_service.go`：

```go
package service

import (
    "context"
    "errors"

    "dynamic_mcp_go_server/internal/domain/model"
    "dynamic_mcp_go_server/internal/domain/repository"
)

var (
    ErrOnlyHTTPServiceServerCanHaveTools = errors.New("only http_service server can have tools")
    ErrToolNameAlreadyExists              = errors.New("tool with same name already exists in this server")
    ErrHTTPServiceNotFound                = errors.New("http service not found")
)

type CreateToolFromHTTPServiceCommand struct {
    Name        string
    Description string
    ServerID    uint
    ServiceID   uint
    InputExtra  []byte
    OutputMapping []byte
}

// ToolDomainService 工具领域服务
type ToolDomainService struct {
    toolStore    repository.ToolStore
    serverStore  repository.MCPServerStore
    serviceStore repository.ServiceStore
}

func NewToolDomainService(
    toolStore repository.ToolStore,
    serverStore repository.MCPServerStore,
    serviceStore repository.ServiceStore,
) *ToolDomainService {
    return &ToolDomainService{
        toolStore:    toolStore,
        serverStore:  serverStore,
        serviceStore: serviceStore,
    }
}

// CreateToolFromHTTPService 从 HTTP Service 创建工具
func (s *ToolDomainService) CreateToolFromHTTPService(ctx context.Context, cmd CreateToolFromHTTPServiceCommand) (*model.ToolDefinition, error) {
    // 1. 校验 MCPServer 存在且类型为 http_service
    server, err := s.serverStore.GetByID(ctx, cmd.ServerID)
    if err != nil {
        return nil, errors.New("mcp server not found")
    }
    if server.Type != "http_service" {
        return nil, ErrOnlyHTTPServiceServerCanHaveTools
    }

    // 2. 校验 HTTPService 存在
    _, err = s.serviceStore.Get(ctx, cmd.ServiceID)
    if err != nil {
        return nil, ErrHTTPServiceNotFound
    }

    // 3. 校验工具名称不重复
    existing, _ := s.toolStore.GetByNameAndServer(ctx, cmd.Name, server.VAuthKey)
    if existing != nil {
        return nil, ErrToolNameAlreadyExists
    }

    // 4. 创建工具
    tool := &model.ToolDefinition{
        Name:          cmd.Name,
        Description:   cmd.Description,
        VAuthKey:      server.VAuthKey,
        ServiceID:     cmd.ServiceID,
        InputExtra:    cmd.InputExtra,
        OutputMapping: cmd.OutputMapping,
        Enabled:       true,
    }

    if err := s.toolStore.Save(ctx, tool); err != nil {
        return nil, err
    }

    return tool, nil
}
```

- [ ] **Step 3.2: 创建 ToolService 应用服务**
如果没有 `tool_service.go` 则创建：

```go
package service

import (
    "context"

    "dynamic_mcp_go_server/internal/domain/model"
    domainService "dynamic_mcp_go_server/internal/domain/service"
)

// ToolService 工具应用服务
type ToolService struct {
    toolDomainService *domainService.ToolDomainService
}

func NewToolService(
    toolDomainService *domainService.ToolDomainService,
) *ToolService {
    return &ToolService{
        toolDomainService: toolDomainService,
    }
}

// CreateFromHTTPService 从 HTTPService 创建工具
func (s *ToolService) CreateFromHTTPService(ctx context.Context, cmd domainService.CreateToolFromHTTPServiceCommand) (*model.ToolDefinition, error) {
    return s.toolDomainService.CreateToolFromHTTPService(ctx, cmd)
}
```

- [ ] **Step 3.3: 运行编译确保通过**

```bash
go build ./...
```

Expected: 编译通过

---

### Task 4: 添加从 HTTP Service 创建工具的 Handler

**Files:**
- Modify: `internal/api/http/handler/mcp/mcp_server_handler.go`
- Modify: `internal/api/http/router.go`

- [ ] **Step 4.1: 更新 MCPServerHandler 结构**
在 `mcp_server_handler.go` 中添加 `toolService` 字段：

```go
type MCPServerHandler struct {
    service     *service.MCPServerService
    toolService *service.ToolService
    db          *gorm.DB
    logger      logger.Logger
}

func NewMCPServerHandler(
    svc *service.MCPServerService,
    toolSvc *service.ToolService,
    db *gorm.DB,
    log logger.Logger,
) *MCPServerHandler {
    return &MCPServerHandler{
        service:     svc,
        toolService: toolSvc,
        db:          db,
        logger:      log,
    }
}
```

- [ ] **Step 4.2: 添加创建工具请求和 Handler**
添加新的请求结构体和 Handler：

```go
// CreateToolFromHTTPServiceRequest 从 HTTP Service 创建工具请求
type CreateToolFromHTTPServiceRequest struct {
    Name           string          `json:"name" binding:"required"`
    Description    string          `json:"description"`
    ServiceID      uint            `json:"service_id" binding:"required"`
    InputExtra     json.RawMessage `json:"input_extra"`
    OutputMapping  json.RawMessage `json:"output_mapping"`
}

// CreateToolFromHTTPService POST /api/admin/mcp-servers/:id/tools/from-http-service
func (h *MCPServerHandler) CreateToolFromHTTPService(ctx *gin.Context) {
    idParam := ctx.Param("id")
    id, err := strconv.ParseUint(idParam, 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error": "invalid server id",
            "id":    idParam,
        })
        return
    }

    var req CreateToolFromHTTPServiceRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "error":   "invalid request body",
            "details": err.Error(),
        })
        return
    }

    tool, err := h.toolService.CreateFromHTTPService(ctx.Request.Context(), domainService.CreateToolFromHTTPServiceCommand{
        Name:          req.Name,
        Description:   req.Description,
        ServerID:      uint(id),
        ServiceID:     req.ServiceID,
        InputExtra:    req.InputExtra,
        OutputMapping: req.OutputMapping,
    })
    if err != nil {
        switch {
        case err.Error() == "mcp server not found":
            ctx.JSON(http.StatusNotFound, gin.H{
                "error": "mcp server not found",
            })
        case errors.Is(err, domainService.ErrOnlyHTTPServiceServerCanHaveTools):
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": err.Error(),
            })
        case errors.Is(err, domainService.ErrToolNameAlreadyExists):
            ctx.JSON(http.StatusConflict, gin.H{
                "error": err.Error(),
            })
        case errors.Is(err, domainService.ErrHTTPServiceNotFound):
            ctx.JSON(http.StatusNotFound, gin.H{
                "error": err.Error(),
            })
        default:
            ctx.JSON(http.StatusInternalServerError, gin.H{
                "error":   "failed to create tool",
                "details": err.Error(),
            })
        }
        return
    }

    ctx.JSON(http.StatusCreated, gin.H{
        "message": "tool created successfully",
        "tool":    tool,
    })
}
```

需要添加 `encoding/json` 和 `errors` 导入，以及 `domainService` 包导入。

- [ ] **Step 4.3: 更新 router**
在 `router.go` 中正确初始化依赖并注册路由：

```go
// 初始化 domain service
toolDomainService := domainService.NewToolDomainService(
    &database.GORMToolStore{DB: gormDB},
    mcpServerStore,
    &database.GORMServiceStore{DB: gormDB},
)

// 初始化 application service
toolService := service.NewToolService(toolDomainService)

// 创建 handler
mcpHandler := mcp.NewMCPServerHandler(mcpServerService, toolService, gormDB, log)

// 注册路由
adminGroup.POST("/mcp-servers/:id/tools/from-http-service", mcpHandler.CreateToolFromHTTPService)
```

需要添加正确的导入路径。

- [ ] **Step 4.4: 运行编译确保通过**

```bash
go build ./...
```

Expected: 编译通过

---

### Task 5: 优化 MCPServerHandler 中直接查库的代码（DDD 整改）

**Files:**
- Modify: `internal/api/http/handler/mcp/mcp_server_handler.go`
- Modify: `internal/service/mcp_server_service.go`

- [ ] **Step 5.1: 在 MCPServerService 添加缺失的方法**
将 handler 中直接查库的逻辑移到 service 层：
- GetServerTools
- AddToolsToServer
- RemoveToolFromServer

修改后 handler 中不再直接使用 `h.db` 调用 GORM。

- [ ] **Step 5.2: 更新 handler 使用 service 层**

- [ ] **Step 5.3: 运行编译确保通过**

```bash
go build ./...
```

Expected: 编译通过

---

### Task 6: 运行完整编译和测试

**Files:**

- [ ] **Step 6.1: 运行完整编译**

```bash
go build ./...
```

Expected: 编译通过

---

## 验收

完成后验证：

1. ✅ 创建 MCP Server 后，修改 type 时返回错误
2. ✅ 可以从 HTTP Service 创建 Tool 关联到 http_service 类型的 MCP Server
3. ✅ 同一 MCP Server 下工具名称不能重复
4. ✅ 非 http_service 类型的 Server 不能关联工具
5. ✅ Handler 层不直接使用 GORM，都通过 Service 层
