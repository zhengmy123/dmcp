# 系统配置功能设计

## 1. 概述

为前端提供全局唯一的 API Host 配置存储功能，采用 Key-Value 模式，支持后续扩展其他系统配置。

## 2. 数据模型

### 2.1 数据库表

```sql
CREATE TABLE `system_configs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `config_key` varchar(64) NOT NULL COMMENT '配置键',
  `config_value` text COMMENT '配置值',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';
```

### 2.2 领域模型

```go
type SystemConfig struct {
    ID          uint64    `gorm:"primaryKey;autoIncrement"`
    ConfigKey   string    `gorm:"column:config_key;type:varchar(64);not null;uniqueIndex"`
    ConfigValue string    `gorm:"column:config_value;type:text"`
    CreatedAt   time.Time `gorm:"autoCreateTime"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (SystemConfig) TableName() string {
    return "system_configs"
}
```

## 3. API 接口

### 3.1 获取配置

- **路径**: `GET /api/v1/system/config/:key`
- **响应**:

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

- **错误**: 配置不存在时返回 404

### 3.2 更新配置

- **路径**: `PUT /api/v1/system/config/:key`
- **请求体**:

```json
{
  "config_value": "http://localhost:18080"
}
```

- **响应**:

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

## 4. 目录结构

按 DDD 模式组织：

```
internal/
├── domain/
│   └── model/
│       └── system_config.go      # 领域模型
├── domain/repository/
│   └── system_config.go          # 仓储接口
├── infrastructure/
│   └── store/
│       └── systemconfig/
│           └── store.go          # GORM 存储实现
├── service/
│   └── system_config.go           # 应用服务
├── api/http/
│   ├── handler/
│   │   └── system_config.go       # HTTP 处理器
│   └── router.go                  # 路由注册
```

## 5. 前端改动

### 5.1 API 层

新增 `/api/system/config/:key` 接口调用

### 5.2 SettingsPage.vue

- 页面加载时从后端 API 获取 `api_host` 配置
- 保存时调用 PUT 接口更新

## 6. 实现步骤

1. 创建数据库表 `system_configs`
2. 创建领域模型 `SystemConfig`
3. 创建仓储接口和 GORM 实现
4. 创建应用服务
5. 创建 HTTP Handler 和路由
6. 前端 API 和页面对接

## 7. 初始数据

系统初始化时插入默认配置：

| config_key | config_value |
|------------|--------------|
| api_host | http://localhost:18080 |
