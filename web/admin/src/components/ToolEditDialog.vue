<template>
  <teleport to="body">
    <transition name="fade">
      <div v-if="visible" class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 bg-black bg-opacity-50" @click="$emit('close')"></div>
          <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-hidden fade-in">
            <!-- Header -->
            <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h3 class="text-lg font-semibold text-gray-900">{{ editingTool ? '编辑工具' : '创建工具' }}</h3>
              <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>

            <!-- Content -->
            <div class="p-6 overflow-y-auto max-h-[calc(90vh-130px)]">
              <!-- Draft notice -->
              <div v-if="hasDraft && !editingTool" class="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg flex items-center justify-between">
                <div class="flex items-center text-sm text-blue-700">
                  <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                  </svg>
                  检测到未提交的草稿
                </div>
                <button @click="clearDraft" class="text-xs text-blue-500 hover:text-blue-700">清除草稿</button>
              </div>

              <form @submit.prevent="handleSubmit" class="space-y-5">
                <!-- Basic Info -->
                <div class="space-y-4">
                  <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                    <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                    </svg>
                    基本信息
                  </h4>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">工具名称 *</label>
                    <input
                      v-model="form.name"
                      type="text"
                      required
                      class="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors"
                      :class="{
                        'border-gray-300': toolNameValidation.valid || !form.name,
                        'border-red-400 bg-red-50': !toolNameValidation.valid && form.name
                      }"
                      placeholder="唯一标识的工具名称"
                      :disabled="!!editingTool"
                      @blur="toolNameTouched = true"
                    >
                    <p v-if="!toolNameValidation.valid && form.name" class="mt-1 text-xs text-red-500 flex items-center">
                      <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                      </svg>
                      {{ toolNameValidation.message }}
                    </p>
                    <p v-else class="mt-1 text-xs text-gray-400">
                      1-64个字符，只允许字母、数字、下划线(_)、连字符(-)、点(.)
                    </p>
                  </div>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">描述</label>
                    <textarea v-model="form.description" rows="2"
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      placeholder="工具功能描述"></textarea>
                  </div>
                </div>

                <!-- Server & Service Selection -->
                <div class="space-y-4 pt-4 border-t border-gray-100">
                  <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                    <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
                    </svg>
                    关联配置
                  </h4>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">HTTP 服务 *</label>
                    <select v-model="form.service_id" required
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      @change="onServiceChange">
                      <option :value="0">选择服务</option>
                      <option v-for="service in services" :key="service.id" :value="service.id">
                        {{ service.name }}
                      </option>
                    </select>
                  </div>
                </div>

                <!-- Input Parameters -->
                <div class="space-y-4 pt-4 border-t border-gray-100">
                  <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                    <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                    </svg>
                    入参定义
                  </h4>
                  <p class="text-xs text-gray-500">入参从 HTTP 服务的 InputSchema 自动同步</p>

                  <!-- Empty state -->
                  <div v-if="inputParams.length === 0" class="text-center py-6 bg-gray-50 rounded-xl border border-gray-200">
                    <svg class="w-8 h-8 text-gray-300 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                    </svg>
                    <p class="text-sm text-gray-500">暂无入参定义，请先选择 HTTP 服务</p>
                  </div>

                  <!-- Parameter rows -->
                  <div v-else class="space-y-2">
                    <div v-for="(param, index) in inputParams" :key="index"
                      :class="[
                        'flex items-center gap-2 rounded-xl p-3 border',
                        hasInputMapping(param.name)
                          ? 'bg-blue-50/50 border-blue-200'
                          : 'bg-white border-gray-200 hover:border-gray-300'
                      ]">

                      <!-- Parameter name input -->
                      <div class="flex-1">
                        <input v-model="param.name" placeholder="参数名"
                          class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all">
                      </div>

                      <!-- Colon separator -->
                      <div class="text-gray-400 text-sm">:</div>

                      <!-- HTTP Schema Mapping Field - Highlighted -->
                      <div class="flex-shrink-0">
                        <span v-if="getInputMapping(param.name)"
                          :class="[
                            'inline-flex items-center gap-1 px-2.5 py-1 text-xs font-medium rounded-md',
                            'bg-green-100 text-green-700 border border-green-200'
                          ]">
                          <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                            <path d="M13 10V3L4 14h7v7l9-11h-7z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                          </svg>
                          <span class="font-mono">{{ getInputMapping(param.name).target }}</span>
                        </span>
                        <span v-else :class="[
                          'inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-md',
                          'bg-gray-100 text-gray-400 border border-gray-200'
                        ]">
                          未映射
                        </span>
                      </div>

                      <!-- Type tag -->
                      <div class="flex-shrink-0">
                        <span :class="[
                          'inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-md',
                          param.type === 'string' ? 'bg-blue-100 text-blue-700' : '',
                          (param.type === 'integer' || param.type === 'number') ? 'bg-emerald-100 text-emerald-700' : '',
                          param.type === 'boolean' ? 'bg-amber-100 text-amber-700' : '',
                          (param.type === 'object' || param.type === 'array') ? 'bg-purple-100 text-purple-700' : '',
                          !param.type ? 'bg-gray-100 text-gray-600' : ''
                        ]">
                          {{ param.type || 'unknown' }}
                        </span>
                      </div>

                      <!-- Description input -->
                      <div class="flex-1">
                        <input v-model="param.description" placeholder="参数描述"
                          class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm transition-colors focus:ring-2 focus:ring-blue-100 focus:border-blue-400">
                      </div>

                      <!-- Required indicator -->
                      <div class="flex-shrink-0">
                        <span v-if="param.schema_required"
                          :class="[
                            'inline-flex items-center px-2 py-1 text-xs font-medium rounded-md',
                            'bg-red-50 text-red-500 border border-red-100'
                          ]">
                          <svg class="w-3 h-3 mr-1" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                            <path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                          </svg>
                          必填
                        </span>
                        <span v-else-if="param.required"
                          :class="[
                            'inline-flex items-center px-2 py-1 text-xs font-medium rounded-md',
                            'bg-red-50 text-red-600 border border-red-100 cursor-pointer hover:bg-red-100'
                          ]"
                          @click="param.required = false">
                          <svg class="w-3 h-3 mr-1" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                            <path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                          </svg>
                          必填
                        </span>
                        <label v-else :class="[
                          'flex items-center gap-1 text-xs text-gray-500 cursor-pointer hover:text-blue-600'
                        ]">
                          <input v-model="param.required" type="checkbox" :class="[
                            'rounded border-gray-300 text-blue-600 w-3.5 h-3.5'
                          ]">
                          可选
                        </label>
                      </div>

                      <!-- Delete button -->
                      <button
                        v-if="param.schema_required"
                        type="button"
                        disabled
                        class="p-1.5 text-gray-300 rounded-lg cursor-not-allowed">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                        </svg>
                      </button>
                      <button
                        v-else
                        type="button"
                        @click="removeInputParam(index)"
                        class="p-1.5 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>

                <!-- Deleted Fields -->
                <div v-if="deletedInputFields.length > 0" class="pt-4 border-t border-gray-100">
                  <div class="flex items-center gap-2 mb-3">
                    <svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                    </svg>
                    <span class="text-sm text-gray-700">已删除字段（点击重新添加）</span>
                  </div>
                  <div class="flex flex-wrap gap-2">
                    <div
                      v-for="field in deletedInputFields.filter(f => !hasInputMapping(f.name))"
                      :key="field.original_name || field.name"
                      class="flex items-center gap-2 px-3 py-1.5 bg-white border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 transition-all cursor-pointer"
                      @click="restoreInputField(field)">
                      <code class="text-sm text-gray-700 font-mono">{{ field.name }}</code>
                      <span :class="[
                        'px-2 py-0.5 text-xs rounded-md',
                        field.type === 'string' ? 'bg-blue-100 text-blue-700' : '',
                        (field.type === 'integer' || field.type === 'number') ? 'bg-emerald-100 text-emerald-700' : '',
                        field.type === 'boolean' ? 'bg-amber-100 text-amber-700' : '',
                        (field.type === 'object' || field.type === 'array') ? 'bg-purple-100 text-purple-700' : '',
                        !field.type ? 'bg-gray-100 text-gray-600' : ''
                      ]">{{ field.type }}</span>
                      <span class="text-blue-500 text-sm font-medium">+</span>
                    </div>
                  </div>
                </div>

                <!-- Output Mapping -->
                <div class="space-y-4 pt-4 border-t border-gray-100">
                  <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                    <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/>
                    </svg>
                    出参映射
                  </h4>
                  <p class="text-xs text-gray-500">选择 HTTP 服务出参字段，自动生成 1:1 映射</p>

                  <!-- Quick add fields -->
                  <div v-if="flatFieldList.length > 0" class="bg-blue-50 rounded-xl p-4">
                    <label class="block text-sm font-medium text-gray-700 mb-3 flex items-center">
                      <svg class="w-4 h-4 mr-1.5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
                      </svg>
                      快速添加（点击字段即添加映射）
                    </label>
                    <div class="flex flex-wrap gap-2">
                      <button
                        v-for="node in flatFieldList"
                        :key="node.path"
                        type="button"
                        @click="addOutputMappingWithField(node.path, node.type)"
                        class="flex items-center gap-2 px-3 py-2 bg-white border border-gray-200 rounded-lg hover:border-blue-300 hover:bg-blue-50 transition-all shadow-sm"
                      >
                        <span class="text-sm font-mono text-gray-700">{{ node.path }}</span>
                        <span :class="[
                          'px-2 py-0.5 text-xs rounded-md',
                          node.type === 'string' ? 'bg-blue-100 text-blue-700' : '',
                          (node.type === 'integer' || node.type === 'number') ? 'bg-emerald-100 text-emerald-700' : '',
                          node.type === 'boolean' ? 'bg-amber-100 text-amber-700' : '',
                          (node.type === 'object' || node.type === 'array') ? 'bg-purple-100 text-purple-700' : '',
                          !node.type ? 'bg-gray-100 text-gray-600' : ''
                        ]">
                          {{ node.type }}
                        </span>
                      </button>
                    </div>
                  </div>

                  <!-- Empty state -->
                  <div v-if="outputMappings.length === 0" class="text-center py-6 bg-gray-50 rounded-xl">
                    <p class="text-sm text-gray-500">暂无出参映射，点击上方字段添加</p>
                  </div>

                  <!-- Mapping rows -->
                  <div v-else class="space-y-2">
                    <div v-for="(mapping, index) in outputMappings" :key="index"
                      class="flex items-center gap-3 bg-white border border-gray-200 p-3 rounded-xl">

                      <!-- Target field input -->
                      <input v-model="mapping.target_field" placeholder="目标字段"
                        class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all">

                      <!-- Arrow -->
                      <span class="text-gray-400 text-lg flex-shrink-0">→</span>

                      <!-- Source field select with type -->
                      <div class="flex items-center gap-2 flex-1">
                        <select v-model="mapping.source_field"
                          class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm font-mono focus:ring-2 focus:ring-blue-100 focus:border-blue-400 transition-all bg-white">
                          <option value="">选择源字段</option>
                          <option v-for="node in flatFieldList" :key="node.path" :value="node.path">
                            {{ node.path }}
                          </option>
                        </select>
                        <span v-if="getMappingTypeLabel(mapping)" :class="[
                          'px-2 py-1 text-xs rounded-md flex-shrink-0',
                          'bg-blue-100 text-blue-700 border border-blue-200'
                        ]">
                          {{ getMappingTypeLabel(mapping) }}
                        </span>
                      </div>

                      <!-- Delete button -->
                      <button type="button" @click="removeOutputMapping(index)"
                        class="p-1.5 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors flex-shrink-0">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>

                <!-- Actions -->
                <div class="flex justify-end gap-3 pt-4 border-t border-gray-200">
                  <button type="button" @click="$emit('close')"
                    class="px-5 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50 transition-colors">
                    取消
                  </button>
                  <button type="submit"
                    class="px-5 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors shadow-sm">
                    {{ editingTool ? '保存' : '创建' }}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup>
