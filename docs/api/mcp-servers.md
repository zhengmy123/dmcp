# MCP Server 管理

**需要认证**: 是（JWT）

## 接口列表

### GET /api/admin/mcp-servers

获取 MCP Server 列表（支持分页和搜索）。

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 10，最大 100 |
| name | string | 否 | 名称搜索（模糊匹配） |
| state | int | 否 | 状态筛选：1-正常，0-已删除 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "servers": [
      {
        "id": 1,
        "name": "server1",
        "type": "http_service",
        "vauth_key": "abc123",
        "description": "",
        "http_server_url": "",
        "auth_header": "",
        "timeout_seconds": 30,
        "extra_headers": "",
        "state": 1,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "tool_count": 5
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "total_page": 10
  }
}
```

**响应字段说明**:
| 字段 | 类型 | 说明 |
|------|------|------|
| servers | array | 服务器列表 |
| servers[].tool_count | int | 绑定的工具数量 |
| servers[].state | int | 状态：1-正常，0-已删除 |
| total | int | 总记录数 |
| page | int | 当前页码 |
| page_size | int | 每页数量 |
| total_page | int | 总页数 |

### GET /api/admin/mcp-servers/:id

获取单个 MCP Server 详情。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务器 ID |

### POST /api/admin/mcp-servers

创建 MCP Server。

**请求体**:
```json
{
  "type": "string",
  "name": "string",
  "description": "string",
  "http_server_url": "string",
  "auth_header": "string",
  "timeout_seconds": 30,
  "extra_headers": "string"
}
```

**说明**: type 可以是 "local" 或 "http_service"

### PUT /api/admin/mcp-servers/:id

更新 MCP Server。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务器 ID |

**请求体**: 同创建服务器

**注意**: 服务器类型（type）创建后不能修改。

### DELETE /api/admin/mcp-servers/:id

删除 MCP Server（软删除）。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务器 ID |

### POST /api/admin/mcp-servers/:id/restore

恢复已删除的 MCP Server。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务器 ID |

### GET /api/admin/mcp-servers/:id/tools

获取服务器绑定的工具列表。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务器 ID |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "server_id": 1,
    "vauth_key": "string",
    "tools": [ ... ],
    "count": 10
  }
}
```

### POST /api/admin/mcp-servers/:id/tools

向服务器添加工具。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务器 ID |

**请求体**:
```json
{
  "tools": [ ... ]
}
```

### DELETE /api/admin/mcp-servers/:id/tools/:toolName

从服务器移除工具。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务器 ID |
| toolName | string | 是 | 工具名称 |

### POST /api/admin/mcp-servers/:id/tools/from-http-service

从 HTTP 服务创建工具并绑定到服务器。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务器 ID |

**请求体**:
```json
{
  "name": "string",
  "description": "string",
  "service_id": 1,
  "input_extra": [],
  "output_mapping": []
}
```

### POST /api/admin/mcp-servers/:id/sync-build

同步服务器构建信息。当工具或 HTTP 服务有变更时，调用此接口可以生成最新的构建版本。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务器 ID |

**业务逻辑**:
- 收集当前服务器绑定的所有工具及其关联的 HTTP 服务信息
- 计算新的 hash 值
- 如果 hash 与当前 active build 相同，则不生成新版本
- 如果 hash 不同，则创建新的构建版本（新的 build_uuid）

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "server_id": 1
  }
}
```

**错误响应**:
- 404: 服务器不存在
- 500: 服务器内部错误
