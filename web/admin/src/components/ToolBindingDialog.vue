<template>
  <teleport to="body">
    <transition name="fade">
      <div v-if="visible" class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 bg-black bg-opacity-50" @click="$emit('close')"></div>
          <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-5xl max-h-[90vh] overflow-hidden fade-in">
            <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between bg-gradient-to-r from-primary-50 to-white">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 bg-primary-100 rounded-lg flex items-center justify-center">
                  <svg class="w-4 h-4 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/>
                  </svg>
                </div>
                <h3 class="text-lg font-semibold text-gray-900">工具绑定管理</h3>
                <span v-if="selectedTool" class="px-2.5 py-1 text-xs font-medium bg-primary-100 text-primary-700 rounded-full flex items-center gap-1">
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                  </svg>
                  {{ selectedTool.name }}
                </span>
              </div>
              <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600 p-1 rounded-lg hover:bg-gray-100 transition-colors">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <div class="p-6">
              <div v-if="!selectedTool" class="text-center py-12">
                <div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
                  </svg>
                </div>
                <p class="text-gray-500">请选择一个工具</p>
              </div>
              <div v-else class="flex gap-4 h-[65vh]">
                <div class="flex-1 flex flex-col">
                  <div class="mb-3 flex items-center justify-between">
                    <h4 class="text-sm font-semibold text-gray-700 flex items-center gap-2">
                      <span class="w-2 h-2 bg-green-500 rounded-full"></span>
                      可绑定 Server
                      <span class="text-xs font-normal text-gray-400">({{ filteredAvailableServers.length }})</span>
                    </h4>
                  </div>
                  <div class="relative mb-3">
                    <input v-model="leftSearch" type="text" placeholder="搜索名称或描述..."
                      class="w-full pl-9 pr-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-primary-100 focus:border-primary-400 bg-gray-50">
                    <svg class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                    </svg>
                    <button v-if="leftSearch" @click="leftSearch = ''" class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                      </svg>
                    </button>
                  </div>
                  <div class="flex-1 overflow-y-auto border border-gray-200 rounded-xl bg-gray-50/50">
                    <div v-for="server in filteredAvailableServers" :key="server.id"
                      @click="toggleLeftSelection(server.id)"
                      class="flex items-center gap-3 p-3 hover:bg-white cursor-pointer border-b border-gray-100 last:border-b-0 transition-all duration-200"
                      :class="{ 'bg-primary-50 shadow-sm': leftSelected.includes(server.id) }">
                      <div class="relative">
                        <input type="checkbox" :checked="leftSelected.includes(server.id)"
                          class="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500">
                        <div v-if="leftSelected.includes(server.id)" class="absolute -inset-1 bg-primary-200 rounded animate-pulse opacity-50"></div>
                      </div>
                      <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-2">
                          <p class="text-sm font-medium text-gray-900 truncate">{{ server.name }}</p>
                          <span v-if="server.state === 1" class="w-1.5 h-1.5 bg-green-500 rounded-full flex-shrink-0"></span>
                          <span v-else class="w-1.5 h-1.5 bg-gray-300 rounded-full flex-shrink-0"></span>
                        </div>
                        <p class="text-xs text-gray-400 truncate mt-0.5">{{ server.description || '暂无描述' }}</p>
                      </div>
                      <div v-if="server.type" class="text-xs px-2 py-0.5 bg-gray-100 text-gray-500 rounded truncate max-w-[80px]">
                        {{ server.type }}
                      </div>
                    </div>
                    <div v-if="filteredAvailableServers.length === 0" class="p-8 text-center">
                      <div class="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-2">
                        <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                        </svg>
                      </div>
                      <p class="text-gray-400 text-sm">暂无可绑定的 Server</p>
                    </div>
                  </div>
                </div>

                <div class="flex flex-col justify-center gap-3">
                  <button @click="handleBind"
                    :disabled="leftSelected.length === 0"
                    class="px-4 py-2.5 bg-gradient-to-r from-primary-500 to-primary-600 text-white text-sm font-medium rounded-lg hover:from-primary-600 hover:to-primary-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-2 shadow-sm transition-all duration-200 hover:shadow-md"
                    :class="{ 'transform scale-105': leftSelected.length > 0 }">
                    <span v-if="leftSelected.length > 0" class="bg-white/20 px-1.5 py-0.5 rounded text-xs">{{ leftSelected.length }}</span>
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"/>
                    </svg>
                    <span>绑定</span>
                  </button>
                  <button @click="handleUnbind"
                    :disabled="rightSelected.length === 0"
                    class="px-4 py-2.5 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-red-50 hover:border-red-200 hover:text-red-600 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-2 transition-all duration-200"
                    :class="{ 'border-red-300 bg-red-50 text-red-600': rightSelected.length > 0 }">
                    <span v-if="rightSelected.length > 0" class="bg-red-100 text-red-600 px-1.5 py-0.5 rounded text-xs">{{ rightSelected.length }}</span>
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
                    </svg>
                    <span>解除</span>
                  </button>
                </div>

                <div class="flex-1 flex flex-col">
                  <div class="mb-3 flex items-center justify-between">
                    <h4 class="text-sm font-semibold text-gray-700 flex items-center gap-2">
                      <span class="w-2 h-2 bg-primary-500 rounded-full"></span>
                      已绑定 Server
                      <span class="text-xs font-normal text-gray-400">({{ filteredBoundServers.length }})</span>
                    </h4>
                  </div>
                  <div class="relative mb-3">
                    <input v-model="rightSearch" type="text" placeholder="搜索名称或描述..."
                      class="w-full pl-9 pr-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-primary-100 focus:border-primary-400 bg-gray-50">
                    <svg class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                    </svg>
                    <button v-if="rightSearch" @click="rightSearch = ''" class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                      </svg>
                    </button>
                  </div>
                  <div class="flex-1 overflow-y-auto border border-gray-200 rounded-xl bg-gray-50/50">
                    <div v-for="binding in filteredBoundServers" :key="binding.id"
                      @click="toggleRightSelection(binding.id)"
                      class="flex items-center gap-3 p-3 hover:bg-white cursor-pointer border-b border-gray-100 last:border-b-0 transition-all duration-200"
                      :class="{ 'bg-primary-50 shadow-sm': rightSelected.includes(binding.id) }">
                      <div class="relative">
                        <input type="checkbox" :checked="rightSelected.includes(binding.id)"
                          class="w-4 h-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500">
                        <div v-if="rightSelected.includes(binding.id)" class="absolute -inset-1 bg-primary-200 rounded animate-pulse opacity-50"></div>
                      </div>
                      <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-2">
                          <p class="text-sm font-medium text-gray-900 truncate">{{ binding.server?.name || 'Unknown' }}</p>
                          <span v-if="binding.server?.enabled" class="w-1.5 h-1.5 bg-green-500 rounded-full flex-shrink-0"></span>
                          <span v-else class="w-1.5 h-1.5 bg-gray-300 rounded-full flex-shrink-0"></span>
                        </div>
                        <p class="text-xs text-gray-400 truncate mt-0.5">{{ binding.server?.description || '暂无描述' }}</p>
                      </div>
                      <div v-if="binding.server?.type" class="text-xs px-2 py-0.5 bg-gray-100 text-gray-500 rounded truncate max-w-[80px]">
                        {{ binding.server.type }}
                      </div>
                    </div>
                    <div v-if="filteredBoundServers.length === 0" class="p-8 text-center">
                      <div class="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-2">
                        <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                        </svg>
                      </div>
                      <p class="text-gray-400 text-sm">暂无绑定</p>
                    </div>
                  </div>
                </div>
              </div>
              <div v-if="selectedTool" class="mt-4 pt-4 border-t border-gray-100 flex items-center justify-between text-sm text-gray-500">
                <div class="flex items-center gap-4">
                  <span>已绑定: <strong class="text-primary-600">{{ filteredBoundServers.length }}</strong> 个 Server</span>
                  <span class="text-gray-300">|</span>
                  <span>可绑定: <strong class="text-green-600">{{ filteredAvailableServers.length }}</strong> 个 Server</span>
                </div>
                <div class="flex items-center gap-2">
                  <button @click="loadData" class="text-gray-400 hover:text-primary-600 transition-colors flex items-center gap-1">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                    </svg>
                    刷新
                  </button>
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
import { ref, computed, watch } from 'vue'
import { useToolBindingsStore } from '@/stores/toolBindings'
import { useMCPServersStore } from '@/stores/mcpServers'

