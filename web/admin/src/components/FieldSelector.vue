<template>
  <div class="field-selector">
    <button
      type="button"
      @click="showPopup = true"
      class="w-full px-4 py-2.5 border border-gray-300 rounded-xl text-left bg-white hover:bg-gray-50 flex items-center justify-between shadow-sm hover:shadow transition-shadow"
    >
      <span v-if="selectedNode" class="flex items-center gap-2">
        <span class="text-gray-900 font-mono text-base">{{ selectedNode.path }}</span>
        <span class="px-2 py-0.5 text-xs rounded-md bg-gray-100 text-gray-500">
          {{ selectedNode.type }}
        </span>
      </span>
      <span v-else class="text-gray-400 text-base">选择源字段</span>
      <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
      </svg>
    </button>

    <!-- 弹框 -->
    <teleport to="body">
      <transition name="fade">
        <div v-if="showPopup" class="fixed inset-0 z-50 overflow-y-auto" @click.self="showPopup = false">
          <div class="flex items-center justify-center min-h-screen px-4">
            <div class="fixed inset-0 bg-black bg-opacity-30" @click="showPopup = false"></div>
            <div class="relative bg-white rounded-2xl shadow-2xl w-full max-w-lg max-h-[70vh] overflow-hidden fade-in">
              <div class="px-6 py-4 border-b border-gray-100 flex items-center justify-between bg-gradient-to-r from-gray-50 to-white">
                <h3 class="text-lg font-semibold text-gray-900">选择源字段</h3>
                <button @click="showPopup = false" class="text-gray-400 hover:text-gray-600 hover:bg-gray-100 p-2 rounded-lg transition-colors">
                  <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <div class="p-6 overflow-y-auto max-h-[calc(70vh-70px)] bg-gray-50/50">
                <SchemaFieldTree
                  :nodes="nodes"
                  :selected-path="modelValue"
                  @select="handleSelect"
                />
              </div>
            </div>
          </div>
        </div>
      </transition>
    </teleport>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import SchemaFieldTree from './SchemaFieldTree.vue'
import { getNodeByPath } from '@/utils/schemaHelper'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  nodes: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue'])

const showPopup = ref(false)

const selectedNode = computed(() => {
  if (!props.modelValue) return null
  return getNodeByPath(props.nodes, props.modelValue)
})

const handleSelect = (path) => {
  emit('update:modelValue', path)
  showPopup.value = false
}
</script>
