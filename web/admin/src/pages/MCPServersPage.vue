<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">MCP Server 管理</h2>
        <p class="text-sm text-gray-500 mt-1">管理 MCP Server 配置及关联工具</p>
      </div>
      <div class="flex items-center space-x-2">
        <button
          @click="refreshServers"
          class="inline-flex items-center px-3 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <svg class="w-4 h-4 mr-1.5" :class="{ 'animate-spin': mcpServersStore.loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
          刷新
        </button>
        <button
          @click="openCreateModal"
          class="inline-flex items-center px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 btn-transition"
        >
          <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
          </svg>
          创建 Server
        </button>
      </div>
    </div>

    <!-- Search Filters -->
    <div class="bg-white rounded-xl border border-gray-200 p-4">
      <div class="flex flex-wrap gap-4">
        <div class="flex-1 min-w-[200px]">
          <input
            v-model="searchForm.name"
            type="text"
            placeholder="搜索名称..."
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            @keyup.enter="handleSearch"
          >
        </div>
        <div class="w-40">
          <select
            v-model="searchForm.state"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            @change="handleSearch"
          >
            <option value="">全部状态</option>
            <option :value="1">正常</option>
            <option :value="0">已删除</option>
          </select>
        </div>
        <button
          @click="handleSearch"
          class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700"
        >
          搜索
        </button>
        <button
          @click="handleReset"
          class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50"
        >
          重置
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="mcpServersStore.loading" class="text-center py-12">
      <div class="loading-spinner mx-auto"></div>
      <p class="text-gray-500 mt-2">加载中...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="mcpServersStore.servers.length === 0" class="text-center py-12 bg-white rounded-xl border border-gray-200">
      <svg class="w-12 h-12 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
      </svg>
      <p class="text-gray-500">暂无 MCP Server</p>
      <button @click="openCreateModal" class="mt-4 text-primary-600 hover:text-primary-700 text-sm font-medium">
        创建第一个 Server
      </button>
    </div>

    <!-- Servers Table -->
    <div v-else class="bg-white rounded-xl border border-gray-200 overflow-hidden">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">名称</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">VAuth Key</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">工具数</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">描述</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="server in mcpServersStore.servers" :key="server.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <div class="w-8 h-8 bg-primary-100 rounded-lg flex items-center justify-center mr-3">
                  <svg class="w-4 h-4 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
                  </svg>
                </div>
                <span class="font-medium text-gray-900">{{ server.name }}</span>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="flex items-center">
                <button
                  @click="copyVAuthKey(server.vauth_key, server.id)"
                  class="mr-1 p-1 rounded transition-colors"
                  :class="copiedId === server.id ? 'text-green-600' : 'text-gray-400 hover:text-primary-600 hover:bg-primary-50'"
                  title="复制 VAuth Key"
                >
                  <svg v-if="copiedId !== server.id" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                  </svg>
                  <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                  </svg>
                </button>
                <code class="text-sm text-gray-600 bg-gray-100 px-2 py-0.5 rounded">{{ server.vauth_key }}</code>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span class="px-2.5 py-1 text-xs font-medium bg-blue-100 text-blue-700 rounded-full">
                {{ server.tool_count || 0 }} 个工具
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span
                class="px-2.5 py-1 text-xs font-medium rounded-full"
                :class="server.state === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'"
              >
                {{ server.state === 1 ? '正常' : '已删除' }}
              </span>
            </td>
            <td class="px-6 py-4">
              <span class="text-sm text-gray-500 line-clamp-1">{{ server.description || '无描述' }}</span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm">
              <div class="flex items-center justify-end space-x-2">
                <button
                  @click="openBindingDialog(server)"
                  class="px-3 py-1 text-xs font-medium rounded-lg transition-colors bg-blue-100 text-blue-700 hover:bg-blue-200"
                  title="管理工具"
                >
                  管理工具
                </button>
                <button
                  @click="openEditModal(server)"
                  class="p-1.5 text-gray-400 hover:text-primary-600 hover:bg-primary-50 rounded-lg transition-colors"
                  title="编辑"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                  </svg>
                </button>
                <button
                  v-if="server.state === 1"
                  @click="handleDelete(server)"
                  class="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                  title="删除"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                  </svg>
                </button>
                <button
                  v-else
                  @click="handleRestore(server)"
                  class="p-1.5 text-green-600 hover:bg-green-50 rounded-lg transition-colors"
                  title="启用"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create/Edit Modal -->
    <teleport to="body">
      <transition name="fade">
        <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto">
          <div class="flex items-center justify-center min-h-screen px-4">
            <div class="fixed inset-0 bg-black bg-opacity-50" @click="showModal = false"></div>
            <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg max-h-[90vh] overflow-hidden fade-in">
              <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
                <h3 class="text-lg font-semibold text-gray-900">{{ editingServer ? '编辑 Server' : '创建 Server' }}</h3>
                <button @click="showModal = false" class="text-gray-400 hover:text-gray-600">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <div class="p-6 overflow-y-auto max-h-[calc(90vh-130px)]">
                <form @submit.prevent="handleSubmit" class="space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">名称 *</label>
                    <input v-model="form.name" type="text" required
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      placeholder="MCP Server 名称">
                  </div>

                  <div v-if="!editingServer">
                    <label class="block text-sm font-medium text-gray-700 mb-1">类型 *</label>
                    <select v-model="form.type" required
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500">
                      <option value="http_service">HTTP Service</option>
                      <option value="proxy">MCP Proxy</option>
                    </select>
                    <p class="text-xs text-gray-400 mt-1">创建后不可修改</p>
                  </div>

                  <div v-if="editingServer">
                    <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
                    <input :value="form.type === 'http_service' ? 'HTTP Service' : 'MCP Proxy'" type="text" disabled
                      class="w-full px-3 py-2 border border-gray-200 bg-gray-50 rounded-lg text-gray-500">
                    <p class="text-xs text-gray-400 mt-1">创建后不可修改</p>
                  </div>

                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">描述</label>
                    <textarea v-model="form.description" rows="2"
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      placeholder="可选描述信息"></textarea>
                  </div>

                  <div v-if="form.type === 'proxy'">
                    <label class="block text-sm font-medium text-gray-700 mb-1">代理地址</label>
                    <input v-model="form.http_server_url" type="url"
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      placeholder="https://api.example.com/mcp">
                    <p class="text-xs text-gray-400 mt-1">MCP 代理服务地址</p>
                  </div>

                  <div v-if="form.type === 'proxy'">
                    <label class="block text-sm font-medium text-gray-700 mb-1">认证 Header</label>
                    <input v-model="form.auth_header" type="text"
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      placeholder="Authorization: Bearer xxx">
                    <p class="text-xs text-gray-400 mt-1">可选，代理认证信息</p>
                  </div>

                  <div v-if="form.type === 'proxy'">
                    <label class="block text-sm font-medium text-gray-700 mb-1">超时时间（秒）</label>
                    <input v-model.number="form.timeout_seconds" type="number" min="1" max="300"
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      placeholder="30">
                    <p class="text-xs text-gray-400 mt-1">请求超时时间，默认 30 秒</p>
                  </div>

                  <div v-if="form.type === 'proxy'">
                    <label class="block text-sm font-medium text-gray-700 mb-1">额外 Headers</label>
                    <textarea v-model="form.extra_headers" rows="2"
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      placeholder="X-Custom-Header: value"></textarea>
                    <p class="text-xs text-gray-400 mt-1">可选，JSON 格式额外请求头</p>
                  </div>

                  <div class="flex justify-end space-x-3 pt-4">
                    <button type="button" @click="showModal = false"
                      class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50">
                      取消
                    </button>
                    <button type="submit" :disabled="mcpServersStore.loading"
                      class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50">
                      {{ mcpServersStore.loading ? '保存中...' : '保存' }}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </teleport>

    <!-- Server Binding Dialog -->
    <ServerBindingDialog
      :visible="showBindingDialog"
      :selected-server="selectedServerForBinding"
      @close="showBindingDialog = false"
    />

    <!-- Confirm Dialog -->
    <ConfirmDialog ref="confirmDialog" />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useMCPServersStore } from '@/stores/mcpServers'