const props = defineProps({
  visible: { type: Boolean, default: false },
  selectedTool: { type: Object, default: null }
})

const emit = defineEmits(['close'])

const toolBindingsStore = useToolBindingsStore()
const mcpServersStore = useMCPServersStore()

const bindings = ref([])
const leftSearch = ref('')
const rightSearch = ref('')
const leftSelected = ref([])
const rightSelected = ref([])

const availableServers = computed(() => {
  const boundServerIds = bindings.value.map(b => b.server_id)
  return mcpServersStore.servers.filter(s => !boundServerIds.includes(s.id))
})

const filteredAvailableServers = computed(() => {
  if (!leftSearch.value) return availableServers.value
  const q = leftSearch.value.toLowerCase()
  return availableServers.value.filter(s =>
    s.name.toLowerCase().includes(q) || (s.description && s.description.toLowerCase().includes(q))
  )
})

const bindingsWithServer = computed(() => {
  return bindings.value.map(binding => ({
    ...binding,
    server: mcpServersStore.servers.find(s => s.id === binding.server_id)
  })).filter(b => b.server)
})

const filteredBoundServers = computed(() => {
  if (!rightSearch.value) return bindingsWithServer.value
  const q = rightSearch.value.toLowerCase()
  return bindingsWithServer.value.filter(b =>
    b.server?.name.toLowerCase().includes(q) || (b.server?.description && b.server?.description.toLowerCase().includes(q))
  )
})

watch(() => props.visible, async (newVal) => {
  if (newVal && props.selectedTool) {
    await loadData()
    leftSelected.value = []
    rightSelected.value = []
  }
})

const loadData = async () => {
  await mcpServersStore.fetchServers()
  bindings.value = await toolBindingsStore.getToolBindings(props.selectedTool.id)
}

const toggleLeftSelection = (id) => {
  const idx = leftSelected.value.indexOf(id)
  if (idx === -1) leftSelected.value.push(id)
  else leftSelected.value.splice(idx, 1)
}

const toggleRightSelection = (id) => {
  const idx = rightSelected.value.indexOf(id)
  if (idx === -1) rightSelected.value.push(id)
  else rightSelected.value.splice(idx, 1)
}

const handleBind = async () => {
  if (!leftSelected.value.length) return
  for (const serverId of leftSelected.value) {
    await toolBindingsStore.bindTool(props.selectedTool.id, serverId)
  }
  await loadData()
  leftSelected.value = []
}

const handleUnbind = async () => {
  if (!rightSelected.value.length) return
  const bindingsToRemove = bindingsWithServer.value.filter(b => rightSelected.value.includes(b.id))
  for (const binding of bindingsToRemove) {
    await toolBindingsStore.unbindTool(props.selectedTool.id, binding.server_id)
  }
  await loadData()
  rightSelected.value = []
}
</script>
