# 工具创建/编辑页面优化设计文档

**日期**: 2026-04-17  
**版本**: 1.0  
**作者**: AI Assistant

## 一、概述

本文档描述了动态 MCP Go Server 项目中工具创建/编辑页面的优化方案，主要包括入参定义区域和出参映射区域的功能调整。

## 二、目标

### 2.1 入参定义优化
- 选择 HTTP 服务后自动同步入参定义，移除"从服务同步"按钮
- 支持修改字段名，修改后同步更新入参映射
- 完善必填状态控制逻辑
- 重新设计删除字段恢复区域

### 2.2 出参映射优化
- 选择 HTTP 服务后自动拉取出参 schema，移除"从服务同步"按钮
- 移除完整映射模式，统一使用自定义字段模式
- 实现 1:1 自动映射
- 每个映射项一行显示，包含路径和类型
- 新增 FieldSelector 组件用于弹框选择源字段

## 三、详细设计

### 3.1 入参定义区域

#### 3.1.1 自动同步逻辑
- 选择 HTTP 服务时，调用 `onServiceChange()` 方法
- 该方法自动调用 `syncInputFromService()` 同步入参
- 移除入参定义区域的"从服务同步"按钮

#### 3.1.2 字段名修改逻辑
**数据结构说明:**
```go
// InputMappingField 入参映射字段
type InputMappingField struct {
	Source      string // 源字段（来自MCP工具入参）- 用户修改的字段名
	Target      string // 目标字段名（HTTP服务InputSchema）- HTTP原始字段名
	Description string
}
```

**修改流程:**
1. 用户修改 `param.name`（UI显示和MCP工具入参名）
2. 监听字段名变化，同步更新入参映射中对应项的 `source` 字段
3. `target` 字段保持 HTTP 原始字段名不变（用于调用时传给 HTTP 服务）

**实现方式:**
- 使用 `watch` 监听 `inputParams` 的变化
- 当检测到字段名变化时，查找对应的入参映射项并更新 `source`

#### 3.1.3 必填状态控制
- **HTTP schema 必填字段**（`schema_required: true`）：
  - 显示"必填"标签（禁用状态）
  - 不显示删除按钮
- **HTTP schema 非必填字段**（`schema_required: false`）：
  - 可以切换"可选"/"必填"
  - 显示删除按钮

#### 3.1.4 删除/恢复区域重新设计
- 采用卡片式布局展示已删除字段
- 显示字段名、类型、描述
- 支持单个恢复和批量恢复
- 位于入参定义区域下方

### 3.2 出参映射区域

#### 3.2.1 自动同步逻辑
- 选择 HTTP 服务时，调用 `onServiceChange()` 方法
- 该方法自动调用 `syncOutputFromService()` 同步出参
- 移除出参映射区域的"从服务同步"按钮
- 移除完整映射模式，统一使用自定义字段模式

#### 3.2.2 新增 FieldSelector 组件
**组件职责:**
- 提供弹框式字段选择器
- 显示 schema 树形结构
- 支持选择源字段
- 显示选中字段的路径和类型

**Props:**
- `modelValue`: 当前选中的字段路径
- `nodes`: schema 树形节点列表

**Events:**
- `update:modelValue`: 选中字段变化时触发

#### 3.2.3 schemaHelper.js 新增函数
```javascript
// 根据路径获取节点的类型
export function getFieldTypeByPath(schema, path)

// 根据路径从节点列表中获取节点
export function getNodeByPath(nodes, path)
```

#### 3.2.4 出参映射 UI 重构
**布局结构:**
1. 快速选择区域：显示所有可用字段，点击即添加映射
2. 映射列表区域：每个映射项一行显示
   - 目标字段输入框
   - 映射箭头（←）
   - FieldSelector 组件（选择源字段）
   - 类型标签
   - 删除按钮

**1:1 自动映射:**
- 点击快速选择区域的字段时，自动添加映射
- 目标字段名自动填充为源字段的最后一段（如 `user.name` → `name`）

### 3.3 受影响的文件

| 文件路径 | 操作类型 | 说明 |
|---------|---------|------|
| `web/admin/src/components/ToolEditDialog.vue` | 修改 | 主组件，包含入参和出参区域 |
| `web/admin/src/components/FieldSelector.vue` | 新增 | 字段选择弹框组件 |
| `web/admin/src/utils/schemaHelper.js` | 修改 | 添加辅助函数 |

## 四、数据流程

```
用户选择 HTTP 服务
    ↓
onServiceChange() 被调用
    ↓
┌─────────────┬───────────────┐
↓             ↓               ↓
syncInputFromService()  syncOutputFromService()
    ↓             ↓               ↓
更新入参列表    更新出参字段列表
更新入参映射    更新出参树形结构
    ↓             ↓
用户修改字段名    用户点击快速选择字段
    ↓             ↓
watch 检测变化    addOutputMappingWithField()
    ↓             ↓
更新入参映射 source  添加 1:1 映射
```

## 五、错误处理

- HTTP 服务 schema 解析失败时，显示空状态并记录错误日志
- 字段名修改时，确保映射项存在才进行更新
- 删除字段时，检查是否为 schema 必填字段，是则不允许删除

## 六、兼容性

- 保持现有的数据结构不变
- 向后兼容已有的工具定义
- 不影响后端 API 接口

## 七、验收标准

1. 选择 HTTP 服务后，入参和出参自动同步，无需手动点击同步按钮
2. 修改入参字段名后，入参映射的 source 字段同步更新
3. HTTP schema 必填字段不能修改必填状态和删除
4. 删除的字段可以在恢复区域重新添加
5. 出参映射快速选择区域点击字段后自动添加 1:1 映射
6. 出参映射每个项一行显示，包含目标字段、源字段选择器和类型标签
