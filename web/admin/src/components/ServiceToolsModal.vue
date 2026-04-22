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
import { servicesApi } from '@/api/services'

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
    const res = await servicesApi.getServiceTools(props.service.id)
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