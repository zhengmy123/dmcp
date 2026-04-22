<template>
  <teleport to="body">
    <transition name="fade">
      <div v-if="visible" class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 bg-black bg-opacity-50" @click="$emit('close')"></div>
          <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-6xl max-h-[98vh] overflow-hidden fade-in">
            <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <div class="flex items-center gap-3">
                <h3 class="text-lg font-semibold text-gray-900">MCP 调试</h3>
                <span v-if="server" class="px-2 py-0.5 text-xs font-medium bg-blue-100 text-blue-700 rounded-full">{{ server.name }}</span>
              </div>
              <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <div class="p-6 overflow-y-auto max-h-[calc(90vh-130px)]">
              <div v-if="!server" class="text-center py-12">
                <p class="text-gray-500">请选择一个 Server</p>
              </div>
              <div v-else class="grid grid-cols-2 gap-6">
                <!-- Left: Request -->
                <div class="space-y-4">
                  <h4 class="text-sm font-semibold text-gray-800">请求配置</h4>

                  <!-- MCP URL -->
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">MCP 端点</label>
                    <div class="flex items-center gap-2">
                      <span class="px-2 py-1 text-xs font-medium bg-primary-100 text-primary-700 rounded">POST</span>
                      <input v-model="mcpUrl" type="text"
                        class="flex-1 px-3 py-1.5 border border-gray-300 rounded-lg text-sm font-mono focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                        placeholder="https://...">
                    </div>
                  </div>

                  <!-- Request Headers -->
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">请求头</label>
                    <div v-for="(h, i) in headers" :key="i" class="flex gap-1 mb-1">
                      <input v-model="h.key" placeholder="Key" class="flex-1 px-2 py-1 border border-gray-300 rounded text-xs">
                      <input v-model="h.value" placeholder="Value" class="flex-1 px-2 py-1 border border-gray-300 rounded text-xs">
                      <button @click="removeHeader(i)" class="text-red-400 hover:text-red-600 px-1">
                        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                      </button>
                    </div>
                    <button @click="addHeader" class="text-xs text-primary-600 hover:text-primary-700">+ 添加</button>
                  </div>

                  <!-- Method Selector -->
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1.5">方法</label>
                    <div class="flex flex-wrap gap-1">
                      <button v-for="tab in tabs" :key="tab.method"
                        type="button"
                        @click="activeTab = tab.method"
                        :class="activeTab === tab.method
                          ? 'bg-primary-600 text-white shadow-sm'
                          : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
                        class="px-2.5 py-1 text-xs font-medium rounded-md transition-colors">
                        {{ tab.label }}
                      </button>
                    </div>
                  </div>

                  <!-- tools/call params -->
                  <div v-if="activeTab === 'tools/call'" class="space-y-3">
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">工具名称</label>
                      <input v-model="toolName" type="text"
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                        placeholder="输入工具名称，如: search_users">
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">
                        调用参数
                        <span class="text-xs text-gray-400 font-normal">(JSON)</span>
                      </label>
                      <textarea v-model="toolArguments" rows="6"
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-xs focus:ring-2 focus:ring-primary-500 focus:border-primary-500 resize-none"
                        placeholder='{"query": "xxx", "limit": 10}'
                        spellcheck="false"></textarea>
                      <div v-if="parsedToolArguments" class="mt-2 border border-gray-200 rounded-lg overflow-hidden max-h-48 overflow-y-auto">
                        <VueJsonPretty :data="parsedToolArguments" :deep="3" :show-line="false" />
                      </div>
                      <p v-if="toolArgumentsError" class="text-xs text-red-500 mt-1">{{ toolArgumentsError }}</p>
                    </div>
                  </div>

                  <div v-else class="bg-gray-50 rounded-lg p-4 text-center text-sm text-gray-400">
                    无需参数配置
                  </div>

                  <button @click="sendRequest"
                    :disabled="sending || !mcpUrl"
                    class="w-full px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50 flex items-center justify-center gap-2">
                    <svg v-if="sending" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    {{ sending ? '请求中...' : '发送请求' }}
                  </button>
                </div>

                <!-- Right: Response -->
                <div class="space-y-4">
                  <h4 class="text-sm font-semibold text-gray-800">响应结果</h4>
                  <div v-if="response" class="space-y-3">
                    <!-- Status -->
                    <div class="flex items-center gap-2">
                      <span class="text-sm font-medium">状态:</span>
                      <span class="px-2 py-0.5 text-xs font-medium rounded-full"
                        :class="responseError ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'">
                        {{ responseError ? '失败' : '成功' }}
                      </span>
                      <span v-if="responseTime !== null" class="text-xs text-gray-500">{{ responseTime }}ms</span>
                    </div>

                    <!-- Response Body -->
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">响应体</label>
                      <div v-if="response" class="border border-gray-200 rounded-lg overflow-hidden max-h-96 overflow-y-auto">
                        <VueJsonPretty v-if="isResponseObject" :data="response" :deep="3" :show-line="false" />
                        <pre v-else class="p-3 text-xs font-mono text-gray-700 whitespace-pre-wrap break-all">{{ formattedResponse }}</pre>
                      </div>
                      <div v-else class="bg-gray-50 rounded-lg p-3 text-xs text-gray-400">无响应体</div>
                    </div>
                  </div>
                  <div v-else class="flex flex-col items-center justify-center py-16 text-gray-400">
                    <svg class="w-12 h-12 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                    </svg>
                    <p class="text-sm">点击"发送请求"开始调试</p>
                  </div>
                </div>
              </div>
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
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'

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
const toolArguments = ref('')
const toolArgumentsError = ref('')

const headers = reactive([
  { key: '', value: '' }
])

const sending = ref(false)
const response = ref(null)
const responseError = ref(false)
const responseTime = ref(null)

let requestId = 1
let sessionId = null

const parsedToolArguments = computed(() => {
  if (!toolArguments.value.trim()) return null
  try {
    toolArgumentsError.value = ''
    return JSON.parse(toolArguments.value)
  } catch (e) {
    toolArgumentsError.value = 'JSON 格式错误'
    return null
  }
})

const isResponseObject = computed(() => {
  if (!response.value) return false
  return typeof response.value === 'object' && response.value !== null
})

const formattedResponse = computed(() => {
  if (!response.value) return ''
  return JSON.stringify(response.value, null, 2)
})

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

const resetRequest = () => {
  activeTab.value = 'tools/list'
  toolName.value = ''
  toolArguments.value = ''
  toolArgumentsError.value = ''
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

  sending.value = true
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
    sending.value = false
  }
}

const sendRequest = async () => {
  if (!mcpUrl.value) return

  sending.value = true
  const startTime = Date.now()

  try {
    let params = {}

    if (activeTab.value === 'tools/call') {
      if (!toolName.value) {
        responseError.value = true
        response.value = { message: '请输入工具名称' }
        responseTime.value = Date.now() - startTime
        sending.value = false
        return
      }

      try {
        params = parsedToolArguments.value || {}
      } catch (e) {
        responseError.value = true
        response.value = { message: 'JSON 参数格式错误' }
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