import { ref, reactive, watch, computed } from 'vue'
import { useServicesStore } from '@/stores/services'
import { servicesApi } from '@/api/services'
import { extractSchemaFields, schemaToTree, getNodeByPath } from '@/utils/schemaHelper'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  editingTool: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close', 'saved'])

const servicesStore = useServicesStore()

const services = computed(() => servicesStore.services)

const form = reactive({
  name: '',
  description: '',
  service_id: 0
})

const inputParams = ref([])
const inputMappings = ref([])
const inputSchemaFields = ref([])
const outputMappings = ref([])
const outputSchemaFields = ref([])
const outputSchemaTree = ref([])
const deletedInputFields = ref([])
const toolNameTouched = ref(false)

const DRAFT_CACHE_KEY = 'tool_edit_draft'
const hasDraft = ref(false)

const TOOL_NAME_PATTERN = /^[a-zA-Z0-9_.-]{1,64}$/
const TOOL_NAME_MAX_LENGTH = 64

const validateToolName = (name) => {
  if (!name || name.trim() === '') {
    return { valid: false, message: '工具名称不能为空' }
  }
  if (name.length > TOOL_NAME_MAX_LENGTH) {
    return { valid: false, message: '工具名称不能超过 ' + TOOL_NAME_MAX_LENGTH + ' 个字符' }
  }
  if (!TOOL_NAME_PATTERN.test(name)) {
    return { valid: false, message: '只能包含字母、数字、下划线(_)、连字符(-)、点(.)' }
  }
  return { valid: true, message: '' }
}

