# 工具绑定管理

**需要认证**: 是（JWT）

## 接口列表

### GET /api/admin/tool-bindings/:toolId

获取工具的绑定列表。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| toolId | uint | 是 | 工具 ID |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "bindings": [ ... ],
    "count": 10
  }
}
```

### POST /api/admin/tool-bindings

绑定工具到服务器。

**请求体**:
```json
{
  "tool_id": 1,
  "server_id": 1
}
```

### DELETE /api/admin/tool-bindings/:toolId/:serverId

解绑工具和服务器。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| toolId | uint | 是 | 工具 ID |
| serverId | uint | 是 | 服务器 ID |

### POST /api/admin/tool-bindings/batch-bind

批量绑定工具到服务器。

**请求体**:
```json
{
  "tool_ids": [1, 2, 3],
  "server_ids": [1, 2]
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "count": 6
  }
}
```

### DELETE /api/admin/tool-bindings/batch-unbind

批量解绑工具和服务器。

**请求体**:
```json
{
  "binding_ids": [1, 2, 3]
}
```

### GET /api/admin/server-bindings/:serverId

获取服务器的绑定列表。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| serverId | uint | 是 | 服务器 ID |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "bindings": [ ... ],
    "count": 10
  }
}
```
