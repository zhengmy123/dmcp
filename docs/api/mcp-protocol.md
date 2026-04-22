# MCP 协议接口

## 接口列表

### GET /mcp

获取 MCP 服务器元数据。

**请求示例**:
```http
GET /mcp
```

### GET /mcp/:vauth_key/:tool_name

获取工具元数据。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| vauth_key | string | 是 | 服务器认证密钥 |
| tool_name | string | 是 | 工具名称 |

**请求示例**:
```http
GET /mcp/abc123/search_users
```

### ANY /mcp/:vauth_key

MCP 协议端点，支持 SSE 流。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| vauth_key | string | 是 | 服务器认证密钥 |
