<template>
  <teleport to="body">
    <transition name="fade">
      <div v-if="visible" class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 bg-black bg-opacity-50" @click="handleCancel"></div>
          <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md p-6 fade-in">
            <div class="flex items-center mb-4">
              <div class="flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center mr-4"
                :class="type === 'danger' ? 'bg-red-100' : 'bg-blue-100'">
                <svg v-if="type === 'danger'" class="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
                </svg>
                <svg v-else class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
              </div>
              <h3 class="text-lg font-semibold text-gray-900">{{ title }}</h3>
            </div>
            <p class="text-gray-600 mb-6">{{ message }}</p>
            <div class="flex justify-end space-x-3">
              <button
                @click="handleCancel"
                class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50"
              >
                {{ cancelText }}
              </button>
              <button
                @click="handleConfirm"
                class="px-4 py-2 text-white text-sm font-medium rounded-lg"
                :class="type === 'danger' ? 'bg-red-600 hover:bg-red-700' : 'bg-primary-600 hover:bg-primary-700'"
              >
                {{ confirmText }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup>
import { ref } from 'vue'

const visible = ref(false)
const title = ref('确认')
const message = ref('')
const type = ref('info')
const confirmText = ref('确定')
const cancelText = ref('取消')
let resolvePromise = null

const show = (options) => {
  return new Promise((resolve) => {
    title.value = options.title || '确认'
    message.value = options.message
    type.value = options.type || 'info'
    confirmText.value = options.confirmText || '确定'
    cancelText.value = options.cancelText || '取消'
    resolvePromise = resolve
    visible.value = true
  })
}

const handleConfirm = () => {
  visible.value = false
  if (resolvePromise) resolvePromise(true)
}

const handleCancel = () => {
  visible.value = false
  if (resolvePromise) resolvePromise(false)
}

defineExpose({ show })
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
</style>
