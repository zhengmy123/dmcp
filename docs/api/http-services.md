# HTTP 服务管理

**需要认证**: 是（JWT）

## 接口列表

### GET /api/v1/services

获取服务列表。

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "services": [ ... ],
    "count": 10
  }
}
```

### GET /api/v1/services/simple

获取简化版服务列表（仅包含 id 和 name）。

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "services": [
      {
        "id": 1,
        "name": "Service Name"
      }
    ],
    "count": 10
  }
}
```

### GET /api/v1/services/:id

获取单个服务详情。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务 ID |

### POST /api/v1/services

创建新服务。

**请求体**:
```json
{
  "name": "string",
  "method": "GET|POST|PUT|DELETE",
  "target_url": "string",
  "headers": {},
  "input_schema": {},
  "output_schema": {},
  "body_type": "json|form|raw"
}
```

### PUT /api/v1/services/:id

更新服务信息。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务 ID |

**请求体**: 同创建服务

### DELETE /api/v1/services/:id

删除服务。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务 ID |

### POST /api/v1/execute/:id

执行服务。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务 ID |

**请求体**:
```json
{
  "headers": {},
  "body": {},
  "query": {}
}
```

### POST /api/v1/services/:id/debug

调试服务。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务 ID |

**请求体**:
```json
{
  "headers": {},
  "body": {},
  "query": {},
  "body_type": "string"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "response": {
      "success": true,
      "status_code": 200,
      "request_headers": {},
      "response_headers": {},
      "request_body": {},
      "response_body": {},
      "duration_ms": 100,
      "error": "",
      "input_schema": {},
      "output_schema": {}
    }
  }
}
```

### POST /webhook/:id

Webhook 处理器。

**需要认证**: 是（JWT）

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 服务 ID |