const saveDraft = () => {
  const hasContent = form.name || form.description || form.service_id ||
    inputParams.value.length > 0 || inputMappings.value.length > 0 || outputMappings.value.length > 0
  if (!hasContent) return

  const draft = {
    form: { ...form },
    inputParams: [...inputParams.value],
    inputMappings: [...inputMappings.value],
    outputMappings: [...outputMappings.value],
    deletedInputFields: [...deletedInputFields.value],
    timestamp: Date.now()
  }
  localStorage.setItem(DRAFT_CACHE_KEY, JSON.stringify(draft))
  hasDraft.value = true
}

const loadDraft = () => {
  const cached = localStorage.getItem(DRAFT_CACHE_KEY)
  if (cached) {
    try {
      const draft = JSON.parse(cached)
      if (Date.now() - draft.timestamp < 24 * 60 * 60 * 1000) {
        Object.assign(form, draft.form)
        inputParams.value = draft.inputParams || []
        inputMappings.value = draft.inputMappings || []
        outputMappings.value = draft.outputMappings || []
        deletedInputFields.value = draft.deletedInputFields || []
        hasDraft.value = true
        return true
      }
    } catch (e) {
      console.error('Failed to load draft:', e)
    }
  }
  return false
}

const clearDraft = () => {
  localStorage.removeItem(DRAFT_CACHE_KEY)
  hasDraft.value = false
  resetForm()
}

