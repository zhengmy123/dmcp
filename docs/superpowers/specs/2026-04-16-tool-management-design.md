# MCP Server 工具管理功能设计 (DDD 风格)

## 背景

1. 新增工具定义管理页面，将 HTTP Service 转换为 MCP 工具
2. 支持多个 Tool 关联同一个 http-service 类型的 MCP Server
3. MCP Server 类型创建后不可修改
4. VAuthKey 复制按钮（列表页、编辑弹窗、创建成功提示）

## DDD 分层架构

```
┌─────────────────────────────────────────────────────────┐
│  Interface Layer (Handler/API)                          │
│  - HTTP Request/Response 处理                          │
│  - 参数绑定、校验                                        │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Application Layer (Service)                           │
│  - MCPServerService                                     │
│  - ToolService                                          │
│  - 业务用例编排                                          │
│  - 跨 aggregate 业务规则校验                             │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Domain Layer (Model + Domain Service)                  │
│  - MCPServer (Aggregate Root)                           │
│  - ToolDefinition (Entity)                             │
│  - HTTPService (Entity, 不同 Bounded Context)           │
│  - 领域业务规则                                          │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Infrastructure Layer (Repository + Impl)                │
│  - GORMMCPServerDAO (Repository 实现)                   │
│  - GORMToolStore (Repository 实现)                       │
│  - GORMServiceStore (Repository 实现)                   │
└─────────────────────────────────────────────────────────┘
```

## 领域模型

### MCPServer (Aggregate Root)

```go
type MCPServer struct {
    ID             uint
    VAuthKey       string
    Name           string
    Description    string
    Type           ServerType  // proxy | http_service
    Enabled        bool
    // ... 其他字段
}

type ServerType string

const (
    ServerTypeProxy       ServerType = "proxy"
    ServerTypeHTTPService ServerType = "http_service"
)

// 领域方法
func (s *MCPServer) CanAddTool() bool
func (s *MCPServer) ValidateTypeChange(newType ServerType) error
```

### ToolDefinition (Entity)

```go
type ToolDefinition struct {
    ID          uint
    Name        string
    Description string
    Parameters  []ParameterDefinition
    VAuthKey    string  // 关联 MCPServer
    ServiceID   uint    // 关联 HTTPService
    Enabled     bool
    // ... 其他字段
}
```

## 领域服务

### MCPServerDomainService

```go
type MCPServerDomainService struct {
    serverStore repository.MCPServerStore
    toolStore  repository.ToolStore
}

// ValidateTypeChange 类型不可修改校验
func (s *MCPServerDomainService) ValidateTypeChange(ctx context.Context, serverID uint, newType ServerType) error {
    server, err := s.serverStore.GetByID(ctx, serverID)
    if err != nil {
        return ErrMCPServerNotFound
    }
    if server.Type != newType {
        return ErrServerTypeCannotBeChanged
    }
    return nil
}

// CanAssociateTool 校验是否可以关联工具
func (s *MCPServerDomainService) CanAssociateTool(ctx context.Context, serverID uint) error {
    server, err := s.serverStore.GetByID(ctx, serverID)
    if err != nil {
        return ErrMCPServerNotFound
    }
    if server.Type != ServerTypeHTTPService {
        return ErrOnlyHTTPServiceServerCanHaveTools
    }
    return nil
}
```

### ToolDomainService

```go
type ToolDomainService struct {
    toolStore    repository.ToolStore
    serverStore  repository.MCPServerStore
    httpServiceStore repository.ServiceStore
}

// CreateToolFromHTTPService 从 HTTPService 创建工具
func (s *ToolDomainService) CreateToolFromHTTPService(ctx context.Context, cmd CreateToolCommand) (*ToolDefinition, error) {
    // 1. 校验 MCPServer 存在且类型为 http_service
    server, err := s.serverStore.GetByID(ctx, cmd.ServerID)
    if err != nil {
        return nil, ErrMCPServerNotFound
    }
    if server.Type != ServerTypeHTTPService {
        return nil, ErrOnlyHTTPServiceServerCanHaveTools
    }

    // 2. 校验 HTTPService 存在
    _, err = s.httpServiceStore.Get(ctx, cmd.ServiceID)
    if err != nil {
        return nil, ErrHTTPServiceNotFound
    }

    // 3. 校验工具名称不重复
    existing, _ := s.toolStore.GetByNameAndServer(ctx, cmd.Name, server.VAuthKey)
    if existing != nil {
        return nil, ErrToolNameAlreadyExists
    }

    // 4. 创建工具
    tool := &ToolDefinition{
        Name:        cmd.Name,
        Description: cmd.Description,
        VAuthKey:    server.VAuthKey,
        ServiceID:   cmd.ServiceID,
        Enabled:     true,
        // ...
    }

    if err := s.toolStore.Save(ctx, tool); err != nil {
        return nil, err
    }

    return tool, nil
}
```

