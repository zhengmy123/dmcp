# 用户认证与授权

## 用户认证

### POST /auth/login

用户登录。

**请求体**:
```json
{
  "username": "string",
  "password": "string"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "jwt_token_string",
    "expires_at": 1234567890,
    "user": {
      "id": 1,
      "username": "admin",
      "name": "Admin",
      "email": "admin@example.com",
      "role": "admin",
    "state": 1,
      "last_login_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

### GET /auth/me

获取当前登录用户信息。

**需要认证**: 是（JWT）

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user": {
      "id": 1,
      "username": "admin",
      "name": "Admin",
      "email": "admin@example.com",
      "role": "admin",
      "state": 1
    }
  }
}
```

### POST /auth/change-password

修改当前用户密码。

**需要认证**: 是（JWT）

**请求体**:
```json
{
  "old_password": "string",
  "new_password": "string"
}
```

## 用户管理（管理员）

**需要认证**: 是（JWT + Admin 角色）

### GET /api/v1/users

获取用户列表。

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "users": [ ... ],
    "count": 10
  }
}
```

### POST /api/v1/users

创建新用户。

**请求体**:
```json
{
  "username": "string",
  "password": "string",
  "name": "string",
  "email": "string",
  "role": "admin|user"
}
```

### PUT /api/v1/users/:id

更新用户信息。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 用户 ID |

**请求体**:
```json
{
  "name": "string",
  "email": "string",
  "role": "admin|user",
  "state": 1
}
```

### DELETE /api/v1/users/:id

删除用户。

**路径参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | uint | 是 | 用户 ID |
