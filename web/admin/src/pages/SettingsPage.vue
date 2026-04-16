<template>
  <div class="max-w-3xl space-y-6">
    <!-- Header -->
    <div>
      <h2 class="text-lg font-semibold text-gray-900">系统设置</h2>
      <p class="text-sm text-gray-500 mt-1">配置 MCP Server 连接参数</p>
    </div>

    <!-- API Configuration -->
    <div class="bg-white rounded-xl border border-gray-200">
      <div class="px-6 py-4 border-b border-gray-200">
        <h3 class="font-semibold text-gray-900">API 配置</h3>
      </div>
      <div class="p-6 space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">服务地址</label>
          <input
            v-model="settings.serverUrl"
            type="url"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            placeholder="http://localhost:18080"
          >
          <p class="text-xs text-gray-500 mt-1">MCP Server 的访问地址</p>
        </div>
        <div class="flex justify-end">
          <button
            @click="saveApiSettings"
            class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700"
          >
            保存
          </button>
        </div>
      </div>
    </div>

    <!-- Connection Info -->
    <div class="bg-white rounded-xl border border-gray-200">
      <div class="px-6 py-4 border-b border-gray-200">
        <h3 class="font-semibold text-gray-900">数据库连接</h3>
      </div>
      <div class="p-6">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block text-xs font-medium text-gray-500 mb-1">主机</label>
            <div class="px-3 py-2 bg-gray-50 rounded-lg text-sm text-gray-700">127.0.0.1:3306</div>
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-500 mb-1">数据库</label>
            <div class="px-3 py-2 bg-gray-50 rounded-lg text-sm text-gray-700">mcp_server</div>
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-500 mb-1">用户名</label>
            <div class="px-3 py-2 bg-gray-50 rounded-lg text-sm text-gray-700">root</div>
          </div>
          <div>
            <label class="block text-xs font-medium text-gray-500 mb-1">表名</label>
            <div class="px-3 py-2 bg-gray-50 rounded-lg text-sm text-gray-700">mcp_tool_definitions</div>
          </div>
        </div>
      </div>
    </div>

    <!-- About -->
    <div class="bg-white rounded-xl border border-gray-200">
      <div class="px-6 py-4 border-b border-gray-200">
        <h3 class="font-semibold text-gray-900">关于</h3>
      </div>
      <div class="p-6">
        <div class="flex items-center space-x-4">
          <div class="w-12 h-12 bg-primary-100 rounded-xl flex items-center justify-center">
            <svg class="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"/>
            </svg>
          </div>
          <div>
            <h4 class="font-semibold text-gray-900">MCP Server</h4>
            <p class="text-sm text-gray-500">版本 1.0.0</p>
            <p class="text-xs text-gray-400 mt-1">Dynamic MCP Server with JWT Authentication</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, inject, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const showToast = inject('showToast')

const settings = reactive({
  serverUrl: authStore.serverUrl
})

const saveApiSettings = () => {
  authStore.updateServerUrl(settings.serverUrl)
  showToast('设置已保存', 'success')
}

onMounted(() => {
  settings.serverUrl = authStore.serverUrl
})
</script>