const toolNameValidation = computed(() => {
  return validateToolName(form.name)
})

const flatFieldList = computed(() => {
  const result = []
  const traverse = (nodes) => {
    for (const node of nodes) {
      if (!node.children || node.children.length === 0) {
        result.push({ path: node.path, type: node.type })
      }
      if (node.children) {
        traverse(node.children)
      }
    }
  }
  traverse(outputSchemaTree.value)
  return result
})

const inputSchemaTree = ref([])

const flatInputFieldList = computed(() => {
  const result = []
  const traverse = (nodes) => {
    for (const node of nodes) {
      if (!node.children || node.children.length === 0) {
        result.push({ path: node.path, type: node.type })
      }
      if (node.children) {
        traverse(node.children)
      }
    }
  }
  traverse(inputSchemaTree.value)
  return result
})

// 判断入参字段是否有映射
const hasInputMapping = (fieldName) => {
  return inputMappings.value.some(m => m.source === fieldName)
}

// 获取入参字段的映射
const getInputMapping = (fieldName) => {
  return inputMappings.value.find(m => m.source === fieldName)
}

const getMappingTypeLabel = (mapping) => {
  if (!mapping.source_field) return ''
  const node = getNodeByPath(outputSchemaTree.value, mapping.source_field)
  return node ? node.type : ''
}

