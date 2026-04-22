# 系统配置

系统配置用于存储全局配置项，采用 Key-Value 模式。

## 获取配置

### GET /api/v1/system/config/:key

获取指定配置项的值。

**需要认证**: 是（JWT）

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| key | string | 配置键名（如 `api_host`） |

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "config_key": "api_host",
    "config_value": "http://localhost:18080"
  }
}
```

**错误响应**:
- 404: 配置不存在

## 更新配置

### PUT /api/v1/system/config/:key

更新指定配置项的值。

**需要认证**: 是（JWT）

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| key | string | 配置键名（如 `api_host`） |

**请求体**:
```json
{
  "config_value": "http://localhost:18080"
}
```

**响应示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "config_key": "api_host",
    "config_value": "http://localhost:18080"
  }
}
```

## 预定义配置项

| 配置键 | 说明 | 默认值 |
|--------|------|--------|
| api_host | MCP Server API 访问地址 | http://localhost:18080 |
