<template>
  <teleport to="body">
    <transition name="fade">
      <div v-if="visible" class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 bg-black bg-opacity-50" @click="$emit('close')"></div>
          <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-4xl max-h-[90vh] overflow-hidden fade-in">
            <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <div class="flex items-center gap-3">
                <h3 class="text-lg font-semibold text-gray-900">Server 工具管理</h3>
                <span v-if="selectedServer" class="px-2.5 py-1 text-xs font-medium bg-primary-100 text-primary-700 rounded-full">
                  {{ selectedServer.name }}
                </span>
              </div>
              <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <div class="p-6">
              <div v-if="!selectedServer" class="text-center py-12">
                <p class="text-gray-500">请选择一个 Server</p>
              </div>
              <div v-else class="flex gap-6 h-[60vh]">
                <div class="flex-1 flex flex-col">
                  <div class="mb-3">
                    <h4 class="text-sm font-semibold text-gray-700 mb-2">可绑定工具</h4>
                    <div class="relative">
                      <input v-model="leftSearch" type="text" placeholder="搜索..."
                        class="w-full pl-9 pr-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-primary-100 focus:border-primary-400">
                      <svg class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                      </svg>
                    </div>
                  </div>
                  <div class="flex-1 overflow-y-auto border border-gray-200 rounded-xl bg-gray-50">
                    <div v-for="tool in filteredAvailableTools" :key="tool.id"
                      @click="toggleLeftSelection(tool.id)"
                      class="flex items-center gap-3 p-3 hover:bg-white cursor-pointer border-b border-gray-100 last:border-b-0"
                      :class="{ 'bg-primary-50': leftSelected.includes(tool.id) }">
                      <input type="checkbox" :checked="leftSelected.includes(tool.id)"
                        class="rounded border-gray-300 text-primary-600">
                      <div class="flex-1 min-w-0">
                        <p class="text-sm font-medium text-gray-900 truncate">{{ tool.name }}</p>
                        <p class="text-xs text-gray-500 truncate">{{ tool.description || '无描述' }}</p>
                      </div>
                    </div>
                    <div v-if="filteredAvailableTools.length === 0" class="p-8 text-center text-gray-400 text-sm">
                      暂无可绑定的工具
                    </div>
                  </div>
                </div>

                <div class="flex flex-col justify-center gap-2">
                  <button @click="handleBind"
                    :disabled="leftSelected.length === 0"
                    class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
                    <span>绑定</span>
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3"/>
                    </svg>
                  </button>
                  <button @click="handleUnbind"
                    :disabled="rightSelected.length === 0"
                    class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"/>
                    </svg>
                    <span>解除</span>
                  </button>
                </div>

                <div class="flex-1 flex flex-col">
                  <div class="mb-3">
                    <h4 class="text-sm font-semibold text-gray-700 mb-2">已绑定工具</h4>
                    <div class="relative">
                      <input v-model="rightSearch" type="text" placeholder="搜索..."
                        class="w-full pl-9 pr-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-primary-100 focus:border-primary-400">
                      <svg class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
                      </svg>
                    </div>
                  </div>
                  <div class="flex-1 overflow-y-auto border border-gray-200 rounded-xl bg-gray-50">
                    <div v-for="binding in filteredBoundTools" :key="binding.id"
                      @click="toggleRightSelection(binding.id)"
                      class="flex items-center gap-3 p-3 hover:bg-white cursor-pointer border-b border-gray-100 last:border-b-0"
                      :class="{ 'bg-primary-50': rightSelected.includes(binding.id) }">
                      <input type="checkbox" :checked="rightSelected.includes(binding.id)"
                        class="rounded border-gray-300 text-primary-600">
                      <div class="flex-1 min-w-0">
                        <p class="text-sm font-medium text-gray-900 truncate">{{ binding.tool?.name || 'Unknown' }}</p>
                        <p class="text-xs text-gray-500 truncate">{{ binding.tool?.description || '无描述' }}</p>
                      </div>
                    </div>
                    <div v-if="filteredBoundTools.length === 0" class="p-8 text-center text-gray-400 text-sm">
                      暂无绑定
                    </div>
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
import { ref, computed, watch } from 'vue'
import { useToolBindingsStore } from '@/stores/toolBindings'
import { useToolsStore } from '@/stores/tools'

const props = defineProps({
  visible: { type: Boolean, default: false },
  selectedServer: { type: Object, default: null }
})

const emit = defineEmits(['close'])

const toolBindingsStore = useToolBindingsStore()
const toolsStore = useToolsStore()

const bindings = ref([])
const leftSearch = ref('')
const rightSearch = ref('')
const leftSelected = ref([])
const rightSelected = ref([])

const availableTools = computed(() => {
  const boundToolIds = bindings.value.map(b => b.tool_id)
  return toolsStore.tools.filter(t => !boundToolIds.includes(t.id))
})

const filteredAvailableTools = computed(() => {
  if (!leftSearch.value) return availableTools.value
  const q = leftSearch.value.toLowerCase()
  return availableTools.value.filter(t =>
    t.name.toLowerCase().includes(q) || (t.description && t.description.toLowerCase().includes(q))
  )
})

const bindingsWithTool = computed(() => {
  return bindings.value.map(binding => ({
    ...binding,
    tool: toolsStore.tools.find(t => t.id === binding.tool_id)
  })).filter(b => b.tool)
})

const filteredBoundTools = computed(() => {
  if (!rightSearch.value) return bindingsWithTool.value
  const q = rightSearch.value.toLowerCase()
  return bindingsWithTool.value.filter(b =>
    b.tool?.name.toLowerCase().includes(q) || (b.tool?.description && b.tool?.description.toLowerCase().includes(q))
  )
})

watch(() => props.visible, async (newVal) => {
  if (newVal && props.selectedServer) {
    await loadData()
    leftSelected.value = []
    rightSelected.value = []
  }
})

const loadData = async () => {
  await toolsStore.fetchToolsForAdmin()
  bindings.value = await toolBindingsStore.getServerBindings(props.selectedServer.id)
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
  for (const toolId of leftSelected.value) {
    await toolBindingsStore.bindTool(toolId, props.selectedServer.id)
  }
  await loadData()
  leftSelected.value = []
}

const handleUnbind = async () => {
  if (!rightSelected.value.length) return
  const bindingsToRemove = bindingsWithTool.value.filter(b => rightSelected.value.includes(b.id))
  for (const binding of bindingsToRemove) {
    await toolBindingsStore.unbindTool(binding.tool_id, props.selectedServer.id)
  }
  await loadData()
  rightSelected.value = []
}
</script>