# HTTP 服务查看关联工具 - 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 HTTP 服务页面添加"查看工具"按钮，点击弹出 Modal 显示该服务关联的工具列表，点击工具可跳转到工具编辑页。

**Architecture:** 后端在 `tool_handler.go` 新增 `GetHTTPServiceTools` 方法（路由已挂载到 `/api/admin/http-services/:id/tools`），前端新增 `ServiceToolsModal.vue` 组件，修改 `ServicesPage.vue` 和 `ToolsPage.vue`。

**Tech Stack:** Go (Gin, GORM, Sonic), Vue 3, Tailwind CSS

---

## 目录

- [Task 1: 后端 Handler - 新增 GetHTTPServiceTools 方法](#task-1-后端-handler---新增-gethttpservicetools-方法)
- [Task 2: 前端 API - 新增 getServiceTools 接口](#task-2-前端-api---新增-getservicetools-接口)
- [Task 3: 前端组件 - 新增 ServiceToolsModal.vue](#task-3-前端组件---新增-servicetoolsmodalvue)
- [Task 4: 前端页面 - ServicesPage.vue 添加按钮和弹框](#task-4-前端页面---servicespagevue-添加按钮和弹框)
- [Task 5: 前端页面 - ToolsPage.vue 支持 editId 参数](#task-5-前端页面---toolspagevue-支持-editid-参数)
- [Task 6: 文档 - 更新 http-services.md](#task-6-文档---更新-http-servicesmd)

---

## Task 1: 后端 Handler - 新增 GetHTTPServiceTools 方法

**Files:**
- Modify: `internal/api/http/handler/mcp/tool_handler.go`

**Files Context:**
- `tool_handler.go` 已通过 `mcpGroup.GET("/http-services/:id/output-schema", toolHandler.GetHTTPServiceOutputSchema)` 挂载
- `ToolQuery` 已有 `ServiceID *uint` 字段，支持按 service_id 过滤

- [ ] **Step 1: 在 tool_handler.go 新增 GetHTTPServiceTools 方法**

在文件末尾（第 388 行之后），添加：

```go
func (h *ToolHandler) GetHTTPServiceTools(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "invalid service id")
		return
	}

	service, err := h.serviceStore.Get(ctx.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(ctx, "service not found")
		return
	}

	serviceID := uint(id)
	query := &repository.ToolQuery{
		ServiceID: &serviceID,
		State:     func() *int { v := 1; return &v }(),
	}

	tools, _, err := h.toolStore.List(ctx.Request.Context(), query, 1, 100)
	if err != nil {
		response.InternalError(ctx, err.Error())
		return
	}

	items := make([]ToolItemResponse, 0, len(tools))
	for _, t := range tools {
		items = append(items, ToolItemResponse{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
		})
	}

	response.Success(ctx, gin.H{
		"service_id":    service.ID,
		"service_name":  service.Name,
		"tools":         items,
		"count":         len(items),
	})
}
```

- [ ] **Step 2: 在 router.go 添加路由**

在 `tool_handler.go` 已有的路由注册部分（大约第 94 行）：
在 `mcpGroup.GET("/http-services/:id/output-schema", toolHandler.GetHTTPServiceOutputSchema)` 后添加：

```go
mcpGroup.GET("/http-services/:id/tools", toolHandler.GetHTTPServiceTools)
```

---

## Task 2: 前端 API - 新增 getServiceTools 接口

**Files:**
- Modify: `web/admin/src/api/tools.js` (或新建 `web/admin/src/api/httpServices.js`)

**Files Context:**
- 现有 API 结构参考 `web/admin/src/api/toolBindings.js`

- [ ] **Step 1: 新增 getServiceTools API 方法**

在 `web/admin/src/api/` 下找到对应的 API 文件，添加：

```javascript
export const httpServicesApi = {
  // ... existing methods if file exists

  getServiceTools(serviceId) {
    return request.get(`/http-services/${serviceId}/tools`)
  },
}
```

如果 `httpServices.js` 不存在，参考 `toolBindings.js` 结构创建。

- [ ] **Step 2: 在 stores 中引用（如需要）**

如果使用 store 管理状态，可参考 `toolBindingsStore` 添加方法。

---

## Task 3: 前端组件 - 新增 ServiceToolsModal.vue

**Files:**
- Create: `web/admin/src/components/ServiceToolsModal.vue`

- [ ] **Step 1: 创建 ServiceToolsModal.vue**

```vue
<template>
  <teleport to="body">
    <transition name="fade">
      <div v-if="visible" class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 bg-black bg-opacity-50" @click="$emit('close')"></div>
          <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md max-h-[70vh] overflow-hidden fade-in">
            <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <div>
                <h3 class="text-lg font-semibold text-gray-900">关联工具</h3>
                <p v-if="service" class="text-sm text-gray-500 mt-0.5">服务: {{ service.name }}</p>
              </div>
              <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <div class="p-6 overflow-y-auto max-h-[calc(70vh-130px)]">
              <div v-if="loading" class="text-center py-8">
                <div class="loading-spinner mx-auto"></div>
                <p class="text-gray-500 mt-2">加载中...</p>
              </div>
              <div v-else-if="tools.length === 0" class="text-center py-8 text-gray-400">
                <svg class="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"/>
                </svg>
                <p>暂无关联工具</p>
              </div>
              <div v-else class="space-y-2">
                <div
                  v-for="tool in tools"
                  :key="tool.id"
                  @click="handleToolClick(tool)"
                  class="flex items-center gap-3 p-3 bg-gray-50 hover:bg-primary-50 rounded-lg cursor-pointer transition-colors"
                >
                  <div class="w-8 h-8 bg-primary-100 rounded-lg flex items-center justify-center">
                    <svg class="w-4 h-4 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                    </svg>
                  </div>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-medium text-gray-900 truncate">{{ tool.name }}</p>
                    <p v-if="tool.description" class="text-xs text-gray-500 truncate">{{ tool.description }}</p>
                  </div>
                  <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                  </svg>
                </div>
              </div>
            </div>
            <div class="px-6 py-4 border-t border-gray-200 flex justify-end">
              <button @click="$emit('close')"
                class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50">
                关闭
              </button>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { httpServicesApi } from '@/api/httpServices'

const props = defineProps({
  visible: { type: Boolean, default: false },
  service: { type: Object, default: null }
})

const emit = defineEmits(['close'])

const router = useRouter()
const tools = ref([])
const loading = ref(false)

watch(() => props.visible, async (newVal) => {
  if (newVal && props.service) {
    await loadTools()
  }
})

const loadTools = async () => {
  if (!props.service) return
  loading.value = true
  try {
    const res = await httpServicesApi.getServiceTools(props.service.id)
    const data = res.data || res
    tools.value = data.tools || []
  } catch (e) {
    console.error('加载工具列表失败:', e)
    tools.value = []
  } finally {
    loading.value = false
  }
}

const handleToolClick = (tool) => {
  router.push({ path: '/tools', query: { editId: String(tool.id) } })
}
</script>
```

---

## Task 4: 前端页面 - ServicesPage.vue 添加按钮和弹框

**Files:**
- Modify: `web/admin/src/pages/ServicesPage.vue`

- [ ] **Step 1: 在 script setup 中导入新组件**

在 `ServicesPage.vue` 的 `<script setup>` 中添加：

```javascript
import ServiceToolsModal from '@/components/ServiceToolsModal.vue'
```

- [ ] **Step 2: 添加状态变量**

在 script setup 中添加：

```javascript
const showToolsModal = ref(false)
const selectedServiceForTools = ref(null)
```

- [ ] **Step 3: 添加打开弹框方法**

在 script setup 中添加：

```javascript
const openToolsModal = (service) => {
  selectedServiceForTools.value = service
  showToolsModal.value = true
}
```

- [ ] **Step 4: 在服务卡片操作栏添加"查看工具"按钮**

在服务卡片的操作栏中（大约第 110-128 行附近），在"调试"按钮后添加：

```html
<button @click="openToolsModal(service)"
  class="px-3 py-1.5 text-xs font-medium text-indigo-600 hover:bg-indigo-50 rounded-lg transition-colors">
  查看工具
</button>
```

- [ ] **Step 5: 在模板底部添加弹框组件**

在 `</div>` 结束标签前（文件末尾的 `</template>` 关闭标签前），添加：

```html
<!-- Service Tools Modal -->
<ServiceToolsModal
  :visible="showToolsModal"
  :service="selectedServiceForTools"
  @close="showToolsModal = false"
/>
```

---

## Task 5: 前端页面 - ToolsPage.vue 支持 editId 参数

**Files:**
- Modify: `web/admin/src/pages/ToolsPage.vue`

- [ ] **Step 1: 在 onMounted 中处理 editId 参数**

在 `onMounted` 钩子中，添加对 URL query 参数的检测：

```javascript
import { useRoute } from 'vue-router'

const route = useRoute()

// ...

const checkEditId = () => {
  const editId = route.query.editId
  if (editId) {
    const tool = toolsStore.tools.find(t => String(t.id) === String(editId))
    if (tool) {
      openEditToolDialog(tool)
    }
  }
}

onMounted(() => {
  toolsStore.fetchToolsForAdmin()
  checkEditId()
})
```

- [ ] **Step 2: 在 watch route.query 以处理跳转后的编辑**

添加 watch 监听：

```javascript
watch(() => route.query.editId, (newEditId) => {
  if (newEditId) {
    const tool = toolsStore.tools.find(t => String(t.id) === String(newEditId))
    if (tool) {
      openEditToolDialog(tool)
    }
  }
})
```

---

## Task 6: 文档 - 更新 http-services.md

**Files:**
- Modify: `docs/api/http-services.md` (如果不存在则创建)

**Files Context:**
- 参考 `docs/api/tools.md` 的格式

- [ ] **Step 1: 添加新接口文档**

```markdown
### GET /api/admin/http-services/:id/tools

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
      { "id": 1, "name": "get_user", "description": "获取用户信息" },
      { "id": 2, "name": "create_user", "description": "创建用户" }
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
| tools[].description | string | 工具描述 |
| count | int | 工具数量 |
```

---

## 实施检查清单

- [ ] 后端 GetHTTPServiceTools 方法已添加
- [ ] 路由已注册
- [ ] 前端 API 方法已添加
- [ ] ServiceToolsModal.vue 组件已创建
- [ ] ServicesPage.vue 按钮和弹框已集成
- [ ] ToolsPage.vue editId 参数支持已添加
- [ ] API 文档已更新
- [ ] 功能测试通过