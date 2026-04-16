<template>
  <teleport to="body">
    <transition name="fade">
      <div v-if="visible" class="fixed inset-0 z-50 overflow-y-auto">
        <div class="flex items-center justify-center min-h-screen px-4">
          <div class="fixed inset-0 bg-black bg-opacity-50" @click="$emit('close')"></div>
          <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-2xl max-h-[90vh] overflow-hidden fade-in">
            <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h3 class="text-lg font-semibold text-gray-900">{{ editingTool ? '编辑工具' : '创建工具' }}</h3>
              <button @click="$emit('close')" class="text-gray-400 hover:text-gray-600">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <div class="p-6 overflow-y-auto max-h-[calc(90vh-130px)]">
              <form @submit.prevent="handleSubmit" class="space-y-5">
                <!-- Basic Info -->
                <div class="space-y-4">
                  <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                    <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
                    基本信息
                  </h4>
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">工具名称 *</label>
                    <input v-model="form.name" type="text" required
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      placeholder="唯一标识的工具名称"
                      :disabled="!!editingTool">
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
                    <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/></svg>
                    关联配置
                  </h4>
                  <div class="grid grid-cols-2 gap-4">
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">MCP Server *</label>
                      <select v-model="form.vauth_key" required
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                        @change="onServerChange">
                        <option value="">选择 Server</option>
                        <option v-for="server in servers" :key="server.id" :value="server.vauth_key">
                          {{ server.name }}
                        </option>
                      </select>
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">HTTP 服务 *</label>
                      <select v-model="form.service_id" required
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                        @change="onServiceChange">
                        <option value="">选择服务</option>
                        <option v-for="service in services" :key="service.id" :value="service.id">
                          {{ service.name }}
                        </option>
                      </select>
                    </div>
                  </div>
                </div>

                <!-- Input Parameters -->
                <div class="space-y-4 pt-4 border-t border-gray-100">
                  <div class="flex items-center justify-between">
                    <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                      <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
                      入参定义
                    </h4>
                    <button type="button" @click="syncFromService"
                      class="text-xs text-primary-600 hover:text-primary-700 font-medium">
                      从服务同步
                    </button>
                  </div>
                  <p class="text-xs text-gray-500">入参将从 HTTP 服务的 InputSchema 自动同步，也可添加额外字段</p>

                  <div v-if="inputParams.length === 0" class="text-center py-6 bg-gray-50 rounded-lg">
                    <p class="text-sm text-gray-500">暂无入参定义</p>
                    <button type="button" @click="addInputParam" class="mt-2 text-xs text-primary-600 hover:text-primary-700 font-medium">
                      + 添加参数
                    </button>
                  </div>

                  <div v-else class="space-y-3">
                    <div v-for="(param, index) in inputParams" :key="index"
                      class="bg-gray-50 p-3 rounded-lg space-y-2">
                      <div class="flex gap-2 items-start">
                        <input v-model="param.name" placeholder="参数名"
                          class="flex-1 px-2 py-1.5 border border-gray-300 rounded text-sm focus:ring-1 focus:ring-primary-500">
                        <select v-model="param.type"
                          class="px-2 py-1.5 border border-gray-300 rounded text-sm focus:ring-1 focus:ring-primary-500">
                          <option value="string">字符串</option>
                          <option value="integer">整数</option>
                          <option value="number">数字</option>
                          <option value="boolean">布尔</option>
                          <option value="array">数组</option>
                          <option value="object">对象</option>
                        </select>
                        <label class="flex items-center gap-1 text-xs text-gray-500 whitespace-nowrap pt-1.5">
                          <input v-model="param.required" type="checkbox" class="rounded border-gray-300 text-primary-600">
                          必填
                        </label>
                        <button type="button" @click="removeInputParam(index)"
                          class="p-1 text-red-400 hover:text-red-600 rounded">
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                        </button>
                      </div>
                      <input v-model="param.description" placeholder="参数描述"
                        class="w-full px-2 py-1 border border-gray-200 rounded text-xs focus:ring-1 focus:ring-primary-500">
                    </div>
                    <button type="button" @click="addInputParam"
                      class="text-sm text-primary-600 hover:text-primary-700 font-medium">
                      + 添加参数
                    </button>
                  </div>
                </div>

                <!-- Output Mapping -->
                <div class="space-y-4 pt-4 border-t border-gray-100">
                  <div class="flex items-center justify-between">
                    <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                      <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/></svg>
                      出参映射
                    </h4>
                    <div class="flex gap-2">
                      <button type="button" @click="outputMappingMode = 'full'"
                        :class="outputMappingMode === 'full' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
                        class="px-3 py-1 text-xs font-medium rounded-lg transition-colors">
                        完整映射
                      </button>
                      <button type="button" @click="outputMappingMode = 'custom'"
                        :class="outputMappingMode === 'custom' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
                        class="px-3 py-1 text-xs font-medium rounded-lg transition-colors">
                        自定义字段
                      </button>
                    </div>
                  </div>
                  
                  <!-- 完整映射模式 -->
                  <div v-if="outputMappingMode === 'full'">
                    <p class="text-xs text-gray-500">完整映射 HTTP 服务的 OutputSchema 字段，字段不可修改</p>
                    
                    <div v-if="outputMappings.length === 0" class="text-center py-6 bg-gray-50 rounded-lg">
                      <p class="text-sm text-gray-500">暂无出参映射</p>
                      <button type="button" @click="handleSyncOutputFromService" class="mt-2 text-xs text-primary-600 hover:text-primary-700 font-medium">
                        从服务同步
                      </button>
                    </div>
                    
                    <div v-else class="space-y-3">
                      <div v-for="(mapping, index) in outputMappings" :key="index"
                        class="bg-gray-50 p-3 rounded-lg space-y-2">
                        <div class="flex gap-2 items-start">
                          <div class="flex-1">
                            <label class="block text-xs text-gray-500 mb-1">源字段</label>
                            <select v-model="mapping.source_field" disabled
                              class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm bg-gray-100 text-gray-500">
                              <option value="">选择源字段</option>
                              <option v-for="field in outputSchemaFields" :key="field" :value="field">
                                {{ field }}
                              </option>
                            </select>
                          </div>
                          <div class="flex-1">
                            <label class="block text-xs text-gray-500 mb-1">目标字段</label>
                            <input v-model="mapping.target_field" placeholder="目标字段名" disabled
                              class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm bg-gray-100 text-gray-500">
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                  
                  <!-- 自定义字段模式 -->
                  <div v-else>
                    <p class="text-xs text-gray-500">自定义出参字段，值来源可选择：引用 HTTP 服务字段 或 默认值</p>
                    
                    <div v-if="outputMappings.length === 0" class="text-center py-6 bg-gray-50 rounded-lg">
                      <p class="text-sm text-gray-500">暂无出参映射</p>
                      <button type="button" @click="addOutputMapping" class="mt-2 text-xs text-primary-600 hover:text-primary-700 font-medium">
                        + 添加字段
                      </button>
                    </div>
                    
                    <div v-else class="space-y-3">
                      <div v-for="(mapping, index) in outputMappings" :key="index"
                        class="bg-gray-50 p-3 rounded-lg space-y-2">
                        <div class="flex gap-2 items-start">
                          <div class="flex-1">
                            <label class="block text-xs text-gray-500 mb-1">目标字段名 *</label>
                            <input v-model="mapping.target_field" placeholder="目标字段名"
                              class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:ring-1 focus:ring-primary-500">
                          </div>
                          <button type="button" @click="removeOutputMapping(index)"
                            class="p-1 text-red-400 hover:text-red-600 rounded mt-5">
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                          </button>
                        </div>
                        
                        <div class="flex gap-2 items-start">
                          <div class="flex-1">
                            <label class="block text-xs text-gray-500 mb-1">值来源</label>
                            <select v-model="mapping.value_type" @change="onMappingValueTypeChange(mapping)"
                              class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:ring-1 focus:ring-primary-500">
                              <option value="field">引用 HTTP 服务字段</option>
                              <option value="default">默认值</option>
                            </select>
                          </div>
                        </div>
                        
                        <!-- 引用字段模式 -->
                        <div v-if="mapping.value_type === 'field'" class="flex gap-2 items-start">
                          <div class="flex-1">
                            <label class="block text-xs text-gray-500 mb-1">选择源字段</label>
                            <select v-model="mapping.source_field"
                              class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:ring-1 focus:ring-primary-500">
                              <option value="">选择源字段</option>
                              <option v-for="field in outputSchemaFields" :key="field" :value="field">
                                {{ field }}
                              </option>
                            </select>
                          </div>
                        </div>
                        
                        <!-- 默认值模式 -->
                        <div v-else class="flex gap-2 items-start">
                          <div class="flex-1">
                            <label class="block text-xs text-gray-500 mb-1">默认值</label>
                            <input v-model="mapping.default_value" placeholder="输入默认值"
                              class="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:ring-1 focus:ring-primary-500">
                          </div>
                        </div>
                      </div>
                      
                      <button type="button" @click="addOutputMapping"
                        class="text-sm text-primary-600 hover:text-primary-700 font-medium">
                        + 添加字段
                      </button>
                    </div>
                  </div>
                  
                  <!-- 可视化编辑模式 -->
                  <div v-else>
                    <p class="text-xs text-gray-500">可视化编辑出参 schema 结构</p>
                    
                    <div class="space-y-2">
                      <SchemaFieldNode
                        v-for="field in outputFields"
                        :key="field.id"
                        :field="field"
                        :depth="0"
                        @update="handleOutputFieldUpdate"
                        @delete="handleOutputFieldDelete"
                      />
                      <button type="button" @click="addOutputField"
                        class="text-sm text-primary-600 hover:text-primary-700 font-medium">
                        + 添加字段
                      </button>
                    </div>
                  </div>
                </div>

                <!-- Actions -->
                <div class="flex justify-end space-x-3 pt-4">
                  <button type="button" @click="$emit('close')"
                    class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50">
                    取消
                  </button>
                  <button type="submit"
                    class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700">
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
import { useMCPServersStore } from '@/stores/mcpServers'
import { useServicesStore } from '@/stores/services'
import { servicesApi } from '@/api/services'
import { extractSchemaFields, createNestedObject, mergeNestedObjects } from '@/utils/schemaHelper'
import SchemaFieldNode from '@/components/SchemaFieldNode.vue'
import {
  fieldsToSchema,
  schemaToFields,
  createField,
  removeFieldById,
  updateFieldById,
} from '@/utils/schemaBuilder'

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