## 应用服务

### MCPServerService

```go
// UpdateServer 更新 MCPServer（添加类型校验）
func (s *MCPServerService) UpdateServer(ctx context.Context, server *MCPServer) error {
    existing, err := s.serverStore.GetByID(ctx, server.ID)
    if err != nil {
        return ErrMCPServerNotFound
    }

    // 核心业务规则：类型不可修改
    if existing.Type != server.Type {
        return ErrServerTypeCannotBeChanged
    }

    return s.serverStore.Save(ctx, server)
}
```

### ToolService

```go
type ToolService struct {
    toolDomainService *ToolDomainService
}

// CreateFromHTTPService 从 HTTPService 创建工具
func (s *ToolService) CreateFromHTTPService(ctx context.Context, cmd CreateToolFromHTTPServiceCommand) (*ToolDefinition, error) {
    return s.toolDomainService.CreateToolFromHTTPService(ctx, cmd)
}
```

## Handler 层（接口层）

### MCPServerHandler

```go
// UpdateServer PUT /api/admin/mcp-servers/:id
func (h *MCPServerHandler) UpdateServer(ctx *gin.Context) {
    // ...

    // 使用 service 层处理业务逻辑
    if err := h.service.UpdateServer(ctx.Request.Context(), server); err != nil {
        if err == service.ErrServerTypeCannotBeChanged {
            ctx.JSON(http.StatusBadRequest, gin.H{
                "error": "server type cannot be changed after creation",
            })
            return
        }
        // ...
    }
}
```

### ToolHandler

```go
// AddToolsToServerRequest 添加工具请求（新增）
type AddToolsToServerRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
    ServerID    uint   `json:"server_id" binding:"required"`
    ServiceID   uint   `json:"service_id" binding:"required"`
    InputSchema json.RawMessage `json:"input_schema"`
}

// AddToolsToServer POST /api/admin/mcp-servers/:id/tools
func (h *ToolHandler) AddToolsToServer(ctx *gin.Context) {
    // ...

    tool, err := h.toolService.CreateFromHTTPService(ctx.Request.Context(), CreateToolFromHTTPServiceCommand{
        Name:        req.Name,
        Description: req.Description,
        ServerID:    req.ServerID,
        ServiceID:   req.ServiceID,
    })
    if err != nil {
        // 统一错误处理
    }
}
```

## Repository 接口

```go
// ToolStore
type ToolStore interface {
    Save(ctx context.Context, tool *ToolDefinition) error
    GetByID(ctx context.Context, id uint) (*ToolDefinition, error)
    GetByNameAndServer(ctx context.Context, name, vauthKey string) (*ToolDefinition, error)
    ListByServerID(ctx context.Context, serverID uint) ([]*ToolDefinition, error)
    Delete(ctx context.Context, id uint) error
}

// MCPServerStore
type MCPServerStore interface {
    GetByID(ctx context.Context, id uint) (*MCPServer, error)
    GetByVAuthKey(ctx context.Context, vauthKey string) (*MCPServer, error)
    Save(ctx context.Context, server *MCPServer) error
    Delete(ctx context.Context, id uint) error
}
```

## 错误定义

```go
var (
    ErrServerTypeCannotBeChanged   = errors.New("server type cannot be changed after creation")
    ErrOnlyHTTPServiceServerCanHaveTools = errors.New("only http_service server can have tools")
    ErrToolNameAlreadyExists       = errors.New("tool with same name already exists in this server")
    ErrHTTPServiceNotFound         = errors.New("http service not found")
)
```

## 前端改动

### 创建工具弹窗

```
┌─────────────────────────────────────────────┐
│  创建 MCP 工具                            [X] │
├─────────────────────────────────────────────┤
│                                             │
│  HTTP Service: [下拉选择 ▼]                 │
│    └─ 显示 name, target_url                 │
│                                             │
│  工具名称: [________________]                │
│                                             │
│  描述: [________________]                    │
│                                             │
│  目标 Server: [下拉选择 http_service 类型 ▼]│
│    └─ 只显示 type=http_service 的 Server    │
│                                             │
│           [取消]  [创建]                     │
└─────────────────────────────────────────────┘
```

### 复制按钮实现

```javascript
// hooks/useCopyToClipboard.js
export const useCopyToClipboard = () => {
  const copy = async (text) => {
    await navigator.clipboard.writeText(text);
    // 可选：使用 antd message.success('复制成功')
  };
  return { copy };
};
```

## 验收标准

| 功能 | 验收条件 |
|------|----------|
| HTTP Service 转工具 | 从 HTTP Service 列表选择后可创建 Tool |
| 多 Tool 关联一 Server | 同一 http_service Server 可创建多个 Tool |
| 类型不可修改 | 后端返回错误，前端 disabled |
| VAuthKey 复制 | 列表/编辑/创建成功提示都有复制按钮 |
| DDD 分层 | Handler → Service → DomainService → Repository |
