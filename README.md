# Dynamic MCP-Go Server

Go 实现的动态 MCP (Model Context Protocol) 服务器，支持将多个 HTTP 接口打包成一个 MCP Server。

## 快速开始

### 使用 Makefile 管理（推荐）

```bash
# 开发模式 - 代码热更新，修改代码后立即生效
make dev

# 生产模式 - 需要先构建镜像
make build
make up

# 其他命令
make logs          # 查看日志
make down          # 停止服务
make restart       # 重启服务
make ps            # 查看状态
make clean         # 清理
```

### Docker Compose 手动操作

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

## 架构

```
┌─────────────────────────────────────────────────────┐
│                    MCP Client                        │
└─────────────────────────┬───────────────────────────┘
                          │ HTTP/MCP Protocol
                          ▼
┌─────────────────────────────────────────────────────┐
│              MCP Server (:18080)                      │
│  ┌─────────────────────────────────────────────┐    │
│  │  Gin HTTP Engine + Token Auth Middleware     │    │
│  └─────────────────────────────────────────────┘    │
│                        │                             │
│  ┌─────────────────────▼────────────────────────┐   │
│  │    Dynamic Registry (MySQL Store)            │   │
│  └─────────────────────────────────────────────┘    │
└─────────────────────────┬───────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────┐
│                  MySQL (:3306)                       │
│  mcp_tool_definitions / mcp_auth_keys / mcp_http_svcs │
└─────────────────────────────────────────────────────┘
```

## 主要功能

- **双轨认证**: MCP 端点使用 Token 验签，后台 API 使用 JWT
- **分组隔离**: MCP 服务按 vauth_key 分组隔离
- **动态工具定义**: 从 MySQL 实时加载 MCP 工具
- **HTTP 服务集成**: 可配置外部 HTTP 服务
- **管理后台**: Vue3 前端管理界面

## 目录结构

```
.
├── cmd/server/           # 服务入口
├── internal/
│   ├── auth/            # 认证管理
│   ├── config/          # 配置管理
│   ├── httpservice/     # HTTP 服务
│   ├── runtime/         # 运行时
│   └── tooldef/         # 工具定义
├── web/admin/           # Vue3 前端
├── docs/
│   ├── architecture.md  # 架构文档
│   └── mysql_migration.sql  # 数据库初始化
├── docker-compose.yml   # 生产配置
├── docker-compose.dev.yml # 开发配置
├── Dockerfile           # 生产镜像
├── Dockerfile.dev       # 开发镜像
├── .air.toml           # 热更新配置
└── Makefile            # 管理脚本
```

## API 接口

### MCP 端点（分组隔离）

MCP 服务按 `vauth_key` 进行分组隔离，每个分组有独立的 MCP Server：

| 路由 | 说明 |
|------|------|
| `/mcp/{vauth_key}` | MCP 协议端点（按 vauth_key 分组，需 X-Mcp-Token 验签） |
| `/mcp/{vauth_key}/{tool_name}` | 单个工具元数据查询（GET） |
| `/mcp` | 路由引导，列出所有分组 |

> **分组隔离**: 每个 `vauth_key` 对应一个独立的 MCP Server 实例，实现服务分组隔离

### 后台管理 API（统一 /api 前缀）

| 路由 | 说明 |
|------|------|
| `/api/v1/services` | HTTP 服务管理（增删改查） |
| `/api/v1/execute/:id` | 服务执行 |
| `/api/v1/auth/tokens` | Token 管理（增删改、启用/禁用、刷新） |
| `/api/v1/users` | 用户管理（仅管理员） |
| `/auth/login` | 用户登录 |
| `/auth/me` | 获取当前用户 |
| `/health` | 健康检查 |

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `MYSQL_DSN` | root:1234qwer@tcp(mysql:3306)/mcp_server?... | MySQL 连接 |
| `ADMIN_TOKEN` | admin-secret-token | 管理员 Token |
| `TOOL_STORE` | mysql | 存储类型 |

## 前端管理后台

```bash
cd web/admin
npm install
npm run dev
```

访问 http://localhost:3000

## 初始化数据库

```bash
# 方式1: 使用 make
make db-init

# 方式2: Docker exec
docker-compose exec -T mysql mysql -uroot -p1234qwer mcp_server < docs/mysql_migration.sql
```
