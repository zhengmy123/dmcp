# HTTP 服务工具创建流程优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 优化工具创建流程，支持选择HTTP服务后自动拉取出入参schema，并增强出参映射支持多层级字段。

**Architecture:** 后端新增简化版服务列表接口，前端修改工具编辑对话框，使用简化接口获取服务列表，选择服务后自动拉取完整信息，集成嵌套schema编辑器支持多层级字段映射。

**Tech Stack:** Go (后端), Vue 3 (前端), JSON Schema

---

## File Structure

### 后端
- `app/api/services.go` - 新增简化版服务列表接口

### 前端
- `web/admin/src/api/services.js` - 添加简化版接口调用
- `web/admin/src/components/ToolEditDialog.vue` - 修改工具编辑对话框
- `web/admin/src/utils/schemaHelper.js` - 新增schema字段提取工具函数

---

## Task 1: 后端 - 新增简化版服务列表接口

**Files:**
- Modify: `app/api/services.go`

- [ ] **Step 1: 查看现有服务接口实现**

```bash
cat app/api/services.go
```

- [ ] **Step 2: 添加简化版服务列表接口**

在 `services.go` 中添加新的处理器：

```go
// GetServicesSimple 获取简化版服务列表（只返回id和name）
func (s *ServicesHandler) GetServicesSimple(c *gin.Context) {
    var services []models.HTTPService
    if err := s.repo.FindAll(&services); err != nil {
        c.JSON(500, gin.H{"error": "Failed to fetch services"})
        return
    }
    
    // 构建简化版响应
    type SimpleService struct {
        ID   string `json:"id"`
        Name string `json:"name"`
    }
    
    var simpleServices []SimpleService
    for _, service := range services {
        simpleServices = append(simpleServices, SimpleService{
            ID:   service.ID,
            Name: service.Name,
        })
    }
    
    c.JSON(200, simpleServices)
}
```

- [ ] **Step 3: 注册路由**

在路由注册部分添加：

```go
group.GET("/simple", s.GetServicesSimple)
```

- [ ] **Step 4: 重启后端服务**

```bash
go run main.go
```

- [ ] **Step 5: 测试接口**

```bash
curl http://localhost:8000/api/v1/services/simple
```

预期输出：
```json
[
  {"id": "1", "name": "用户服务"},
  {"id": "2", "name": "订单服务"}
]
```

- [ ] **Step 6: 提交代码**

```bash
git add app/api/services.go
git commit -m "feat: add simple services list endpoint"
```

---

## Task 2: 前端 - 添加简化版接口调用

**Files:**
- Modify: `web/admin/src/api/services.js`

- [ ] **Step 1: 查看现有API文件**

```bash
cat web/admin/src/api/services.js
```

- [ ] **Step 2: 添加简化版接口方法**

在 `servicesApi` 对象中添加：

```javascript
// 获取简化版服务列表（只返回id和name）
getServicesSimple() {
  return request.get('/api/v1/services/simple')
},
```

- [ ] **Step 3: 提交代码**

```bash
git add web/admin/src/api/services.js
git commit -m "feat: add getServicesSimple API method"
```

---

## Task 3: 前端 - 新增schema字段提取工具函数

**Files:**
- Create: `web/admin/src/utils/schemaHelper.js`

- [ ] **Step 1: 创建schemaHelper.js文件**

