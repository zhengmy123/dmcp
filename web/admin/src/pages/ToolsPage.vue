<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">工具定义</h2>
        <p class="text-sm text-gray-500 mt-1">管理所有 MCP 工具定义</p>
      </div>
      <div class="flex items-center space-x-3">
        <div class="relative">
          <input
            v-model="searchInput"
            type="text"
            placeholder="搜索工具名称..."
            class="w-64 pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            @keyup.enter="handleSearch"
          >
          <svg class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
          </svg>
        </div>
        <button
          @click="handleSearch"
          class="px-3 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
        >
          搜索
        </button>
        <button
          @click="refreshTools"
          class="inline-flex items-center px-3 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <svg class="w-4 h-4 mr-1.5" :class="{ 'animate-spin': toolsStore.loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
          刷新
        </button>
        <button
          @click="openCreateToolDialog"
          class="inline-flex items-center px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 btn-transition"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
          </svg>
          添加工具
        </button>
      </div>
    </div>

    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">名称</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">描述</th>
              <th class="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">参数数</th>
              <th class="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">入参映射</th>
              <th class="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">出参映射</th>
              <th class="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">创建时间</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-if="toolsStore.loading && toolsStore.tools.length === 0">
              <td colspan="9" class="px-6 py-12 text-center">
                <div class="loading-spinner mx-auto"></div>
                <p class="text-gray-500 mt-2">加载中...</p>
              </td>
            </tr>
            <tr v-else-if="toolsStore.tools.length === 0">
              <td colspan="9" class="px-6 py-12 text-center">
                <svg class="w-12 h-12 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"/>
                </svg>
                <p class="text-gray-500">暂无工具定义</p>
                <p class="text-sm text-gray-400 mt-1">点击"添加工具"创建第一个工具</p>
              </td>
            </tr>
            <tr v-for="tool in toolsStore.tools" :key="tool.id" class="hover:bg-gray-50 transition-colors">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{ tool.id }}</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <div class="w-8 h-8 bg-primary-100 rounded-lg flex items-center justify-center mr-3">
                    <svg class="w-4 h-4 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                    </svg>
                  </div>
                  <span class="font-medium text-gray-900">{{ tool.name }}</span>
                </div>
              </td>
              <td class="px-6 py-4">
                <span class="text-sm text-gray-500 line-clamp-1">{{ tool.description || '-' }}</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-center">
                <span class="px-2.5 py-1 text-xs font-medium bg-blue-100 text-blue-700 rounded-full">
                  {{ tool.parameters?.length || 0 }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-center">
                <span class="px-2.5 py-1 text-xs font-medium bg-amber-100 text-amber-700 rounded-full">
                  {{ tool.input_mapping?.length || 0 }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-center">
                <span class="px-2.5 py-1 text-xs font-medium bg-purple-100 text-purple-700 rounded-full">
                  {{ tool.output_mapping?.length || 0 }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-center">
                <span
                  class="px-2.5 py-1 text-xs font-medium rounded-full"
                  :class="tool.state === 1 ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-600'"
                >
                  {{ tool.state === 1 ? '启用' : '禁用' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ formatDate(tool.created_at) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
                <div class="flex items-center justify-end space-x-2">
                  <button
                    @click="openBindingDialog(tool)"
                    class="px-3 py-1 text-xs font-medium rounded-lg transition-colors bg-primary-100 text-primary-700 hover:bg-primary-200"
                    title="管理绑定"
                  >
                    管理绑定
                  </button>
                  <button
                    @click="openEditToolDialog(tool)"
                    class="p-1.5 text-gray-400 hover:text-primary-600 hover:bg-primary-50 rounded-lg transition-colors"
                    title="编辑"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                    </svg>
                  </button>
                  <button
                    @click="handleDelete(tool)"
                    class="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                    title="删除"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="toolsStore.pagination.total > 0" class="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
        <div class="text-sm text-gray-500">
          共 <span class="font-medium">{{ toolsStore.pagination.total }}</span> 条记录，第
          <span class="font-medium">{{ toolsStore.pagination.page }}</span> /
          <span class="font-medium">{{ totalPages }}</span> 页
        </div>
        <div class="flex items-center space-x-2">
          <button
            @click="handlePageChange(toolsStore.pagination.page - 1)"
            :disabled="toolsStore.pagination.page <= 1"
            class="px-3 py-1.5 text-sm border border-gray-300 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 transition-colors"
          >
            上一页
          </button>
          <div class="flex items-center space-x-1">
            <button
              v-for="page in visiblePages"
              :key="page"
              @click="handlePageChange(page)"
              class="w-8 h-8 text-sm rounded-lg transition-colors"
              :class="page === toolsStore.pagination.page
                ? 'bg-primary-600 text-white'
                : 'border border-gray-300 hover:bg-gray-50'"
            >
              {{ page }}
            </button>
          </div>
          <button
            @click="handlePageChange(toolsStore.pagination.page + 1)"
            :disabled="toolsStore.pagination.page >= totalPages"
            class="px-3 py-1.5 text-sm border border-gray-300 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 transition-colors"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <ToolEditDialog
      :visible="showToolDialog"
      :editing-tool="editingTool"
      :save-error="saveError"
      @close="showToolDialog = false"
      @saved="handleToolSaved"
      @error-cleared="saveError = ''"
    />

    <ToolBindingDialog
      :visible="showBindingDialog"
      :selected-tool="selectedToolForBinding"
      @close="showBindingDialog = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useToolsStore } from '@/stores/tools'
import { useToolBindingsStore } from '@/stores/toolBindings'
import ToolEditDialog from '@/components/ToolEditDialog.vue'
import ToolBindingDialog from '@/components/ToolBindingDialog.vue'

const toolsStore = useToolsStore()
const toolBindingsStore = useToolBindingsStore()
const showToolDialog = ref(false)
const editingTool = ref(null)
const showBindingDialog = ref(false)
const selectedToolForBinding = ref(null)
const searchInput = ref('')
const saveError = ref('')

const totalPages = computed(() => {
  return Math.ceil(toolsStore.pagination.total / toolsStore.pagination.pageSize) || 1
})

const visiblePages = computed(() => {
  const current = toolsStore.pagination.page
  const total = totalPages.value
  const pages = []

  if (total <= 7) {
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    if (current <= 4) {
      for (let i = 1; i <= 5; i++) pages.push(i)
      pages.push('...')
      pages.push(total)
    } else if (current >= total - 3) {
      pages.push(1)
      pages.push('...')
      for (let i = total - 4; i <= total; i++) pages.push(i)
    } else {
      pages.push(1)
      pages.push('...')
      for (let i = current - 1; i <= current + 1; i++) pages.push(i)
      pages.push('...')
      pages.push(total)
    }
  }

  return pages.filter((p, idx, arr) => arr.indexOf(p) === idx && p !== '...' || p === '...')
})

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const refreshTools = () => {
  toolsStore.fetchToolsForAdmin()
}

const handleSearch = () => {
  toolsStore.setKeyword(searchInput.value)
}

const handlePageChange = (page) => {
  if (page === '...' || page < 1 || page > totalPages.value) return
  toolsStore.setPage(page)
}

const openCreateToolDialog = () => {
  saveError.value = ''
  editingTool.value = null
  showToolDialog.value = true
}

const openEditToolDialog = (tool) => {
  saveError.value = ''
  editingTool.value = { ...tool }
  showToolDialog.value = true
}

const openBindingDialog = (tool) => {
  selectedToolForBinding.value = tool
  showBindingDialog.value = true
}

const handleToolSaved = async (data) => {
  let success = false
  saveError.value = ''
  if (editingTool.value) {
    success = await toolsStore.updateTool(editingTool.value.id, data)
  } else {
    success = await toolsStore.createTool(data)
  }

  if (success) {
    localStorage.removeItem('tool_edit_draft')
    showToolDialog.value = false
  } else {
    saveError.value = toolsStore.error || '保存失败，请稍后重试'
  }
}

const handleDelete = async (tool) => {
  if (!confirm(`确定要删除工具 "${tool.name}" 吗？`)) return

  const success = await toolsStore.deleteTool(tool.id)
  if (success) {
    toolsStore.fetchToolsForAdmin()
  }
}

onMounted(() => {
  toolsStore.fetchToolsForAdmin()
})
</script>

<style scoped>
.line-clamp-1 {
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>