const getInputMappingTypeLabel = (mapping) => {
  if (!mapping.target) return ''
  const node = getNodeByPath(inputSchemaTree.value, mapping.target)
  return node ? node.type : ''
}

const addInputMappingWithField = (targetField, type) => {
  const fieldName = targetField.split('.').pop()
  inputMappings.value.push({
    source: fieldName,
    target: targetField,
    description: ''
  })
}

const removeInputMapping = (index) => {
  inputMappings.value.splice(index, 1)
}

const addOutputMappingWithField = (sourceField, type) => {
  const fieldName = sourceField.split('.').pop()
  outputMappings.value.push({
    source_field: sourceField,
    target_field: fieldName,
    value_type: 'field',
    default_value: ''
  })
}

const resetForm = () => {
  Object.assign(form, {
    name: '',
    description: '',
    service_id: 0
  })
  inputParams.value = []
  inputMappings.value = []
  inputSchemaFields.value = []
  inputSchemaTree.value = []
  outputMappings.value = []
  outputSchemaFields.value = []
  outputSchemaTree.value = []
  deletedInputFields.value = []
}

const initForm = (tool) => {
  if (tool) {
    Object.assign(form, {
      name: tool.name,
      description: tool.description || '',
      service_id: tool.service_id || 0
    })
    inputParams.value = tool.parameters ? [...tool.parameters] : []
    inputMappings.value = tool.input_mapping ? [...tool.input_mapping] : []
    outputMappings.value = tool.output_mapping ? [...tool.output_mapping] : []

    if (tool.service_id) {
      loadServiceSchemas(tool.service_id)
    }
  } else {
    resetForm()
  }
}

// 加载服务的 input_schema 和 output_schema（用于编辑工具时）
const loadServiceSchemas = async (serviceId) => {
  try {
    const response = await servicesApi.getService(serviceId)
    // request.js 的响应拦截器返回 resData: { code, message, data: { service } }
    const service = response.data?.service || response.data || response
    syncInputFromService(service)
    syncOutputFromService(service)
  } catch (error) {
    console.error('加载服务 schema 失败:', error)
  }
}

// 监听入参字段名变化，同步更新入参映射
watch(inputParams, (newParams, oldParams) => {
  if (!oldParams) return
  
  newParams.forEach((param, index) => {
    const oldParam = oldParams[index]
    if (oldParam && oldParam.name !== param.name) {
      // 字段名变化了，更新映射
      const mappingIndex = inputMappings.value.findIndex(m => m.source === oldParam.name)
      if (mappingIndex !== -1) {
        inputMappings.value[mappingIndex].source = param.name
      }
    }
  })
}, { deep: true })

watch(() => props.visible, (newVal) => {
  if (newVal) {
    if (!props.editingTool) {
      initForm(props.editingTool)
      deletedInputFields.value = []
      const restored = loadDraft()
      if (restored && form.service_id) {
        loadServiceSchemas(form.service_id)
      }
      if (!restored) {
        hasDraft.value = false
      }
    }
    servicesStore.fetchServices()
  }
})

watch(() => props.editingTool, (newVal) => {
  initForm(newVal)
})

watch([() => form, inputParams, inputMappings, outputMappings, deletedInputFields], () => {
  if (props.visible && !props.editingTool) {
    const hasContent = form.name || form.description || form.service_id ||
      inputParams.value.length > 0 || inputMappings.value.length > 0 || outputMappings.value.length > 0
    if (hasContent) {
      saveDraft()
    }
  }
}, { deep: true })

const onServiceChange = async () => {
  if (form.service_id) {
    // 拉取完整服务信息
    try {
      const response = await servicesApi.getService(form.service_id)
      // request.js 的响应拦截器返回 resData: { code, message, data: { service } }
      const service = response.data?.service || response.data || response

      // 同步入参
      syncInputFromService(service)

      // 同步出参字段
      syncOutputFromService(service)
    } catch (error) {
      console.error('获取服务信息失败:', error)
    }
  } else {
    // 清空 outputSchemaFields
    outputSchemaFields.value = []
    outputSchemaTree.value = []
  }
}