const mcpServersStore = useMCPServersStore()
const servicesStore = useServicesStore()

const servers = computed(() => mcpServersStore.servers)
const services = computed(() => servicesStore.services)

const form = reactive({
  name: '',
  description: '',
  vauth_key: '',
  service_id: ''
})

const inputParams = ref([])
const outputMappings = ref([])
const outputSchemaFields = ref([])
const outputFields = ref([])
const outputSchemaMode = ref('mapping') // 'mapping' or 'visual'
const outputMappingMode = ref('full') // 'full' or 'custom'

const resetForm = () => {
  Object.assign(form, {
    name: '',
    description: '',
    vauth_key: '',
    service_id: ''
  })
  inputParams.value = []
  outputMappings.value = []
  outputSchemaFields.value = []
  outputFields.value = []
  outputSchemaMode.value = 'mapping'
  outputMappingMode.value = 'full'
}

const initForm = (tool) => {
  if (tool) {
    Object.assign(form, {
      name: tool.name,
      description: tool.description || '',
      vauth_key: tool.vauth_key || '',
      service_id: tool.service_id || ''
    })
    inputParams.value = tool.parameters ? [...tool.parameters] : []
    outputMappings.value = tool.output_mapping ? [...tool.output_mapping] : []
    // 默认为 full 模式，如果有自定义字段标记则为 custom
    outputMappingMode.value = tool.output_mapping_mode || 'full'
  } else {
    resetForm()
  }
}

