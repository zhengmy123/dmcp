# MCP Server Admin Frontend

Vue3 前端管理后台，用于管理 MCP Server、HTTP 服务、工具定义和 Token 认证。

## 技术栈

- Vue 3 + Composition API
- Vite
- Pinia（状态管理）
- Vue Router
- Tailwind CSS
- TypeScript

## 项目结构

```
src/
├── api/           # API 接口封装
├── components/    # 公共组件
├── layouts/       # 页面布局
├── pages/         # 页面组件
├── router/        # 路由配置
├── stores/        # Pinia 状态管理
├── styles/        # 全局样式
├── types/         # TypeScript 类型定义
└── utils/         # 工具函数
```

## 安装依赖

```bash
cd web/admin
npm install
```

## 开发模式

```bash
npm run dev
```

访问 http://localhost:3000

## 构建

```bash
npm run build
```

## 功能模块

| 模块 | 说明 |
|------|------|
| 仪表盘 | 统计概览、数据可视化 |
| MCP Servers | MCP Server 管理（增删改查） |
| HTTP 服务 | 管理外部 HTTP 服务配置 |
| 工具定义 | MCP 工具的查看和管理 |
| 工具绑定 | 工具与服务绑定关系管理 |
| Token 管理 | 创建、刷新、启用/禁用、删除认证 Token |
| 用户管理 | 用户登录、权限管理（仅管理员） |
| 系统设置 | API 连接参数配置 |

## 环境配置

前端开发环境连接后端服务：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `VITE_API_BASE_URL` | http://localhost:17050 | 后端 API 地址 |