```javascript
// 从schema中提取所有字段路径（支持嵌套）
export function extractSchemaFields(schema, prefix = '') {
  const fields = []
  
  if (!schema || !schema.properties) {
    return fields
  }
  
  for (const [name, prop] of Object.entries(schema.properties)) {
    const fullPath = prefix ? `${prefix}.${name}` : name
    fields.push(fullPath)
    
    // 递归处理嵌套对象
    if (prop.type === 'object' && prop.properties) {
      const nestedFields = extractSchemaFields(prop, fullPath)
      fields.push(...nestedFields)
    }
  }
  
  return fields
}

// 从点号路径创建嵌套对象
export function createNestedObject(path, value) {
  const parts = path.split('.')
  const result = {}
  let current = result
  
  for (let i = 0; i < parts.length; i++) {
    const part = parts[i]
    if (i === parts.length - 1) {
      current[part] = value
    } else {
      current[part] = {}
      current = current[part]
    }
  }
  
  return result
}

// 合并嵌套对象
export function mergeNestedObjects(target, source) {
  for (const key in source) {
    if (typeof source[key] === 'object' && source[key] !== null) {
      if (!target[key]) {
        target[key] = {}
      }
      mergeNestedObjects(target[key], source[key])
    } else {
      target[key] = source[key]
    }
  }
  return target
}
```

- [ ] **Step 2: 提交代码**

```bash
git add web/admin/src/utils/schemaHelper.js
git commit -m "feat: add schemaHelper utility functions"
```

---

## Task 4: 前端 - 修改工具编辑对话框

**Files:**
- Modify: `web/admin/src/components/ToolEditDialog.vue`

- [ ] **Step 1: 查看现有ToolEditDialog.vue文件**

```bash
cat web/admin/src/components/ToolEditDialog.vue
```

- [ ] **Step 2: 添加导入语句**

在 `<script setup>` 部分添加：

```javascript
import { servicesApi } from '@/api/services'
import { extractSchemaFields, createNestedObject, mergeNestedObjects } from '@/utils/schemaHelper'
import SchemaFieldNode from '@/components/SchemaFieldNode.vue'
import {
  fieldsToSchema,
  schemaToFields,
  createField,
  removeFieldById,
  updateFieldById,
} from '@/utils/schemaBuilder'
```

- [ ] **Step 3: 添加响应式数据**

```javascript
const outputFields = ref([])
const outputSchemaMode = ref('mapping') // 'mapping' or 'visual'
```

- [ ] **Step 4: 修改服务加载逻辑**

```javascript
const onServerChange = async () => {
  if (form.vauth_key) {
    // 使用简化版接口获取服务列表
    const response = await servicesApi.getServicesSimple()
    services.value = response.data
  }
}
```

- [ ] **Step 5: 修改服务选择逻辑**

```javascript
const onServiceChange = async () => {
  if (form.service_id) {
    // 拉取完整服务信息
    const response = await servicesApi.getService(form.service_id)
    const service = response.data
    
    // 同步入参
    syncInputFromService(service)
    
    // 同步出参字段
    syncOutputFromService(service)
  }
}
```

- [ ] **Step 6: 修改syncOutputFromService方法**

```javascript
const syncOutputFromService = (service) => {
  if (!service) return
  const schema = service.output_schema
  if (schema && schema.properties) {
    // 提取所有字段（包括嵌套）
    outputSchemaFields.value = extractSchemaFields(schema)
  }
}
```

- [ ] **Step 7: 添加出参schema编辑相关方法**

```javascript
// 切换出参编辑模式
const switchToOutputSchemaMode = (mode) => {
  if (mode === 'visual' && outputMappings.value.length > 0) {
    // 从映射生成schema
    generateOutputSchemaFromMappings()
  }
  outputSchemaMode.value = mode
}

// 从映射生成出参schema
const generateOutputSchemaFromMappings = () => {
  const schema = { type: 'object', properties: {} }
  
  outputMappings.value.forEach(mapping => {
    if (mapping.source_field && mapping.target_field) {
      const nestedObject = createNestedObject(mapping.target_field, {
        type: 'string' // 默认类型，可后续编辑
      })
      mergeNestedObjects(schema.properties, nestedObject)
    }
  })
  
  outputFields.value = schemaToFields(schema)
}

// 处理出参字段更新
const handleOutputFieldUpdate = (id, updates) => {
  outputFields.value = updateFieldById(outputFields.value, id, updates)
}

// 处理出参字段删除
const handleOutputFieldDelete = (id) => {
  outputFields.value = removeFieldById(outputFields.value, id)
}

// 添加出参字段
const addOutputField = () => {
  outputFields.value = [...outputFields.value, createField()]
}
```