watch(() => props.visible, (newVal) => {
  if (newVal) {
    initForm(props.editingTool)
    mcpServersStore.fetchServers()
    servicesStore.fetchServices()
  }
})

watch(() => props.editingTool, (newVal) => {
  initForm(newVal)
})

const onServerChange = async () => {
  if (form.vauth_key) {
    // 使用简化版接口获取服务列表
    try {
      const response = await servicesApi.getServicesSimple()
      servicesStore.services = response.data
    } catch (error) {
      console.error('获取服务列表失败:', error)
    }
  }
}

const onServiceChange = async () => {
  if (form.service_id) {
    // 拉取完整服务信息
    try {
      const response = await servicesApi.getService(form.service_id)
      const service = response.data
      
      // 同步入参
      syncInputFromService(service)
      
      // 同步出参字段
      syncOutputFromService(service)
    } catch (error) {
      console.error('获取服务信息失败:', error)
    }
  }
}

const syncFromService = () => {
  const service = services.value.find(s => s.id === form.service_id)
  if (service) {
    syncInputFromService(service)
  }
}

const syncInputFromService = (service) => {
  if (!service) return
  const schema = service.input_schema
  if (schema && schema.properties) {
    inputParams.value = Object.entries(schema.properties).map(([name, prop]) => ({
      name,
      type: prop.type || 'string',
      description: prop.description || '',
      required: schema.required?.includes(name) || false
    }))
  }
}

