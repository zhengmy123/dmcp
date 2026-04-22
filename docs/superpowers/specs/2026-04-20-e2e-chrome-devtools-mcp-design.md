# E2E 测试使用 Chrome-DevTools-MCP 设计文档

## 概述

本文档描述了如何使用 `chrome-devtools-mcp` 工具通过 MCP 协议执行端到端测试，验证系统的工具绑定、MCP Server 调用以及三级缓存功能。

## 技术选型

| 组件 | 技术 |
|------|------|
| 测试框架 | Playwright + @modelcontextprotocol/sdk |
| 测试语言 | Node.js (ES Modules) |
| MCP 客户端 | @modelcontextprotocol/sdk |
| JSON 处理 | 内置 JSON |

## 项目结构

```
test/e2e/
├── playwright.config.js        # Playwright 配置（已有）
├── package.json                # 依赖配置（已有）
├── tests/
│   ├── mcp-flow.spec.js        # MCP 协议流程测试（已有）
│   └── chrome-mcp.spec.js      # Chrome DevTools MCP E2E 测试（新增）
└── mcp/
    ├── client.js               # MCP 客户端封装
    └── types.js                # MCP 类型定义
```

## 测试场景

### 场景 1: 工具绑定 MCP Server 验证

**验证点：**
- 管理员登录获取 JWT Token
- 创建 HTTP Service（服务定义）
- 创建 Tool 并绑定到 MCP Server
- 通过 MCP 协议查询工具列表，验证绑定是否生效

**预期结果：**
- `GET /mcp/:vauth_key` 返回已绑定工具列表
- 工具元数据（name, description, parameters）正确

### 场景 2: MCP Server 调用验证

**验证点：**
- 通过 MCP 协议的 `tools/call` 或 `tools/execute` 调用工具
- 验证工具执行结果

**预期结果：**
- 工具调用成功
- 返回正确的执行结果

### 场景 3: 三级缓存验证

**验证点：**
- 首次调用：L1 miss → L2 miss → L3 hit → 回填 L2 & L1
- 二次调用：L1 hit（验证缓存命中）
- 缓存失效：删除 Tool 后再次调用，验证缓存更新

**验证方式：**
- 通过响应时间对比验证缓存效果
- 通过日志确认缓存命中情况

## MCP 协议接口

### MCP Server 端点

| 方法 | 说明 |
|------|------|
| `GET /mcp/:vauth_key` | 获取 MCP Server 元数据 |
| `GET /mcp/:vauth_key/:tool_name` | 获取工具元数据 |
| `ANY /mcp/:vauth_key` | MCP 协议主端点（支持 SSE 流） |

### MCP JSON-RPC 2.0 消息格式

**请求：**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/list",
  "params": {}
}
```

**响应：**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "tools": [...]
  }
}
```

### 核心方法

| 方法 | 说明 |
|------|------|
| `initialize` | 初始化 MCP 会话 |
| `tools/list` | 获取可用工具列表 |
| `tools/call` | 调用指定工具 |

## 实现步骤

1. 安装 `@modelcontextprotocol/sdk` 依赖
2. 创建 MCP 客户端封装 (`mcp/client.js`)
3. 创建 Chrome DevTools MCP 测试 (`tests/chrome-mcp.spec.js`)
4. 添加缓存验证辅助函数
5. 运行测试验证

## 验证指标

| 指标 | 预期值 |
|------|--------|
| 工具绑定 | 绑定后立即可通过 MCP 查询到 |
| 首次调用延迟 | < 500ms（含 L1/L2/L3 全部 miss） |
| 缓存命中延迟 | < 50ms |
