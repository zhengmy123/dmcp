<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">工具定义</h2>
        <p class="text-sm text-gray-500 mt-1">查看所有 MCP 工具定义</p>
      </div>
      <div class="flex items-center space-x-2">
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

    <!-- Groups -->
    <div v-if="toolsStore.loading" class="text-center py-12">
      <div class="loading-spinner mx-auto"></div>
      <p class="text-gray-500 mt-2">加载中...</p>
    </div>
    
    <div v-else-if="Object.keys(toolsStore.groupedTools).length === 0" class="text-center py-12 bg-white rounded-xl border border-gray-200">
      <svg class="w-12 h-12 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"/>
      </svg>
      <p class="text-gray-500">暂无工具定义</p>
      <p class="text-sm text-gray-400 mt-1">请在数据库中添加工具定义</p>
    </div>

    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div
        v-for="(tools, group) in toolsStore.groupedTools"
        :key="group"
        class="bg-white rounded-xl border border-gray-200 overflow-hidden"
      >
        <!-- Group Header -->
        <div class="px-6 py-4 bg-gray-50 border-b border-gray-200 flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 bg-primary-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
              </svg>
            </div>
            <div>
              <h3 class="font-semibold text-gray-900">{{ group }}</h3>
              <p class="text-xs text-gray-500">{{ tools.length }} 个工具</p>
            </div>
          </div>
          <span class="px-2.5 py-1 text-xs font-medium bg-primary-100 text-primary-700 rounded-full">
            MCP Server
          </span>
        </div>

        <!-- Tools List -->
        <div class="divide-y divide-gray-100">
          <div
            v-for="tool in tools"
            :key="tool.name"
            class="p-4 hover:bg-gray-50 transition-colors cursor-pointer"
            @click="selectedTool = tool"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1">
                <h4 class="font-medium text-gray-900">{{ tool.name }}</h4>
                <p class="text-sm text-gray-500 mt-1">{{ tool.description || '无描述' }}</p>
                <div v-if="tool.parameters?.length" class="flex flex-wrap gap-1 mt-2">
                  <span
                    v-for="param in tool.parameters.slice(0, 4)"
                    :key="param.name"
                    class="px-2 py-0.5 text-xs bg-gray-100 text-gray-600 rounded"
                  >
                    {{ param.name }}
                    <span v-if="param.required" class="text-red-500">*</span>
                  </span>
                  <span v-if="tool.parameters.length > 4" class="text-xs text-gray-400">
                    +{{ tool.parameters.length - 4 }}
                  </span>
                </div>
              </div>
              <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
              </svg>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tool Detail Modal -->
    <teleport to="body">
      <transition name="fade">
        <div v-if="selectedTool" class="fixed inset-0 z-50 overflow-y-auto">
          <div class="flex items-center justify-center min-h-screen px-4">
            <div class="fixed inset-0 bg-black bg-opacity-50" @click="selectedTool = null"></div>
            <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-2xl max-h-[80vh] overflow-hidden fade-in">
              <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
                <h3 class="text-lg font-semibold text-gray-900">{{ selectedTool.name }}</h3>
                <button @click="selectedTool = null" class="text-gray-400 hover:text-gray-600">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <div class="p-6 overflow-y-auto max-h-[calc(80vh-130px)]">
                <div class="mb-4">
                  <span class="px-2.5 py-1 text-xs font-medium bg-primary-100 text-primary-700 rounded-full">
                    {{ selectedTool.vauth_key }}
                  </span>
                </div>
                <p class="text-gray-600 mb-6">{{ selectedTool.description || '无描述' }}</p>

                <h4 class="font-medium text-gray-900 mb-3">参数定义</h4>
                <div v-if="!selectedTool.parameters?.length" class="text-gray-500 text-sm">无参数</div>
                <div v-else class="space-y-3">
                  <div
                    v-for="param in selectedTool.parameters"
                    :key="param.name"
                    class="p-3 bg-gray-50 rounded-lg"
                  >
                    <div class="flex items-center space-x-2 mb-1">
                      <code class="font-mono text-sm font-medium text-gray-900">{{ param.name }}</code>
                      <span class="text-xs text-gray-500">{{ param.type }}</span>
                      <span v-if="param.required" class="px-1.5 py-0.5 text-xs bg-red-100 text-red-600 rounded">必填</span>
                    </div>
                    <p class="text-sm text-gray-500">{{ param.description }}</p>
                    <div v-if="param.default !== undefined" class="text-xs text-gray-400 mt-1">
                      默认值: {{ param.default }}
                    </div>
                  </div>
                </div>

                <h4 class="font-medium text-gray-900 mb-3 mt-6">JSON Schema</h4>
                <pre class="p-4 bg-gray-900 text-gray-100 rounded-lg text-sm overflow-x-auto">{{ formatJson(selectedTool.parameters) }}</pre>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </teleport>

    <!-- Tool Edit Dialog -->
    <ToolEditDialog
      :visible="showToolDialog"
      :editing-tool="editingTool"
      @close="showToolDialog = false"
      @saved="handleToolSaved"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useToolsStore } from '@/stores/tools'
import ToolEditDialog from '@/components/ToolEditDialog.vue'

const toolsStore = useToolsStore()
const selectedTool = ref(null)
const showToolDialog = ref(false)
const editingTool = ref(null)

const refreshTools = () => {
  toolsStore.fetchTools()
}

const formatJson = (obj) => {
  return JSON.stringify(obj, null, 2)
}

const openCreateToolDialog = () => {
  editingTool.value = null
  showToolDialog.value = true
}

const openEditToolDialog = (tool) => {
  editingTool.value = tool
  showToolDialog.value = true
}

const handleToolSaved = async (data) => {
  showToolDialog.value = false
  await toolsStore.fetchToolsForAdmin()
}

onMounted(() => {
  toolsStore.fetchTools()
})
</script>