const syncOutputFromService = (service) => {
  if (!service) return
  const schema = service.output_schema
  if (schema && schema.properties) {
    // 提取所有字段（包括嵌套）
    outputSchemaFields.value = extractSchemaFields(schema)
  }
}

const addInputParam = () => {
  inputParams.value.push({
    name: '',
    type: 'string',
    description: '',
    required: false
  })
}

const removeInputParam = (index) => {
  inputParams.value.splice(index, 1)
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

// 处理值类型变化
const onMappingValueTypeChange = (mapping) => {
  if (mapping.value_type === 'field') {
    mapping.default_value = ''
  } else {
    mapping.source_field = ''
  }
}

// 切换出参编辑模式
const switchToOutputSchemaMode = (mode) => {
  if (mode === 'visual' && outputMappings.value.length > 0) {
    // 从映射生成schema
    generateOutputSchemaFromMappings()
  }
  outputSchemaMode.value = mode
}

// 从映射生成出参schema
const generateOutputSchemaFromMappings = () => {
  const schema = { type: 'object', properties: {} }
  
  outputMappings.value.forEach(mapping => {
    if (mapping.source_field && mapping.target_field) {
      const nestedObject = createNestedObject(mapping.target_field, {
        type: 'string' // 默认类型，可后续编辑
      })
      mergeNestedObjects(schema.properties, nestedObject)
    }
  })
  
  outputFields.value = schemaToFields(schema)
}

// 处理出参字段更新
const handleOutputFieldUpdate = (id, updates) => {
  outputFields.value = updateFieldById(outputFields.value, id, updates)
}

// 处理出参字段删除
const handleOutputFieldDelete = (id) => {
  outputFields.value = removeFieldById(outputFields.value, id)
}

// 添加出参字段
const addOutputField = () => {
  outputFields.value = [...outputFields.value, createField()]
}

// 处理从服务同步出参
const handleSyncOutputFromService = async () => {
  if (form.service_id) {
    try {
      const response = await servicesApi.getService(form.service_id)
      const service = response.data
      syncOutputFromService(service)
    } catch (error) {
      console.error('获取服务信息失败:', error)
    }
  }
}

const handleSubmit = async () => {
  let outputSchema = null
  if (outputSchemaMode.value === 'visual' && outputFields.value.length > 0) {
    outputSchema = fieldsToSchema(outputFields.value)
  }

  const payload = {
    name: form.name,
    description: form.description,
    vauth_key: form.vauth_key,
    service_id: form.service_id,
    parameters: inputParams.value.filter(p => p.name),
    output_mapping: outputMappings.value.filter(m => m.source_field && m.target_field),
    output_schema: outputSchema
  }

  emit('saved', payload)
}
</script>