const syncInputFromService = (service) => {
  if (!service) return
  const schema = service.input_schema

  let parsedSchema = schema
  if (typeof schema === 'string') {
    try {
      parsedSchema = JSON.parse(schema)
    } catch (e) {
      console.error('Failed to parse input_schema:', e)
    }
  }

  if (parsedSchema && parsedSchema.properties) {
    const newParams = Object.entries(parsedSchema.properties).map(([name, prop]) => ({
      name,
      original_name: name, // 保存原始字段名
      type: prop.type || 'string',
      description: prop.description || '',
      required: parsedSchema.required?.includes(name) || false,
      schema_required: parsedSchema.required?.includes(name) || false
    }))

    inputParams.value = newParams

    // 自动为每个字段生成 1:1 映射
    inputMappings.value = newParams.map(p => ({
      source: p.name,
      target: p.original_name, // target 使用原始字段名
      description: ''
    }))

    inputSchemaFields.value = extractSchemaFields(parsedSchema)
    inputSchemaTree.value = schemaToTree(parsedSchema)
  } else {
    inputParams.value = []
    inputMappings.value = []
    inputSchemaFields.value = []
    inputSchemaTree.value = []
  }
}

const syncOutputFromService = (service) => {
  if (!service) return
  const schema = service.output_schema
  
  // 处理可能的字符串格式
  let parsedSchema = schema
  if (typeof schema === 'string') {
    try {
      parsedSchema = JSON.parse(schema)
    } catch (e) {
      console.error('Failed to parse output_schema:', e)
    }
  }
  
  if (parsedSchema && parsedSchema.properties) {
    // 提取所有字段（包括嵌套）
    outputSchemaFields.value = extractSchemaFields(parsedSchema)
    // 转换为树形结构
    outputSchemaTree.value = schemaToTree(parsedSchema)
  } else {
    outputSchemaFields.value = []
    outputSchemaTree.value = []
  }
}

const removeInputParam = (index) => {
  const param = inputParams.value[index]
  if (param.schema_required) {
    return
  }

  // 同时移除对应的映射
  const mappingIndex = inputMappings.value.findIndex(m => m.source === param.name)
  if (mappingIndex !== -1) {
    inputMappings.value.splice(mappingIndex, 1)
  }

  deletedInputFields.value.push({
    name: param.name,
    original_name: param.original_name || param.name,
    type: param.type,
    description: param.description,
    required: false,
    schema_required: false
  })
  inputParams.value.splice(index, 1)
}

const restoreInputField = (field) => {
  // 恢复字段
  const newField = { ...field }
  inputParams.value.push(newField)

  // 恢复映射
  inputMappings.value.push({
    source: field.name,
    target: field.original_name || field.name,
    description: ''
  })

  // 从已删除列表移除
  const index = deletedInputFields.value.findIndex(f => 
    (f.original_name && f.original_name === (field.original_name || field.name)) ||
    f.name === field.name
  )
  if (index !== -1) {
    deletedInputFields.value.splice(index, 1)
  }
}

const restoreAllDeletedFields = () => {
  const fieldsToRestore = [...deletedInputFields.value]
  fieldsToRestore.forEach(field => {
    if (!hasInputMapping(field.name)) {
      restoreInputField(field)
    }
  })
}

const addOutputMapping = () => {
  outputMappings.value.push({
    source_field: '',
    target_field: '',
    value_type: 'field',
    default_value: ''
  })
}

const removeOutputMapping = (index) => {
  outputMappings.value.splice(index, 1)
}

const handleSubmit = async () => {
  if (!toolNameValidation.value.valid) {
    toolNameTouched.value = true
    return
  }

  const payload = {
    name: form.name,
    description: form.description,
    service_id: form.service_id,
    parameters: inputParams.value.filter(p => p.name),
    input_mapping: inputMappings.value.filter(m => m.target),
    output_mapping: outputMappings.value.filter(m => m.target_field)
  }

  emit('saved', payload)
}
</script>