import ServerBindingDialog from '@/components/ServerBindingDialog.vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'

const mcpServersStore = useMCPServersStore()

const showModal = ref(false)
const editingServer = ref(null)
const showBindingDialog = ref(false)
const selectedServerForBinding = ref(null)
const confirmDialog = ref(null)
const copiedId = ref(null)

const copyVAuthKey = async (text, id) => {
  try {
    await navigator.clipboard.writeText(text)
    copiedId.value = id
    setTimeout(() => {
      copiedId.value = null
    }, 1500)
  } catch (e) {
    console.error('复制失败:', e)
  }
}

const searchForm = reactive({
  name: '',
  state: ''
})

const handleSearch = () => {
  mcpServersStore.fetchServers({
    name: searchForm.name,
    state: searchForm.state === '' ? undefined : parseInt(searchForm.state)
  })
}

const handleReset = () => {
  searchForm.name = ''
  searchForm.state = ''
  mcpServersStore.fetchServers({})
}

const form = reactive({
  name: '',
  type: 'http_service',
  vauth_key: '',
  description: '',
  http_server_url: '',
  auth_header: '',
  timeout_seconds: 30,
  extra_headers: ''
})

const openBindingDialog = (server) => {
  selectedServerForBinding.value = server
  showBindingDialog.value = true
}

