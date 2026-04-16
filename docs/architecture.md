# Dynamic MCP-Go Server 架构设计

## 概述

Dynamic MCP-Go Server 是一个用 Go 实现的动态 MCP (Model Context Protocol) 服务器，支持将多个 HTTP 接口打包成一个 MCP Server，通过 vauthKey 聚合 tools。

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        MCP Client                                │
└─────────────────────────┬───────────────────────────────────────┘
                          │ HTTP/MCP Protocol
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                      MCP Server (:18080)                         │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   Gin HTTP Engine                        │    │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  │    │
│  │  │ MCP Routes │  │ Admin API   │  │ Service Routes  │  │    │
│  │  │ /mcp/{key} │  │ /api/v1/*   │  │ /webhook/*      │  │    │
│  │  └──────┬──────┘  └──────┬──────┘  └────────┬────────┘  │    │
│  │         │                │                  │            │    │
│  │  ┌──────▼────────────────▼──────────────────▼────────┐   │    │
│  │  │              Auth Middleware                      │   │    │
│  │  │  X-Mcp-Token (MCP) / JWT Bearer (Admin API)    │   │    │
│  │  └──────────────────┬───────────────────────────────┘   │    │
│  └─────────────────────┼───────────────────────────────────┘    │
│                        │                                        │
│  ┌─────────────────────▼───────────────────────────────────┐    │
│  │              Dynamic Registry                            │    │
│  │  - 定期从 Store 同步工具定义                              │    │
│  │  - 按 vauthKey 分组建 MCP Server                         │    │
│  └─────────────────────┬───────────────────────────────────┘    │
│                        │                                        │
│  ┌─────────────────────▼───────────────────────────────────┐    │
│  │              MCP Group Manager                            │    │
│  │  - 管理多个 MCP Server 实例                               │    │
│  │  - 每个 vauthKey 对应一个 MCP Server                     │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────┬───────────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────────┐
│                     Store Layer                                  │
│  ┌──────────────────┐  ┌──────────────────┐                     │
│  │   MySQL Store    │  │  Memory Store    │                     │
│  │ (mcp_tool_defs)  │  │   (fallback)     │                     │
│  └──────────────────┘  └──────────────────┘                     │
└─────────────────────────────────────────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────────┐
│                     MySQL (:3306)                                │
│  ┌─────────────────┐  ┌─────────────────┐  ┌────────────────┐  │
│  │mcp_tool_defs    │  │ mcp_auth_keys   │  │mcp_http_svcs   │  │
│  │                 │  │                 │  │                │  │
│  │ - 工具定义      │  │ - Token管理     │  │ - HTTP服务配置 │  │
│  └─────────────────┘  └─────────────────┘  └────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## 分组隔离架构

MCP 服务采用分组隔离架构，通过 `vauth_key` 将工具定义分组，每个分组对应一个独立的 MCP Server 实例。

### 分组隔离设计

```
┌─────────────────────────────────────────────────────────────────┐
│                    MCP Server (统一入口 :18080)                   │
├─────────────────────────────────────────────────────────────────┤
│  /mcp/user-service    ──→  MCP Server (user-service)           │
│  /mcp/order-service    ──→  MCP Server (order-service)          │
│  /mcp/product-service ──→  MCP Server (product-service)       │
├─────────────────────────────────────────────────────────────────┤
│  /api/v1/*           ──→  后台管理 API (统一前缀)                │
│  /auth/*             ──→  认证 API (公开)                        │
└─────────────────────────────────────────────────────────────────┘
```

### 分组隔离特点

- **独立 MCP Server**: 每个 `vauth_key` 分组有独立的 MCP Server 实例
- **隔离认证**: Token 认证按分组独立验证
- **独立工具集**: 每个分组只包含该分组下的工具定义
- **统一管理**: 后台管理 API 统一前缀 `/api/v1/`

## 双轨认证体系

系统采用两套独立的认证机制：

### MCP 端点 - Token 验签

| 端点 | 认证方式 | Header |
|------|---------|--------|
| `/mcp/{vauth_key}` | X-Mcp-Token 验签 | `X-Mcp-Token: {token}` |

- 使用 `mcp_auth_keys` 表中存储的 Token
- 支持 Token 刷新和启用/禁用
- 按 vauth_key 分组独立验证

### 后台管理 API - JWT 认证

| 端点 | 认证方式 | Header |
|------|---------|--------|
| `/api/v1/*` | JWT Bearer Token | `Authorization: Bearer {jwt_token}` |
| `/auth/login` | 公开登录 | 用户名密码 |

- 使用 `mcp_users` 表中存储的用户
- JWT Token 包含用户信息
- 支持用户管理、密码修改

### 1. cmd/server/main.go

应用入口，负责：
- 配置加载
- 存储层初始化
- 认证管理器初始化
- MCP 服务器创建
- HTTP 服务启动

### 2. internal/config/

配置管理模块：

| 文件 | 说明 |
|------|------|
| `config.go` | 主配置加载 |
| `http_config.go` | HTTP 服务相关配置 |

### 3. internal/auth/

认证管理模块 (`auth.go`):

| 功能 | 说明 |
|------|------|
| Token 认证 | MCP 端点验签（X-Token） |
| Admin 认证 | 后台管理 API 验签（X-Admin-Token） |
| Token 刷新 | 从 MySQL 定期加载最新 Token |
| CRUD 操作 | 创建/删除/启用/禁用 Token |

### 4. internal/tooldef/

工具定义存储模块：

| 文件 | 说明 |
|------|------|
| `types.go` | 工具定义数据结构 |
| `store.go` | Store 接口定义 |
| `mysql_store_enhanced.go` | MySQL 存储实现 |
| `memory_store.go` | 内存存储实现 |

### 5. internal/runtime/

运行时模块：

| 文件 | 说明 |
|------|------|
| `registry.go` | 动态注册表，定期同步工具定义 |
| `group_manager.go` | MCP 分组管理器 |
| `http_handler.go` | MCP HTTP 处理器 |
| `gin_routes.go` | Gin 路由注册 |

### 6. internal/httpservice/

HTTP 服务管理模块：

| 文件 | 说明 |
|------|------|
| `service_manager.go` | HTTP 服务管理器 |
| `controller.go` | HTTP 服务控制器 |
| `store.go` | 服务存储（Memory/MySQL） |
| `signature_manager.go` | 签名管理器 |

## 数据流

### MCP 请求流程

```
Client Request
     │
     ▼
┌─────────────┐
│ Gin Router  │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│ Token Auth MW   │  ← X-Token 验签
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│ MCP Handler     │  ← 处理 MCP 协议
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│ MCP Group Mgr   │  ← 路由到对应 vauthKey
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│ MCP Server      │  ← 执行工具调用
└─────────────────┘
```

### Token 刷新流程

```
Timer (5min)
    │
    ▼
┌─────────────────┐
│ Load From MySQL │  ← SELECT * FROM mcp_auth_keys
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│ Update Memory   │  ← 更新内存中的 Token 缓存
└─────────────────┘
```

## 数据库表结构

### mcp_tool_definitions

存储 MCP 工具定义：

```sql
CREATE TABLE mcp_tool_definitions (
    id          VARCHAR(36) PRIMARY KEY,
    vauth_key   VARCHAR(128) NOT NULL,      -- 服务分组键
    name        VARCHAR(128) NOT NULL,      -- 工具名称
    description TEXT,                       -- 工具描述
    parameters_json JSON,                   -- 参数定义
    enabled     TINYINT(1) DEFAULT 1,       -- 启用状态
    created_at  DATETIME,
    updated_at  DATETIME,
    UNIQUE KEY uk_vauth_name (vauth_key, name)
);
```

### mcp_auth_keys

存储访问令牌：

```sql
CREATE TABLE mcp_auth_keys (
    id          VARCHAR(36) PRIMARY KEY,
    key_id      VARCHAR(64) NOT NULL,       -- Key ID
    token       VARCHAR(128) NOT NULL,      -- 访问令牌
    secret      VARCHAR(256) NOT NULL,     -- 密钥
    name        VARCHAR(128),              -- 名称/描述
    enabled     TINYINT(1) DEFAULT 1,       -- 启用状态
    last_used_at DATETIME,                 -- 最后使用时间
    expires_at  DATETIME,                  -- 过期时间
    created_at  DATETIME,
    updated_at  DATETIME,
    UNIQUE KEY uk_key_id (key_id),
    UNIQUE KEY uk_token (token)
);
```

### mcp_http_services

存储 HTTP 服务配置：

```sql
CREATE TABLE mcp_http_services (
    id                  VARCHAR(36) PRIMARY KEY,
    name                VARCHAR(128) NOT NULL,
    description         TEXT,
    target_url          VARCHAR(512) NOT NULL,
    method              VARCHAR(16) DEFAULT 'POST',
    headers             JSON,
    timeout_seconds     INT DEFAULT 30,
    retry_count         INT DEFAULT 3,
    validation_script    TEXT,
    validation_enabled  TINYINT(1) DEFAULT 0,
    signature_enabled   TINYINT(1) DEFAULT 0,
    signature_algorithm VARCHAR(32) DEFAULT 'HMAC-SHA256',
    signature_key       VARCHAR(256),
    signature_header    VARCHAR(64) DEFAULT 'X-Signature',
    enabled             TINYINT(1) DEFAULT 1,
    created_at          DATETIME,
    updated_at          DATETIME
);
```

## API 接口

### MCP 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| ANY | `/mcp/{vauth_key}` | MCP 协议端点 |
| GET | `/mcp/{vauth_key}/{tool_name}` | 工具元数据 |
| GET | `/mcp` | 路由引导 |

### 后台管理 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/auth/tokens` | 列出 Token |
| POST | `/api/v1/auth/tokens` | 创建 Token |
| DELETE | `/api/v1/auth/tokens/{token}` | 删除 Token |
| POST | `/api/v1/auth/tokens/{token}/refresh` | 刷新 Token |
| POST | `/api/v1/auth/tokens/{token}/enable` | 启用 Token |
| POST | `/api/v1/auth/tokens/{token}/disable` | 禁用 Token |
| GET | `/api/v1/services` | 列出 HTTP 服务 |
| POST | `/api/v1/services` | 创建 HTTP 服务 |
| GET | `/api/v1/services/{id}` | 获取服务详情 |
| PUT | `/api/v1/services/{id}` | 更新服务 |
| DELETE | `/api/v1/services/{id}` | 删除服务 |
| POST | `/api/v1/execute/{id}` | 执行服务调用 |

### 其他

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| POST | `/webhook/{id}` | Webhook 端点 |

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `TOOL_STORE` | mysql | 存储类型 |
| `MYSQL_DSN` | root:1234qwer@tcp(127.0.0.1:3306)/mcp_server?... | MySQL 连接 |
| `ADMIN_TOKEN` | admin-secret-token | 管理员 Token |
| `REFRESH_SECONDS` | 10 | 工具刷新间隔 |
| `HTTP_SERVICE_STORE` | mysql | HTTP 服务存储 |
| `HTTP_MYSQL_DSN` | 同上 | HTTP 服务 MySQL |
| `HTTP_SYNC_INTERVAL` | 60 | HTTP 服务同步间隔 |

## 部署方式

### Docker Compose

```bash
docker-compose up -d
```

自动启动：
- MySQL 8.0 (端口 3306)
- MCP Server (端口 18080)

### 本地运行

```bash
# 初始化数据库
mysql -u root -p < docs/mysql_migration.sql

# 运行服务
go run ./cmd/server
```
