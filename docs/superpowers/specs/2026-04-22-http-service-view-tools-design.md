# HTTP 服务查看关联工具 - 设计文档

**日期**: 2026-04-22

## 需求概述

在 HTTP 服务页面，每个服务卡片上新增一个"查看工具"按钮。点击后弹出居中 Modal，显示该服务关联的工具列表（name、id）。点击列表中的工具可跳转到工具编辑页面。

## 交互流程

1. 在 **ServicesPage** 的每个服务卡片的操作栏中，新增 **"查看工具"** 按钮
2. 点击按钮 → 弹出居中 Modal，显示该 HTTP 服务关联的工具列表
3. 列表项显示：工具名称
4. 点击列表项 → 跳转到 `/tools?editId={toolId}` → 自动打开该工具的编辑对话框
5. Modal 不因跳转而关闭，用户可手动关闭

## UI 设计

### 按钮位置
```
┌─────────────────────────────────────────┐
│  [服务图标]  服务名称  [启用]           │
│  描述文字...                             │
│  [GET] [JSON] https://api.example.com   │
│  [请求转换] [响应转换] [入参Schema]...   │
├─────────────────────────────────────────┤
│  [编辑] [删除]          [调试] [查看工具]│
└─────────────────────────────────────────┘
```

### Modal 布局（居中展示）
```
┌───────────────────────────────────────┐
│  关联工具                        [×]  │
├───────────────────────────────────────┤
│  服务名称: xxx 服务                     │
│                                       │
│  ┌─────────────────────────────────┐  │
│  │ 工具名称1（可点击跳转）         │  │
│  │ 工具名称2（可点击跳转）         │  │
│  │ ...                              │  │
│  │ (空时显示：暂无关联工具)         │  │
│  └─────────────────────────────────┘  │
│                                       │
│              [关闭]                   │
└───────────────────────────────────────┘
```

## 后端接口

### 新增接口

**GET /api/admin/http-services/:id/tools**

获取指定 HTTP 服务关联的工具列表。

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
    "service_name": "用户服务",
    "tools": [
      { "id": 1, "name": "get_user" },
      { "id": 2, "name": "create_user" }
    ],
    "count": 2
  }
}
```

**响应字段说明**:
| 字段 | 类型 | 说明 |
|------|------|------|
| service_id | uint | HTTP 服务 ID |
| service_name | string | HTTP 服务名称 |
| tools | array | 工具列表 |
| tools[].id | uint | 工具 ID |
| tools[].name | string | 工具名称 |
| count | int | 工具数量 |

## 前端修改

### ServicesPage.vue

1. 在每个服务卡片的操作栏新增"查看工具"按钮
2. 新增 ToolListModal 组件（居中展示的 Modal）
3. 点击按钮时调用 API 获取工具列表
4. 点击工具项时使用 router.push 跳转到 /tools?editId=xxx

### ToolsPage.vue

1. 在 onMounted 或 watch 中检测 URL query 参数 editId
2. 如果存在 editId，自动打开该工具的编辑对话框

## 涉及文件

| 文件 | 修改内容 |
|------|----------|
| `internal/api/http/handler/mcp/http_service_handler.go` | 新增 GetServiceTools handler |
| `internal/service/http_service_service.go` | 新增 GetServiceTools 方法 |
| `internal/domain/repository/http_service_repository.go` | 新增 GetToolsByServiceID 方法 |
| `internal/infrastructure/database/gorm_http_service.go` | 实现 GetToolsByServiceID |
| `web/admin/src/pages/ServicesPage.vue` | 新增按钮和弹框 |
| `web/admin/src/components/ServiceToolsModal.vue` | 新增工具列表弹框组件 |
| `web/admin/src/pages/ToolsPage.vue` | 支持 editId 参数自动打开编辑 |
| `docs/api/http-services.md` | 更新接口文档 |

## 实现步骤

1. 后端：新增 GetServiceTools handler
2. 后端：新增 service 方法
3. 后端：新增 repository 方法
4. 前端：新增 ServiceToolsModal.vue 组件
5. 前端：ServicesPage.vue 集成按钮和弹框
6. 前端：ToolsPage.vue 支持 editId 参数
7. 文档：更新 http-services.md