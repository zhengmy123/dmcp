# MCPServer 类型扩展设计方案

## 背景

将 MCPServer 扩展为两种类型：
1. **Proxy Server** - 直接代理到外部 MCP Server
2. **HTTP Service Server** - 本地配置 HTTPService，通过 tools 调用外部 API

## 数据模型

### MCPServer 新增字段

```go
type MCPServer struct {
    ID          uint      `json:"id"`
    VAuthKey    string    `json:"vauth_key"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Type        string    `json:"type"`        // "proxy" | "http_service"
    Enabled     bool      `json:"enabled"`

    // Proxy 类型专用
    HTTPServerURL   string            `json:"http_server_url"`   // 外部 MCP Server 地址
    AuthHeader      string            `json:"auth_header"`       // 认证头，如 "Bearer xxx"
    TimeoutSeconds  int               `json:"timeout_seconds"`   // 超时时间，默认 30
    ExtraHeaders    map[string]string `json:"extra_headers"`     // 额外请求头
}
```

### 数据库迁移

```sql
ALTER TABLE `mcp_servers` ADD COLUMN `type` VARCHAR(32) NOT NULL DEFAULT 'http_service' COMMENT '类型: proxy, http_service';
ALTER TABLE `mcp_servers` ADD COLUMN `http_server_url` VARCHAR(512) COMMENT '代理 Server 的目标 URL';
ALTER TABLE `mcp_servers` ADD COLUMN `auth_header` VARCHAR(256) COMMENT '认证头';
ALTER TABLE `mcp_servers` ADD COLUMN `timeout_seconds` INT NOT NULL DEFAULT 30 COMMENT '超时秒数';
ALTER TABLE `mcp_servers` ADD COLUMN `extra_headers` TEXT COMMENT '额外请求头 JSON';
```

## 架构设计

### 请求路由逻辑

```
请求 → /mcp/:vauth_key
    │
    ├─ 查 MCPServer
    │
    ├─ type == "proxy"
    │     └─ 透传请求到 http_server_url
    │         - 添加 AuthHeader
    │         - 添加 ExtraHeaders
    │         - 超时控制
    │
    └─ type == "http_service"
          └─ 现有逻辑（tools → HTTPService）
```

### Proxy 代理核心逻辑

```go
func (h *MCPServerHandler) ProxyToMCP(ctx context.Context, server *MCPServer, req *http.Request) (*http.Response, error) {
    // 1. 构建目标 URL
    targetURL := server.HTTPServerURL
    if !strings.HasSuffix(targetURL, "/mcp") {
        targetURL = strings.TrimSuffix(targetURL, "/") + "/mcp"
    }

    // 2. 创建代理请求
    proxyReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, req.Body)
    if err != nil {
        return nil, err
    }

    // 3. 复制请求头
    for k, vs := range req.Header {
        for _, v := range vs {
            proxyReq.Header.Add(k, v)
        }
    }

    // 4. 添加认证头
    if server.AuthHeader != "" {
        proxyReq.Header.Set("Authorization", server.AuthHeader)
    }

    // 5. 添加额外请求头
    for k, v := range server.ExtraHeaders {
        proxyReq.Header.Set(k, v)
    }

    // 6. 发送请求
    client := &http.Client{Timeout: time.Duration(server.TimeoutSeconds) * time.Second}
    return client.Do(proxyReq)
}
```

### 错误处理

| 场景 | 返回 |
|------|------|
| Server 不存在 | 404 Not Found |
| Server 已禁用 | 403 Forbidden |
| 代理请求超时 | 504 Gateway Timeout |
| 代理请求失败 | 502 Bad Gateway + 错误信息 |
| 代理返回非 200 | 502 + 响应体 |

## 前端交互设计

### 创建/编辑弹窗 - 类型选择联动

```
┌─────────────────────────────────────────────┐
│  创建 MCP Server                        [X] │
├─────────────────────────────────────────────┤
│                                             │
│  类型选择                                    │
│  ○ 代理 Server    ○ HTTP Service Server    │
│                                             │
│  ───────────────────────────────            │
│                                             │
│  名称: [________________]                    │
│                                             │
│  描述: [________________]                    │
│                                             │
│  ┌─ 代理 Server 选中时显示 ─┐              │
│  │ HTTP URL: [https://...]  │              │
│  │ 认证头:  [Bearer xxx]    │              │
│  │ 超时:    [30] 秒         │              │
│  │ 请求头:  [Key: Value]    │              │
│  └─────────────────────────┘              │
│                                             │
│  ┌─ HTTP Service 选中时显示 ─┐            │
│  │ (无额外字段，tools 在下一步管理) │      │
│  └─────────────────────────┘              │
│                                             │
│           [取消]  [创建]                    │
└─────────────────────────────────────────────┘
```

### 列表页 - 类型筛选与展示

| 名称 | 类型 | 工具数 | 状态 | 操作 |
|------|------|--------|------|------|
| Server A | 代理 | - | 启用 | 编辑 删除 |
| Server B | HTTP Service | 5 | 启用 | 编辑 管理工具 删除 |

- 下拉筛选：全部 / 代理 / HTTP Service
- 代理 Server 不显示"工具数"和"管理工具"按钮
- HTTP Service Server 显示工具数，支持"管理工具"跳转

## 实现计划

### 后端改动
1. Model: `mcp_server.go` 添加新字段
2. 数据库迁移: `mysql_migration.sql` 添加新列
3. Handler: `mcp_server_handler.go` 添加 Proxy 转发逻辑
4. Router: `router.go` 统一入口判断类型
5. Service: `mcp_server_service.go` 可选

### 前端改动
1. API: `mcpServers.js` 更新请求/响应结构
2. Store: `tools.js` 更新数据结构
3. Page: `MCPServersPage.vue` 重构为类型选择器 + 联动表单
4. 列表页增加类型筛选

## 约束

- HTTP Service Server 类型的 tools 只能选择 HTTPService 类型
- Proxy Server 类型不需要管理 tools
