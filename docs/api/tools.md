# 工具管理

**需要认证**: 是（JWT）

## 接口列表

### GET /api/admin/tools

获取工具列表（支持分页和搜索）。

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| page_size | int | 否 | 每页数量，默认 20，最大 100 |
| keyword | string | 否 | 搜索关键词（工具名称） |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "tools": [ ... ]
  }
}
```

### GET /api/admin/tools/:id

获取单个工具详情。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 工具 ID |

### POST /api/admin/tools

创建新工具。

**请求体**:
```json
{
  "name": "string",
  "description": "string",
  "service_id": 1,
  "parameters": [],
  "input_mapping": [],
  "output_mapping": []
}
```

**注意**: 工具名称必须唯一，且只能包含小写字母、数字和下划线。

### PUT /api/admin/tools/:id

更新工具。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 工具 ID |

**请求体**:
```json
{
  "name": "string",
  "description": "string",
  "service_id": 1,
  "parameters": [],
  "input_mapping": [],
  "enabled": true,
  "output_mapping": []
}
```

### DELETE /api/admin/tools/:id

删除工具（软删除，仅禁用）。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 工具 ID |

### GET /api/admin/http-services/:id/output-schema

获取 HTTP 服务的输出 Schema。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | HTTP 服务 ID |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "service_id": 1,
    "name": "string",
    "output_schema": {}
  }
}
```
