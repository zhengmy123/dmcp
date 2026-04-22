<template>
  <teleport to="body">
    <transition name="fade">
      <div v-if="visible" class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 bg-black bg-opacity-50" @click="$emit('close')"></div>
          <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-6xl max-h-[90vh] overflow-hidden fade-in">
            <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <div class="flex items-center gap-3">
                <h3 class="text-lg font-semibold text-gray-900">MCP 调试</h3>
                <span v-if="server" class="px-2.5 py-1 text-xs font-medium bg-primary-100 text-primary-700 rounded-full">
                  {{ server.name }}
                </span>
              </div>
              <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>

            <div v-if="server" class="flex flex-col h-[calc(90vh-80px)]">
              <div class="px-6 py-4 border-b border-gray-100 bg-gray-50">
                <div class="grid grid-cols-2 gap-6">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">MCP 端点 URL</label>
                    <div class="flex items-center gap-2">
                      <div class="flex-1 flex items-center bg-white border border-gray-200 rounded-lg px-3 py-2">
                        <span class="text-xs text-primary-600 font-medium mr-2">POST</span>
                        <input v-model="mcpUrl" type="text"
                          class="flex-1 text-sm text-gray-700 font-mono bg-transparent outline-none"
                          placeholder="https://...">
                      </div>
                      <button
                        @click="fetchToolsList"
                        :disabled="loadingTools"
                        class="p-2 text-gray-400 hover:text-primary-600 hover:bg-primary-50 rounded-lg transition-colors"
                        title="刷新工具列表"
                      >
                        <svg class="w-4 h-4" :class="{ 'animate-spin': loadingTools }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                        </svg>
                      </button>
                    </div>
                  </div>
                  <div>
                    <div class="flex items-center justify-between mb-2">
                      <label class="block text-sm font-medium text-gray-700">请求 Headers</label>
                      <button @click="addHeader" class="text-xs text-primary-600 hover:text-primary-700 flex items-center gap-1">
                        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                        </svg>
                        添加
                      </button>
                    </div>
                    <div class="bg-white border border-gray-200 rounded-lg max-h-24 overflow-y-auto">
                      <div v-for="(header, index) in headers" :key="index" class="flex items-center gap-2 px-3 py-2 border-b border-gray-100 last:border-b-0">
                        <input v-model="header.key" type="text" placeholder="Key"
                          class="flex-1 text-sm text-gray-700 bg-transparent outline-none placeholder-gray-300 border border-gray-200 rounded px-2 py-1">
                        <span class="text-gray-400">:</span>
                        <input v-model="header.value" type="text" placeholder="Value"
                          class="flex-1 text-sm text-gray-700 bg-transparent outline-none placeholder-gray-300 border border-gray-200 rounded px-2 py-1">
                        <button @click="removeHeader(index)" class="p-1 text-gray-300 hover:text-red-500 transition-colors">
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="flex items-center gap-2 px-6 py-3 bg-white border-b border-gray-100">
                <button
                  v-for="tab in tabs"
                  :key="tab.method"
                  @click="switchTab(tab.method)"
                  class="px-4 py-1.5 text-sm font-medium rounded-lg transition-colors"
                  :class="activeTab === tab.method
                    ? 'bg-primary-600 text-white'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
                >
                  {{ tab.label }}
                </button>
              </div>

              <div class="flex flex-1 min-h-0">
                <div class="w-1/2 border-r border-gray-200 flex flex-col bg-white">
                  <div class="px-6 py-3 border-b border-gray-100 flex items-center justify-between">
                    <span class="text-sm font-medium text-gray-700">请求</span>
                  </div>
                  <div class="flex-1 p-6 overflow-y-auto">
                    <div v-if="activeTab === 'tools/list'" class="h-full flex flex-col items-center justify-center text-gray-400">
                      <svg class="w-12 h-12 mb-3 text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
                      </svg>
                      <p class="text-sm">无需参数配置</p>
                    </div>

                    <div v-else class="space-y-4">
                      <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">工具名称</label>
                        <input v-model="toolName" type="text"
                          class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                          placeholder="输入工具名称，如: search_users">
                      </div>
                      <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">调用参数 (JSON)</label>
                        <textarea v-model="toolArguments"
                          class="w-full h-48 px-3 py-2 text-sm font-mono border border-gray-200 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 resize-none"
                          placeholder='{"query": "xxx", "limit": 10}'></textarea>
                      </div>
                    </div>
                  </div>
                  <div class="px-6 py-4 border-t border-gray-100 bg-gray-50">
                    <button
                      @click="sendRequest"
                      :disabled="sending || !mcpUrl"
                      class="w-full px-4 py-2.5 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                    >
                      <svg v-if="sending" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"/>
                      </svg>
                      <span>{{ sending ? '发送中...' : '发送请求' }}</span>
                    </button>
                  </div>
                </div>

                <div class="w-1/2 flex flex-col bg-white">
                  <div class="px-6 py-3 border-b border-gray-100 flex items-center justify-between">
                    <span class="text-sm font-medium text-gray-700">响应</span>
                    <div class="flex items-center gap-4">
                      <span v-if="responseTime !== null" class="text-xs text-gray-500">
                        {{ responseTime }}ms
                      </span>
                      <span v-if="response" class="flex items-center gap-1.5 text-xs font-medium" :class="responseError ? 'text-red-600' : 'text-green-600'">
                        <span class="w-2 h-2 rounded-full" :class="responseError ? 'bg-red-500' : 'bg-green-500'"></span>
                        {{ responseError ? '失败' : '成功' }}
                      </span>
                    </div>
                  </div>
                  <div class="flex-1 p-6 overflow-hidden">
                    <div v-if="!response" class="h-full flex flex-col items-center justify-center text-gray-400">
                      <svg class="w-12 h-12 mb-3 text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
                      </svg>
                      <p class="text-sm">发送请求后显示响应</p>
                    </div>
                    <div v-else class="h-full relative">
                      <pre class="h-full w-full p-4 text-xs font-mono bg-gray-800 text-green-400 rounded-lg overflow-auto whitespace-pre-wrap">{{ formattedResponse }}</pre>
                      <button
                        @click="copyResponse"
                        class="absolute top-3 right-3 px-2.5 py-1 text-xs bg-gray-700 text-gray-300 rounded hover:bg-gray-600 transition-colors flex items-center gap-1"
                      >
                        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                        </svg>
                        {{ copied ? '已复制' : '复制' }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div v-else class="p-12 text-center">
              <p class="text-gray-500">请选择一个 Server</p>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup>
import { ref, computed, watch, reactive } from 'vue'
import { systemConfigApi } from '@/api/systemConfig'

const props = defineProps({
  visible: { type: Boolean, default: false },
  server: { type: Object, default: null }
})

const emit = defineEmits(['close'])

const API_HOST_KEY = 'api_host'
const mcpServerUrl = ref('')
const mcpUrl = ref('')

const tabs = [
  { method: 'tools/list', label: 'tools/list' },
  { method: 'tools/call', label: 'tools/call' }
]
const activeTab = ref('tools/list')
const toolName = ref('')
const toolArguments = ref('{}')

const headers = reactive([
  { key: '', value: '' }
])

const sending = ref(false)
const loadingTools = ref(false)
const response = ref(null)
const responseError = ref(false)
const responseTime = ref(null)
const copied = ref(false)

let requestId = 1
let sessionId = null

watch(() => props.visible, async (newVal) => {
  if (newVal && props.server) {
    await initMcpUrl()
    resetRequest()
    await fetchToolsList()
  }
})

watch(() => props.server, async (newVal) => {
  if (newVal && props.visible) {
    await initMcpUrl()
    resetRequest()
    await fetchToolsList()
  }
})

const initMcpUrl = async () => {
  try {
    const res = await systemConfigApi.getConfig(API_HOST_KEY)
    if (res.data && res.data.config_value) {
      mcpServerUrl.value = res.data.config_value
      mcpUrl.value = mcpServerUrl.value.replace(/\/$/, '') + '/mcp/' + props.server.vauth_key
    }
  } catch (e) {
    console.error('failed to load mcp server url:', e)
  }
}

const switchTab = (method) => {
  activeTab.value = method
}

const resetRequest = () => {
  activeTab.value = 'tools/list'
  toolName.value = ''
  toolArguments.value = '{}'
  headers.length = 0
  headers.push({ key: '', value: '' })
  response.value = null
  responseError.value = false
  responseTime.value = null
  requestId = 1
  sessionId = null
}

const addHeader = () => {
  headers.push({ key: '', value: '' })
}

const removeHeader = (index) => {
  headers.splice(index, 1)
  if (headers.length === 0) {
    headers.push({ key: '', value: '' })
  }
}

const buildHeaders = () => {
  const result = {
    'Content-Type': 'application/json'
  }
  for (const h of headers) {
    if (h.key.trim()) {
      result[h.key.trim()] = h.value
    }
  }
  if (sessionId) {
    result['Mcp-Session-Id'] = sessionId
  }
  return result
}

const fetchToolsList = async () => {
  if (!mcpUrl.value) return

  loadingTools.value = true
  const startTime = Date.now()

  try {
    const body = {
      jsonrpc: '2.0',
      id: requestId++,
      method: 'tools/list',
      params: {}
    }

    const res = await fetch(mcpUrl.value, {
      method: 'POST',
      headers: buildHeaders(),
      body: JSON.stringify(body)
    })

    const sessionIdHeader = res.headers.get('mcp-session-id')
    if (sessionIdHeader) {
      sessionId = sessionIdHeader
    }

    const data = await res.json()
    responseTime.value = Date.now() - startTime

    if (data.error) {
      responseError.value = true
      response.value = data.error
    } else {
      responseError.value = false
      response.value = data
    }
  } catch (e) {
    responseTime.value = Date.now() - startTime
    responseError.value = true
    response.value = { message: e.message || '请求失败' }
  } finally {
    loadingTools.value = false
  }
}

const sendRequest = async () => {
  if (!mcpUrl.value) return

  sending.value = true
  const startTime = Date.now()

  try {
    let params = {}

    if (activeTab.value === 'tools/call') {
      try {
        params = JSON.parse(toolArguments.value || '{}')
      } catch (e) {
        responseError.value = true
        response.value = { message: 'JSON 参数格式错误' }
        responseTime.value = Date.now() - startTime
        sending.value = false
        return
      }

      if (!toolName.value) {
        responseError.value = true
        response.value = { message: '请输入工具名称' }
        responseTime.value = Date.now() - startTime
        sending.value = false
        return
      }

      params = {
        name: toolName.value,
        arguments: params
      }
    }

    const body = {
      jsonrpc: '2.0',
      id: requestId++,
      method: activeTab.value,
      params
    }

    const res = await fetch(mcpUrl.value, {
      method: 'POST',
      headers: buildHeaders(),
      body: JSON.stringify(body)
    })

    const sessionIdHeader = res.headers.get('mcp-session-id')
    if (sessionIdHeader) {
      sessionId = sessionIdHeader
    }

    const data = await res.json()
    responseTime.value = Date.now() - startTime

    if (data.error) {
      responseError.value = true
      response.value = data.error
    } else {
      responseError.value = false
      response.value = data
    }
  } catch (e) {
    responseTime.value = Date.now() - startTime
    responseError.value = true
    response.value = { message: e.message || '请求失败' }
  } finally {
    sending.value = false
  }
}

const formattedResponse = computed(() => {
  if (!response.value) return ''
  return JSON.stringify(response.value, null, 2)
})

const copyResponse = async () => {
  try {
    await navigator.clipboard.writeText(formattedResponse.value)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 1500)
  } catch (e) {
    console.error('复制失败:', e)
  }
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.fade-in {
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