const refreshServers = () => {
  mcpServersStore.fetchServers()
}

const openCreateModal = () => {
  editingServer.value = null
  Object.assign(form, {
    name: '',
    type: 'http_service',
    vauth_key: '',
    description: '',
    http_server_url: '',
    auth_header: '',
    timeout_seconds: 30,
    extra_headers: ''
  })
  showModal.value = true
}

const openEditModal = (server) => {
  editingServer.value = server
  Object.assign(form, {
    name: server.name,
    type: server.type || 'http_service',
    vauth_key: server.vauth_key,
    description: server.description || '',
    http_server_url: server.http_server_url || '',
    auth_header: server.auth_header || '',
    timeout_seconds: server.timeout_seconds || 30,
    extra_headers: server.extra_headers || ''
  })
  showModal.value = true
}

const handleSubmit = async () => {
  try {
    if (editingServer.value) {
      const { type, vauth_key, ...updateData } = form
      await mcpServersStore.updateServer(editingServer.value.id, updateData)
    } else {
      const { vauth_key, ...createData } = form
      await mcpServersStore.createServer(createData)
    }
    showModal.value = false
  } catch (e) {
    console.error('保存失败:', e)
  }
}

const handleDelete = async (server) => {
  const confirmed = await confirmDialog.value.show({
    title: '确认删除',
    message: `确定要删除 Server "${server.name}" 吗？\n\n删除后该服务器将被禁用，但可以重新启用。`,
    type: 'danger',
    confirmText: '删除',
    cancelText: '取消'
  })
  if (!confirmed) return
  try {
    await mcpServersStore.deleteServer(server.id)
  } catch (e) {
    console.error('删除失败:', e)
  }
}

const handleRestore = async (server) => {
  const confirmed = await confirmDialog.value.show({
    title: '确认启用',
    message: `确定要启用 Server "${server.name}" 吗？\n\n启用后该服务器将恢复正常使用。`,
    type: 'info',
    confirmText: '启用',
    cancelText: '取消'
  })
  if (!confirmed) return
  try {
    await mcpServersStore.restoreServer(server.id)
  } catch (e) {
    console.error('启用失败:', e)
  }
}

onMounted(() => {
  mcpServersStore.fetchServers()
})
</script>
