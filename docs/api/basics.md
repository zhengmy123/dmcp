# 基础接口

## 健康检查

### GET /health

健康检查接口，用于验证服务是否正常运行。

**请求示例**:
```http
GET /health
```

**响应示例**:
```json
{
  "status": "ok"
}
```
