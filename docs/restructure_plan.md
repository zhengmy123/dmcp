# 项目目录结构（DDD + 现代化 Go 架构）

```
internal/
├── common/
│   ├── logger/            # 【日志抽象层】
│   │   ├── logger.go      # Logger 接口定义 + Field 便捷函数
│   │   ├── zap_logger.go  # zap 实现
│   │   └── factory.go     # NewFileLogger 工厂函数
│   └── middleware/         # 【全项目通用中间件】
│       ├── cors.go        # 跨域
│       ├── recovery.go    # 崩溃恢复（含链路ID记录）
│       ├── logger.go      # 请求日志（使用 Logger 接口）
│       ├── requestid.go   # 链路ID（UUID v4 高效生成）
│       └── trace.go       # 链路追踪（Go runtime/trace Task + Region）
│
├── api/
│   ├── http/              # HTTP API 层（使用 common 中间件）
│   │   ├── handler/       # 请求处理器
│   │   └── router.go      # 路由注册
│   └── middleware/         # API 层特有中间件
│       ├── auth.go        # JWT 认证 + Admin 权限
│       └── signature.go   # 签名验证 + MCP Token 认证
│
├── service/               # 应用服务层（业务编排）
│   ├── auth_service.go
│   ├── user_service.go
│   ├── http_service.go
│   ├── signature_service.go
│   ├── script_validator.go
│   ├── registry.go
│   ├── group_manager.go
│   ├── http_routes.go
│   └── scoped_handler.go
│
├── domain/                # 领域层
│   ├── model/             # 领域模型
│   └── repository/        # 仓储接口
│
├── infrastructure/        # 基础设施层
│   ├── auth/              # JWT 实现
│   ├── database/          # 数据库 DAO
│   └── store/             # 存储实现
│       ├── httpservice/
│       └── tooldef/
│
└── config/                # 配置
```

## 全局中间件链（main.go 中注册顺序）

```
Recovery → RequestID → Trace → Cors → Logger
```

## Logger 接口

所有模块通过 `logger.Logger` 接口记录日志，不直接依赖 `*zap.Logger`。
接口提供 `Debug/Info/Warn/Error/Fatal` + `With()` 方法。

## 链路追踪

- `RequestID()` 中间件生成 UUID v4 作为 request_id
- `Trace()` 中间件为每个请求创建 `runtime/trace.Task`
- Logger 通过 `With(request_id, trace_task)` 将链路信息注入日志
- 所有日志自动携带 request_id

## generateRequestID

采用 `github.com/google/uuid` 的 UUID v4（高效、全局唯一、无冲突）。
