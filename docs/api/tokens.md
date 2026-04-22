# Token 管理

**需要认证**: 是（JWT）

## 接口列表

### GET /api/v1/auth/tokens

获取 Token 列表。

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [ ... ],
    "total": 10,
    "page": 1,
    "page_size": 10
  }
}
```

### POST /api/v1/auth/tokens

创建新 Token。

**请求体**:
```json
{
  "key_id": "string",
  "token": "string",
  "secret": "string",
  "name": "string"
}
```

**说明**: 如果不提供 key_id、token、secret，系统会自动生成。

### DELETE /api/v1/auth/tokens/:token

删除 Token。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | Token 值 |

### POST /api/v1/auth/tokens/:token/refresh

刷新 Token。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | 原 Token 值 |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "old_token": "string",
    "new_token": "string",
    "new_secret": "string"
  }
}
```

### POST /api/v1/auth/tokens/:token/enable

启用 Token。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | Token 值 |

### POST /api/v1/auth/tokens/:token/disable

禁用 Token。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | Token 值 |