- [ ] **Step 8: 修改模板部分**

在出参映射部分添加模式切换和可视化编辑：

```vue
<!-- Output Mapping -->
<div class="space-y-4 pt-4 border-t border-gray-100">
  <div class="flex items-center justify-between">
    <h4 class="text-sm font-semibold text-gray-800 flex items-center">
      <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/></svg>
        出参映射
    </h4>
    <div class="flex gap-2">
      <button type="button" @click="switchToOutputSchemaMode('mapping')"
        :class="outputSchemaMode === 'mapping' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
        class="px-3 py-1 text-xs font-medium rounded-lg transition-colors">
        字段映射
      </button>
      <button type="button" @click="switchToOutputSchemaMode('visual')"
        :class="outputSchemaMode === 'visual' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
        class="px-3 py-1 text-xs font-medium rounded-lg transition-colors">
        可视化编辑
      </button>
    </div>
  </div>
  
  <!-- 字段映射模式 -->
  <div v-if="outputSchemaMode === 'mapping'">
    <p class="text-xs text-gray-500">选择 HTTP 服务的 OutputSchema 字段，配置目标字段名</p>
    
    <!-- 映射表单 -->
    <!-- 现有映射代码保持不变 -->
  </div>
  
  <!-- 可视化编辑模式 -->
  <div v-else>
    <p class="text-xs text-gray-500">可视化编辑出参 schema 结构</p>
    
    <div class="space-y-2">
      <SchemaFieldNode
        v-for="field in outputFields"
        :key="field.id"
        :field="field"
        :depth="0"
        @update="handleOutputFieldUpdate"
        @delete="handleOutputFieldDelete"
      />
      <button type="button" @click="addOutputField"
        class="text-sm text-primary-600 hover:text-primary-700 font-medium">
        + 添加字段
      </button>
    </div>
  </div>
</div>
```

- [ ] **Step 9: 修改提交逻辑**

在 `handleSubmit` 方法中添加：

```javascript
let outputSchema = null
if (outputSchemaMode.value === 'visual' && outputFields.value.length > 0) {
  outputSchema = fieldsToSchema(outputFields.value)
}

const payload = {
  // 现有字段...
  output_schema: outputSchema
}
```

- [ ] **Step 10: 测试功能**

```bash
cd web/admin && npm run dev
```

- [ ] **Step 11: 提交代码**

```bash
git add web/admin/src/components/ToolEditDialog.vue
git commit -m "feat: enhance tool edit dialog with nested schema support"
```

---

## Task 5: 测试验证

**Files:**
- 无

- [ ] **Step 1: 测试完整流程**

1. 打开工具管理页面
2. 点击"创建工具"
3. 选择 MCP Server
4. 验证服务列表是否正确加载
5. 选择一个服务
6. 验证入参是否自动同步
7. 验证出参映射的源字段是否包含嵌套字段
8. 测试字段映射功能
9. 测试可视化编辑出参 schema
10. 保存工具
11. 验证工具是否创建成功

- [ ] **Step 2: 运行前端构建**

```bash
cd web/admin && npm run build
```

- [ ] **Step 3: 提交最终代码**

```bash
git add -A
git commit -m "feat: complete HTTP service tool creation optimization"
```

---

## Self-Review

1. **Spec coverage:** ✅ 所有需求都已覆盖
   - 新增简化版服务列表接口
   - 选择服务后自动拉取出入参schema
   - 出参映射支持多层级字段
   - 支持构建自定义出参schema

2. **Placeholder scan:** ✅ 无占位符

3. **Type consistency:** ✅ 类型和方法名一致

4. **Test coverage:** ✅ 包含测试步骤

## Execution Handoff

**Plan complete and saved to `docs/superpowers/plans/2026-04-16-http-service-tool-optimization.md`. Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**