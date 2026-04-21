<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">HTTP 服务</h2>
        <p class="text-sm text-gray-500 mt-1">管理 HTTP 服务配置，支持自定义Header、动态JS验签脚本、入参出参JSON Schema</p>
      </div>
      <button
        @click="openCreateModal"
        class="inline-flex items-center px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 btn-transition"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
        </svg>
        创建服务
      </button>
    </div>

    <!-- Services Grid -->
    <div v-if="servicesStore.loading" class="text-center py-12">
      <div class="loading-spinner mx-auto"></div>
      <p class="text-gray-500 mt-2">加载中...</p>
    </div>

    <div v-else-if="servicesStore.services.length === 0" class="text-center py-12 bg-white rounded-xl border border-gray-200">
      <svg class="w-12 h-12 text-gray-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
      </svg>
      <p class="text-gray-500">暂无 HTTP 服务</p>
      <button @click="openCreateModal" class="mt-4 text-primary-600 hover:text-primary-700 text-sm font-medium">
        创建第一个服务
      </button>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="service in servicesStore.services" :key="service.id"
        class="bg-white rounded-xl border border-gray-200 p-6 card-hover">
        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
              </svg>
            </div>
            <div>
              <h3 class="font-semibold text-gray-900">{{ service.name }}</h3>
              <span class="px-2 py-0.5 text-xs font-medium rounded-full"
                :class="service.state === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">
                {{ service.state === 1 ? '启用' : '禁用' }}
              </span>
            </div>
          </div>
        </div>

        <p class="text-sm text-gray-500 mb-3 line-clamp-2">{{ service.description || '无描述' }}</p>

        <div class="space-y-1.5 mb-3">
          <div class="flex items-center text-xs text-gray-500">
            <span class="px-2 py-0.5 bg-gray-100 rounded mr-2 font-medium">{{ service.method }}</span>
            <span v-if="service.body_type && service.body_type !== 'JSON'" class="px-2 py-0.5 bg-blue-50 text-blue-600 rounded mr-2 font-medium">{{ service.body_type }}</span>
            <span class="truncate">{{ service.target_url }}</span>
          </div>
        </div>

        <!-- Feature Tags -->
        <div class="flex flex-wrap gap-1.5 mb-3">
          <span v-if="service.request_transform_script" class="px-2 py-0.5 text-xs bg-indigo-100 text-indigo-700 rounded-full">请求转换</span>
          <span v-if="service.response_transform_script" class="px-2 py-0.5 text-xs bg-cyan-100 text-cyan-700 rounded-full">响应转换</span>
          <span v-if="service.input_schema && Object.keys(service.input_schema?.properties || {}).length" class="px-2 py-0.5 text-xs bg-emerald-100 text-emerald-700 rounded-full">入参Schema</span>
          <span v-if="service.output_schema && Object.keys(service.output_schema?.properties || {}).length" class="px-2 py-0.5 text-xs bg-teal-100 text-teal-700 rounded-full">出参Schema</span>
        </div>

        <div class="flex items-center justify-between pt-4 border-t border-gray-100">
          <div class="flex items-center space-x-1">
            <button @click="openEditModal(service)"
              class="p-2 text-gray-400 hover:text-primary-600 hover:bg-primary-50 rounded-lg transition-colors" title="编辑">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
              </svg>
            </button>
            <button @click="handleDelete(service)"
              class="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors" title="删除">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
              </svg>
            </button>
          </div>
          <button @click="openDebugModal(service)"
            class="px-3 py-1.5 text-xs font-medium text-primary-600 hover:bg-primary-50 rounded-lg transition-colors">
            调试
          </button>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <teleport to="body">
      <transition name="fade">
        <div v-if="showModal" class="fixed inset-0 z-50 overflow-y-auto">
          <div class="flex items-center justify-center min-h-screen px-4">
            <div class="fixed inset-0 bg-black bg-opacity-50" @click="showModal = false"></div>
            <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-3xl max-h-[90vh] overflow-hidden fade-in">
              <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
                <h3 class="text-lg font-semibold text-gray-900">{{ editingService ? '编辑服务' : '创建服务' }}</h3>
                <button @click="showModal = false" class="text-gray-400 hover:text-gray-600">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <div class="p-6 overflow-y-auto max-h-[calc(90vh-130px)]">
                <form @submit.prevent="handleSubmit" class="space-y-5">
                  <!-- Basic Info Section -->
                  <div class="space-y-4">
                    <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                      <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
                      基本信息
                    </h4>
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">服务名称 *</label>
                      <input v-model="form.name" type="text" required
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500">
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">描述</label>
                      <textarea v-model="form.description" rows="2"
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"></textarea>
                    </div>
                    <div class="grid grid-cols-2 gap-4">
                      <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1">HTTP 方法</label>
                        <select v-model="form.method"
                          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500">
                          <option value="GET">GET</option>
                          <option value="POST">POST</option>
                          <option value="PUT">PUT</option>
                          <option value="DELETE">DELETE</option>
                          <option value="PATCH">PATCH</option>
                        </select>
                      </div>
                      <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1">超时(秒)</label>
                        <input v-model.number="form.timeout_seconds" type="number" min="1"
                          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500">
                      </div>
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">请求体类型</label>
                      <select v-model="form.body_type"
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500">
                        <option value="none">none</option>
                        <option value="form-data">form-data</option>
                        <option value="urlencoded">x-www-form-urlencoded</option>
                        <option value="JSON">JSON (默认)</option>
                        <option value="raw">raw</option>
                        <option value="binary">binary</option>
                        <option value="msgpack">msgpack</option>
                      </select>
                      <p class="text-xs text-gray-400 mt-1">调试时将根据此类型自动设置 Content-Type</p>
                    </div>
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">目标 URL *</label>
                      <input v-model="form.target_url" type="url" required
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                        placeholder="https://api.example.com/endpoint">
                    </div>
                  </div>

                  <!-- Headers Section -->
                  <div class="space-y-3 pt-4 border-t border-gray-100">
                    <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                      <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z"/></svg>
                      自定义请求头
                    </h4>
                    <div v-for="(header, index) in headersList" :key="index" class="flex gap-2">
                      <input v-model="header.key" placeholder="Header Name"
                        class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500">
                      <input v-model="header.value" placeholder="Header Value"
                        class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-primary-500 focus:border-primary-500">
                      <button type="button" @click="removeHeader(index)"
                        class="p-2 text-red-400 hover:text-red-600 hover:bg-red-50 rounded-lg">
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                      </button>
                    </div>
                    <button type="button" @click="addHeader"
                      class="text-sm text-primary-600 hover:text-primary-700 font-medium">
                      + 添加请求头
                    </button>
                  </div>

                  <!-- Transform Scripts Section -->
                  <div class="space-y-3 pt-4 border-t border-gray-100">
                    <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                      <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
                      请求/响应转换脚本
                    </h4>
                    <div class="flex items-center gap-4">
                      <label class="flex items-center">
                        <input v-model="requestTransformEnabled" type="checkbox" class="rounded border-gray-300 text-primary-600 focus:ring-primary-500">
                        <span class="ml-2 text-sm text-gray-700">启用请求转换</span>
                      </label>
                      <label class="flex items-center">
                        <input v-model="responseTransformEnabled" type="checkbox" class="rounded border-gray-300 text-primary-600 focus:ring-primary-500">
                        <span class="ml-2 text-sm text-gray-700">启用响应转换</span>
                      </label>
                    </div>
                    <div v-if="requestTransformEnabled">
                      <div class="flex items-center justify-between mb-1">
                        <label class="block text-sm font-medium text-gray-700">
                          请求转换脚本
                          <span class="text-xs text-gray-400 font-normal ml-1">（发送请求前执行）</span>
                        </label>
                        <button type="button" @click="showRequestTransformExample = true" class="text-xs text-primary-600 hover:text-primary-700 font-medium">
                          查看示例
                        </button>
                      </div>
                      <textarea v-model="form.request_transform_script" rows="5" placeholder="// 可用变量: context.headers, context.body, context.time, context.method, context.url&#10;// 设置 transformedHeaders, transformedBody, transformedURLParams 修改请求数据"
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-xs focus:ring-2 focus:ring-primary-500 focus:border-primary-500"></textarea>
                    </div>
                    <div v-if="responseTransformEnabled">
                      <div class="flex items-center justify-between mb-1">
                        <label class="block text-sm font-medium text-gray-700">
                          响应转换脚本
                          <span class="text-xs text-gray-400 font-normal ml-1">（收到响应后执行）</span>
                        </label>
                        <button type="button" @click="showResponseTransformExample = true" class="text-xs text-primary-600 hover:text-primary-700 font-medium">
                          查看示例
                        </button>
                      </div>
                      <textarea v-model="form.response_transform_script" rows="5" placeholder="// 可用变量: context.headers, context.body, context.time&#10;// 设置 transformedHeaders, transformedBody 修改响应数据"
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-xs focus:ring-2 focus:ring-primary-500 focus:border-primary-500"></textarea>
                    </div>
                  </div>

                  <!-- JSON Schema Section (Dual Mode) -->
                  <div class="space-y-4 pt-4 border-t border-gray-100">
                    <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                      <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
                      入参 Schema
                    </h4>

                    <!-- Mode Switch -->
                    <div class="flex items-center gap-2">
                      <button type="button" @click="inputSchemaMode = 'visual'"
                        :class="inputSchemaMode === 'visual' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
                        class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors">
                        可视化编辑
                      </button>
                      <button type="button" @click="switchToJsonMode('input')"
                        :class="inputSchemaMode === 'json' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
                        class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors">
                        JSON 编辑
                      </button>
                      <button type="button" @click="openPasteModal('input')"
                        class="px-3 py-1.5 text-xs font-medium rounded-lg bg-emerald-50 text-emerald-700 hover:bg-emerald-100 transition-colors">
                        粘贴 JSON 生成
                      </button>
                    </div>

                    <!-- Visual Mode: Input Schema -->
                    <div v-if="inputSchemaMode === 'visual'" class="space-y-2">
                      <SchemaFieldNode
                        v-for="field in inputFields"
                        :key="field.id"
                        :field="field"
                        :depth="0"
                        @update="handleInputFieldUpdate"
                        @delete="handleInputFieldDelete"
                      />
                      <button type="button" @click="addInputField"
                        class="text-sm text-primary-600 hover:text-primary-700 font-medium">
                        + 添加字段
                      </button>
                    </div>

                    <!-- JSON Mode: Input Schema -->
                    <div v-if="inputSchemaMode === 'json'">
                      <textarea v-model="inputSchemaStr" rows="8"
                        placeholder='{"type":"object","properties":{"query":{"type":"string","description":"搜索关键词"}},"required":["query"]}'
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-xs focus:ring-2 focus:ring-primary-500 focus:border-primary-500"></textarea>
                      <p v-if="inputSchemaError" class="text-xs text-red-500 mt-1">{{ inputSchemaError }}</p>
                      <button type="button" @click="syncInputJsonToFields"
                        class="mt-1 text-xs text-gray-500 hover:text-primary-600">
                        ← 从 JSON 同步到可视化
                      </button>
                    </div>
                  </div>

                  <!-- Output Schema (Dual Mode) -->
                  <div class="space-y-4 pt-4 border-t border-gray-100">
                    <h4 class="text-sm font-semibold text-gray-800 flex items-center">
                      <svg class="w-4 h-4 mr-1.5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
                      出参 Schema
                    </h4>

                    <div class="flex items-center gap-2">
                      <button type="button" @click="outputSchemaMode = 'visual'"
                        :class="outputSchemaMode === 'visual' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
                        class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors">
                        可视化编辑
                      </button>
                      <button type="button" @click="switchToJsonMode('output')"
                        :class="outputSchemaMode === 'json' ? 'bg-primary-600 text-white' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
                        class="px-3 py-1.5 text-xs font-medium rounded-lg transition-colors">
                        JSON 编辑
                      </button>
                      <button type="button" @click="openPasteModal('output')"
                        class="px-3 py-1.5 text-xs font-medium rounded-lg bg-emerald-50 text-emerald-700 hover:bg-emerald-100 transition-colors">
                        粘贴 JSON 生成
                      </button>
                    </div>

                    <!-- Visual Mode: Output Schema -->
                    <div v-if="outputSchemaMode === 'visual'" class="space-y-2">
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

                    <!-- JSON Mode: Output Schema -->
                    <div v-if="outputSchemaMode === 'json'">
                      <textarea v-model="outputSchemaStr" rows="8"
                        placeholder='{"type":"object","properties":{"code":{"type":"integer"},"data":{"type":"object"}}}'
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-xs focus:ring-2 focus:ring-primary-500 focus:border-primary-500"></textarea>
                      <p v-if="outputSchemaError" class="text-xs text-red-500 mt-1">{{ outputSchemaError }}</p>
                      <button type="button" @click="syncOutputJsonToFields"
                        class="mt-1 text-xs text-gray-500 hover:text-primary-600">
                        ← 从 JSON 同步到可视化
                      </button>
                    </div>
                  </div>

                  <!-- Toggle -->
                  <div class="pt-4 border-t border-gray-100">
                    <label class="flex items-center">
                      <input v-model="form.state" type="checkbox" :true-value="1" :false-value="0" class="rounded border-gray-300 text-primary-600 focus:ring-primary-500">
                      <span class="ml-2 text-sm text-gray-700">启用服务</span>
                    </label>
                  </div>

                  <div class="flex justify-end space-x-3 pt-4">
                    <button type="button" @click="showModal = false"
                      class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50">
                      取消
                    </button>
                    <button type="submit" :disabled="servicesStore.loading"
                      class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50">
                      {{ servicesStore.loading ? '保存中...' : '保存' }}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </teleport>

    <!-- Example Modal: Request Transform -->
    <teleport to="body">
      <transition name="fade">
        <div v-if="showRequestTransformExample" class="fixed inset-0 z-50 overflow-y-auto">
          <div class="flex items-center justify-center min-h-screen px-4">
            <div class="fixed inset-0 bg-black bg-opacity-50" @click="showRequestTransformExample = false"></div>
            <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-3xl max-h-[90vh] overflow-hidden fade-in">
              <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
                <h3 class="text-lg font-semibold text-gray-900">请求转换脚本示例</h3>
                <button @click="showRequestTransformExample = false" class="text-gray-400 hover:text-gray-600">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <div class="p-6 overflow-y-auto max-h-[calc(90vh-130px)]">
                <pre class="bg-gray-900 text-gray-100 p-4 rounded-lg text-xs font-mono whitespace-pre-wrap overflow-x-auto">{{ requestTransformExamples }}</pre>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </teleport>

    <!-- Example Modal: Response Transform -->
    <teleport to="body">
      <transition name="fade">
        <div v-if="showResponseTransformExample" class="fixed inset-0 z-50 overflow-y-auto">
          <div class="flex items-center justify-center min-h-screen px-4">
            <div class="fixed inset-0 bg-black bg-opacity-50" @click="showResponseTransformExample = false"></div>
            <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-3xl max-h-[90vh] overflow-hidden fade-in">
              <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
                <h3 class="text-lg font-semibold text-gray-900">响应转换脚本示例</h3>
                <button @click="showResponseTransformExample = false" class="text-gray-400 hover:text-gray-600">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <div class="p-6 overflow-y-auto max-h-[calc(90vh-130px)]">
                <pre class="bg-gray-900 text-gray-100 p-4 rounded-lg text-xs font-mono whitespace-pre-wrap overflow-x-auto">{{ responseTransformExamples }}</pre>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </teleport>

    <!-- Debug Modal -->
    <teleport to="body">
      <transition name="fade">
        <div v-if="showDebugModal" class="fixed inset-0 z-50 overflow-y-auto">
          <div class="flex items-center justify-center min-h-screen px-4">
            <div class="fixed inset-0 bg-black bg-opacity-50" @click="showDebugModal = false"></div>
            <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-4xl max-h-[90vh] overflow-hidden fade-in">
              <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <h3 class="text-lg font-semibold text-gray-900">接口调试</h3>
                  <span class="px-2 py-0.5 text-xs font-medium bg-blue-100 text-blue-700 rounded-full">{{ debugService?.name }}</span>
                  <span class="px-2 py-0.5 text-xs font-medium bg-gray-100 text-gray-600 rounded-full">{{ debugService?.method }}</span>
                </div>
                <button @click="showDebugModal = false" class="text-gray-400 hover:text-gray-600">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <div class="p-6 overflow-y-auto max-h-[calc(90vh-130px)]">
                <div class="grid grid-cols-2 gap-6">
                  <!-- Left: Request -->
                  <div class="space-y-4">
                    <h4 class="text-sm font-semibold text-gray-800">请求配置</h4>
                    <div class="text-xs text-gray-500 mb-2">{{ debugService?.target_url }}</div>

                    <!-- Request Headers -->
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1">请求头</label>
                      <div v-for="(h, i) in debugHeaders" :key="i" class="flex gap-1 mb-1">
                        <input v-model="h.key" placeholder="Key" class="flex-1 px-2 py-1 border border-gray-300 rounded text-xs">
                        <input v-model="h.value" placeholder="Value" class="flex-1 px-2 py-1 border border-gray-300 rounded text-xs">
                        <button @click="debugHeaders.splice(i, 1)" class="text-red-400 hover:text-red-600 px-1">
                          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                        </button>
                      </div>
                      <button @click="debugHeaders.push({ key: '', value: '' })" class="text-xs text-primary-600 hover:text-primary-700">+ 添加</button>
                    </div>

                    <!-- Body Type Selector -->
                    <div>
                      <label class="block text-sm font-medium text-gray-700 mb-1.5">请求体类型</label>
                      <div class="flex flex-wrap gap-1">
                        <button v-for="opt in bodyTypeOptions" :key="opt.value"
                          type="button"
                          @click="onDebugBodyTypeChange(opt.value)"
                          :class="debugBodyType === opt.value
                            ? 'bg-primary-600 text-white shadow-sm'
                            : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
                          class="px-2.5 py-1 text-xs font-medium rounded-md transition-colors">
                          {{ opt.label }}
                        </button>
                      </div>
                    </div>

                    <!-- Request Body: form-data / urlencoded key-value editor -->
                    <div v-if="showBodyEditor && (debugBodyType === 'form-data' || debugBodyType === 'urlencoded')">
                      <label class="block text-sm font-medium text-gray-700 mb-1">
                        请求体
                        <span class="text-xs text-gray-400 font-normal">({{ debugBodyType === 'form-data' ? 'Multipart Form' : 'URL Encoded' }} 键值对)</span>
                      </label>
                      <div v-for="(item, i) in debugFormFields" :key="i" class="flex gap-1 mb-1">
                        <input v-model="item.key" placeholder="Key" class="flex-1 px-2 py-1 border border-gray-300 rounded text-xs">
                        <input v-model="item.value" placeholder="Value" class="flex-1 px-2 py-1 border border-gray-300 rounded text-xs">
                        <button @click="debugFormFields.splice(i, 1)" class="text-red-400 hover:text-red-600 px-1">
                          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
                        </button>
                      </div>
                      <button @click="debugFormFields.push({ key: '', value: '' })" class="text-xs text-primary-600 hover:text-primary-700">+ 添加字段</button>
                    </div>

                    <!-- Request Body: JSON / raw / binary / msgpack text editor -->
                    <div v-if="showBodyEditor && !['form-data', 'urlencoded'].includes(debugBodyType)">
                      <label class="block text-sm font-medium text-gray-700 mb-1">
                        请求体
                        <span v-if="debugBodyType === 'JSON'" class="text-xs text-gray-400 font-normal">(JSON)</span>
                        <span v-else-if="debugBodyType === 'raw'" class="text-xs text-gray-400 font-normal">(Text)</span>
                        <span v-else-if="debugBodyType === 'binary'" class="text-xs text-gray-400 font-normal">(Base64 or raw text)</span>
                        <span v-else-if="debugBodyType === 'msgpack'" class="text-xs text-gray-400 font-normal">(Hex or raw text)</span>
                      </label>
                      <textarea v-model="debugBodyStr" rows="6"
                        class="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-xs focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                        :placeholder="bodyPlaceholder"
                        @blur="formatRequestBody"
                        spellcheck="false"></textarea>
                      <div v-if="debugBodyType === 'JSON' && parsedDebugBody" class="mt-2 border border-gray-200 rounded-lg overflow-hidden max-h-60 overflow-y-auto">
                        <JsonViewer :value="parsedDebugBody" :expand-depth="3" sort />
                      </div>
                      <p v-if="debugBodyError" class="text-xs text-red-500 mt-1">{{ debugBodyError }}</p>
                    </div>

                    <!-- Input Schema hint -->
                    <div v-if="debugService?.input_schema && Object.keys(debugService.input_schema?.properties || {}).length"
                      class="bg-blue-50 border border-blue-200 rounded-lg p-3">
                      <div class="text-xs font-medium text-blue-700 mb-1">入参 Schema 参考</div>
                      <div class="text-xs text-blue-600 space-y-0.5">
                        <div v-for="(prop, key) in debugService.input_schema.properties" :key="key">
                          <span class="font-mono">{{ key }}</span>
                          <span class="text-blue-400">: {{ prop.type }}</span>
                          <span v-if="debugService.input_schema.required?.includes(key)" class="text-red-500">*</span>
                          <span v-if="prop.description" class="text-blue-400"> - {{ prop.description }}</span>
                        </div>
                      </div>
                    </div>

                    <button @click="runDebug"
                      :disabled="debugLoading"
                      class="w-full px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50 flex items-center justify-center gap-2">
                      <svg v-if="debugLoading" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                      {{ debugLoading ? '请求中...' : '发送请求' }}
                    </button>
                  </div>

                  <!-- Right: Response -->
                  <div class="space-y-4">
                    <h4 class="text-sm font-semibold text-gray-800">响应结果</h4>
                    <div v-if="debugResult" class="space-y-3">
                      <!-- Status -->
                      <div class="flex items-center gap-2">
                        <span class="text-sm font-medium">状态码:</span>
                        <span class="px-2 py-0.5 text-xs font-medium rounded-full"
                          :class="debugResult.success ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">
                          {{ debugResult.status_code || 'N/A' }}
                        </span>
                        <span class="text-xs text-gray-500">{{ debugResult.duration_ms }}ms</span>
                      </div>

                      <!-- Response Headers -->
                      <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1">响应头</label>
                        <div class="bg-gray-50 rounded-lg p-3 max-h-32 overflow-y-auto">
                          <div v-for="(v, k) in debugResult.response_headers" :key="k"
                            class="text-xs font-mono text-gray-600 flex gap-2">
                            <span class="font-semibold">{{ k }}:</span>
                            <span class="truncate">{{ v }}</span>
                          </div>
                          <p v-if="!debugResult.response_headers || !Object.keys(debugResult.response_headers).length"
                            class="text-xs text-gray-400">无</p>
                        </div>
                      </div>

                      <!-- Response Body -->
                      <div>
                        <label class="block text-sm font-medium text-gray-700 mb-1">响应体</label>
                        <div v-if="hasResponseBody" class="border border-gray-200 rounded-lg overflow-hidden max-h-80 overflow-y-auto">
                          <VueJsonPretty v-if="isResponseBodyObject" :data="debugResult.response_body" :deep="3" :show-line="false" />
                          <pre v-else class="p-3 text-xs font-mono text-gray-700 whitespace-pre-wrap break-all">{{ formatResponseBody }}</pre>
                        </div>
                        <div v-else class="bg-gray-50 rounded-lg p-3 text-xs text-gray-400">无响应体</div>
                      </div>

                      <!-- Error -->
                      <div v-if="debugResult.error" class="bg-red-50 border border-red-200 rounded-lg p-3">
                        <div class="text-xs font-medium text-red-700 mb-1">错误信息</div>
                        <div class="text-xs text-red-600">{{ debugResult.error }}</div>
                      </div>

                      <!-- Output Schema validation hint -->
                      <div v-if="debugService?.output_schema && Object.keys(debugService.output_schema?.properties || {}).length"
                        class="bg-green-50 border border-green-200 rounded-lg p-3">
                        <div class="text-xs font-medium text-green-700 mb-1">出参 Schema 参考</div>
                        <div class="text-xs text-green-600 space-y-0.5">
                          <div v-for="(prop, key) in debugService.output_schema.properties" :key="key">
                            <span class="font-mono">{{ key }}</span>
                            <span class="text-green-400">: {{ prop.type }}</span>
                            <span v-if="debugService.output_schema.required?.includes(key)" class="text-red-500">*</span>
                          </div>
                        </div>
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

    <!-- Paste JSON Modal -->
    <teleport to="body">
      <transition name="fade">
        <div v-if="showPasteModal" class="fixed inset-0 z-50 overflow-y-auto">
          <div class="flex items-center justify-center min-h-screen px-4">
            <div class="fixed inset-0 bg-black bg-opacity-50" @click="showPasteModal = false"></div>
            <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-3xl max-h-[90vh] overflow-hidden fade-in">
              <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
                <h3 class="text-lg font-semibold text-gray-900">粘贴 JSON 生成 Schema</h3>
                <button @click="showPasteModal = false" class="text-gray-400 hover:text-gray-600">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <div class="p-6 overflow-y-auto max-h-[calc(90vh-130px)]">
                <div class="space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">粘贴 JSON 样本</label>
                    <textarea v-model="pasteJsonStr" rows="8"
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg font-mono text-xs focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      placeholder='{"code": 0, "data": {"error": "", "status_code": 200}}'></textarea>
                    <p v-if="pasteError" class="text-xs text-red-500 mt-1">{{ pasteError }}</p>
                  </div>
                  <div class="flex gap-2">
                    <button @click="parsePasteJson" class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700">
                      解析
                    </button>
                  </div>

                  <!-- Preview -->
                  <div v-if="pastePreviewFields.length > 0">
                    <label class="block text-sm font-medium text-gray-700 mb-1">预览生成的字段</label>
                    <div class="border border-gray-200 rounded-lg p-4 bg-gray-50 max-h-60 overflow-y-auto">
                      <div v-for="field in pastePreviewFields" :key="field.id">
                        <div class="text-xs font-mono text-gray-700">
                          {{ field.name }}: {{ field.type }}
                          <span v-if="field.children && field.children.length > 0" class="text-gray-400 ml-1">
                            (含 {{ field.children.length }} 个子字段)
                          </span>
                        </div>
                        <div v-if="field.children" class="ml-4 mt-1">
                          <div v-for="child in field.children" :key="child.id" class="text-xs font-mono text-gray-500">
                            └ {{ child.name }}: {{ child.type }}
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
                <button @click="showPasteModal = false" class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50">
                  取消
                </button>
                <button @click="applyPasteJson" :disabled="pastePreviewFields.length === 0"
                  class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50">
                  应用到 {{ pasteTarget === 'input' ? '入参' : '出参' }} Schema
                </button>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </teleport>
  </div>
</template>

<script setup>
import { ref, reactive, inject, onMounted, computed, watch } from 'vue'
import { useServicesStore } from '@/stores/services'
import { servicesApi } from '@/api/services'
import VueJsonPretty from 'vue-json-pretty'
import 'vue-json-pretty/lib/styles.css'
import SchemaFieldNode from '@/components/SchemaFieldNode.vue'
import {
  fieldsToSchema,
  schemaToFields,
  parseJSONToFields,
  createField,
  removeFieldById,
  updateFieldById,
} from '@/utils/schemaBuilder'

const servicesStore = useServicesStore()
const showToast = inject('showToast')

// Service form modal state
const showModal = ref(false)
const editingService = ref(null)
const headersList = ref([{ key: '', value: '' }])

// Schema editing state
const inputSchemaMode = ref('visual')  // 'visual' or 'json'
const outputSchemaMode = ref('visual')
const inputSchemaStr = ref('')
const outputSchemaStr = ref('')
const inputSchemaError = ref('')
const outputSchemaError = ref('')

// Visual schema fields
const inputFields = ref([])
const outputFields = ref([])

// Paste modal state
const showPasteModal = ref(false)
const pasteTarget = ref('output')
const pasteJsonStr = ref('')
const pastePreviewFields = ref([])
const pasteError = ref('')

// Transform script toggles
const requestTransformEnabled = ref(false)
const responseTransformEnabled = ref(false)

// Example modal state
const showRequestTransformExample = ref(false)
const showResponseTransformExample = ref(false)

// Debug modal state
const showDebugModal = ref(false)
const debugService = ref(null)
const debugHeaders = ref([])
const debugBodyStr = ref('')
const debugBodyError = ref('')
const debugResult = ref(null)
const debugLoading = ref(false)
const debugBodyType = ref('JSON') // none, form-data, urlencoded, JSON, raw, binary, msgpack
const debugFormFields = ref([{ key: '', value: '' }]) // form-data / urlencoded 键值对

const bodyTypeOptions = [
  { value: 'none', label: 'none' },
  { value: 'form-data', label: 'form-data' },
  { value: 'urlencoded', label: 'x-www-form-urlencoded' },
  { value: 'JSON', label: 'JSON' },
  { value: 'raw', label: 'raw' },
  { value: 'binary', label: 'binary' },
  { value: 'msgpack', label: 'msgpack' },
]

const bodyTypeContentTypeMap = {
  'form-data': 'multipart/form-data',
  'urlencoded': 'application/x-www-form-urlencoded',
  'JSON': 'application/json',
  'binary': 'application/octet-stream',
  'msgpack': 'application/msgpack',
  'raw': 'text/plain',
}

const showBodyEditor = computed(() => !['none'].includes(debugBodyType.value))

const bodyPlaceholder = computed(() => {
  if (debugBodyType.value === 'raw') return 'raw text content'
  return '{"key": "value"}'
})

const onDebugBodyTypeChange = (newType) => {
  const oldType = debugBodyType.value
  debugBodyType.value = newType

  // 移除旧类型自动添加的 Content-Type
  const oldCT = bodyTypeContentTypeMap[oldType]
  if (oldCT) {
    debugHeaders.value = debugHeaders.value.filter(h => h.key.toLowerCase() !== 'content-type')
  }

  // 自动添加新类型的 Content-Type
  const newCT = bodyTypeContentTypeMap[newType]
  if (newCT) {
    const hasCT = debugHeaders.value.some(h => h.key.toLowerCase() === 'content-type')
    if (!hasCT) {
      debugHeaders.value.push({ key: 'Content-Type', value: newCT })
    }
  }

  // 切换到 form-data / urlencoded 时：从 JSON 文本解析键值对
  if ((newType === 'form-data' || newType === 'urlencoded') && !(oldType === 'form-data' || oldType === 'urlencoded')) {
    if (debugBodyStr.value.trim()) {
      try {
        const obj = JSON.parse(debugBodyStr.value)
        if (typeof obj === 'object' && obj !== null && !Array.isArray(obj)) {
          debugFormFields.value = Object.entries(obj).map(([k, v]) => ({ key: k, value: String(v) }))
        }
      } catch { /* ignore */ }
    }
    if (debugFormFields.value.length === 0) {
      debugFormFields.value = [{ key: '', value: '' }]
    }
  }

  // 切换回 JSON / raw 等文本模式时：从键值对合成 JSON
  if (!(newType === 'form-data' || newType === 'urlencoded') && (oldType === 'form-data' || oldType === 'urlencoded')) {
    const obj = {}
    debugFormFields.value.forEach(f => { if (f.key) obj[f.key] = f.value })
    if (Object.keys(obj).length) {
      debugBodyStr.value = JSON.stringify(obj, null, 2)
    }
  }
}

const form = reactive({
  name: '',
  description: '',
  method: 'POST',
  target_url: '',
  timeout_seconds: 30,
  body_type: 'JSON',
  state: 1,
  request_transform_script: '',
  response_transform_script: '',
})

// ---- Schema Visual Editor Helpers ----
const addInputField = () => {
  inputFields.value = [...inputFields.value, createField()]
}
const addOutputField = () => {
  outputFields.value = [...outputFields.value, createField()]
}

// 处理字段更新
const handleInputFieldUpdate = (id, updates) => {
  inputFields.value = updateFieldById(inputFields.value, id, updates)
}
const handleOutputFieldUpdate = (id, updates) => {
  outputFields.value = updateFieldById(outputFields.value, id, updates)
}

// 处理字段删除
const handleInputFieldDelete = (id) => {
  inputFields.value = removeFieldById(inputFields.value, id)
}
const handleOutputFieldDelete = (id) => {
  outputFields.value = removeFieldById(outputFields.value, id)
}

// 粘贴解析相关
const openPasteModal = (target) => {
  pasteTarget.value = target
  pasteJsonStr.value = ''
  pasteError.value = ''
  pastePreviewFields.value = []
  showPasteModal.value = true
}

const parsePasteJson = () => {
  pasteError.value = ''
  pastePreviewFields.value = []
  try {
    const parsed = JSON.parse(pasteJsonStr.value)
    pastePreviewFields.value = parseJSONToFields(parsed)
  } catch (e) {
    pasteError.value = 'JSON 格式不正确: ' + e.message
  }
}

const applyPasteJson = () => {
  if (pasteTarget.value === 'input') {
    inputFields.value = pastePreviewFields.value
  } else {
    outputFields.value = pastePreviewFields.value
  }
  showPasteModal.value = false
}

// Visual fields -> Schema JSON
const syncInputFieldsToSchema = () => {
  inputSchemaStr.value = JSON.stringify(fieldsToSchema(inputFields.value), null, 2)
}
const syncOutputFieldsToSchema = () => {
  outputSchemaStr.value = JSON.stringify(fieldsToSchema(outputFields.value), null, 2)
}

const switchToJsonMode = (target) => {
  if (target === 'input') {
    syncInputFieldsToSchema()
    inputSchemaMode.value = 'json'
  } else {
    syncOutputFieldsToSchema()
    outputSchemaMode.value = 'json'
  }
}

const syncInputJsonToFields = () => {
  try {
    const schema = JSON.parse(inputSchemaStr.value)
    inputFields.value = schemaToFields(schema)
    inputSchemaError.value = ''
  } catch {
    inputSchemaError.value = 'JSON 格式不正确'
  }
}
const syncOutputJsonToFields = () => {
  try {
    const schema = JSON.parse(outputSchemaStr.value)
    outputFields.value = schemaToFields(schema)
    outputSchemaError.value = ''
  } catch {
    outputSchemaError.value = 'JSON 格式不正确'
  }
}

// ---- Header Helpers ----
const addHeader = () => headersList.value.push({ key: '', value: '' })
const removeHeader = (index) => headersList.value.splice(index, 1)

// 示例脚本
const requestTransformExamples = `// ========================================
// 请求转换脚本说明
// ========================================
// 可用变量:
//   context.headers - 请求头 (Map)
//   context.body    - 请求体 (Map/Object)
//   context.time    - 当前时间戳 (毫秒)
//   context.method  - 请求方法 (GET/POST等)
//   context.url     - 请求URL
//
// 可设置以下变量修改请求:
//   transformedHeaders - 新的请求头 (Map)
//   transformedBody    - 新的请求体 (Map/Object)
//   transformedURLParams - URL查询参数 (Map)
//
// 可用加密函数:
//   hmacSha256(key, data) / hmacSha1 / hmacSha512 / hmacMd5
//   sha256(data) / sha1(data) / sha512(data) / md5(data)
//   base64Encode/Decode  hexEncode/Decode  urlEncode/Decode
//   aesEncrypt(data, key, iv) / aesDecrypt(data, key, iv)
//   rsaSign(data, privateKey) / rsaVerify(data, signature, publicKey)
//
// ========================================
// 示例1: 简单签名 - HMAC-SHA256
// ========================================
var timestamp = String(context.time);
var bodyStr = JSON.stringify(context.body || {});
var stringToSign = context.method + context.url + timestamp + bodyStr;
var signature = hmacSha256('your-secret-key', stringToSign);
transformedBody = Object.assign({}, context.body, {
  sign: signature,
  timestamp: timestamp
});

// ========================================
// 示例2: 签名放入Header + Body
// ========================================
var timestamp = String(context.time);
var nonce = Math.random().toString(36).substr(2, 15);
var bodyStr = JSON.stringify(sortByKeys(context.body || {}));
var stringToSign = context.method + '\\n' + context.url + '\\n' + timestamp + '\\n' + nonce + '\\n' + bodyStr;
var signature = hmacSha256('your-secret-key', stringToSign);
transformedHeaders = Object.assign({}, context.headers, {
  'X-Signature': signature,
  'X-Timestamp': timestamp,
  'X-Nonce': nonce
});
transformedBody = context.body;

// ========================================
// 示例3: 签名放入URL参数
// ========================================
var timestamp = String(Math.floor(context.time / 1000));
var bodyStr = JSON.stringify(context.body || {});
var stringToSign = bodyStr + timestamp;
var signature = hexEncode(hmacSha256('your-secret-key', stringToSign));
transformedURLParams = {
  app_id: 'your-app-id',
  sign: signature,
  timestamp: timestamp
};
transformedBody = context.body;

// ========================================
// 示例4: MD5签名 (常见于旧版API)
// ========================================
var bodyStr = JSON.stringify(context.body || {});
var stringToSign = bodyStr + 'your-secret-key';
var signature = md5(stringToSign);
transformedBody = Object.assign({}, context.body, {
  sign: signature
});

// ========================================
// 示例5: AES加密 + HMAC签名
// ========================================
var timestamp = String(context.time);
var bodyStr = JSON.stringify(context.body || {});
var encrypted = aesEncrypt(bodyStr, 'your-aes-key', 'your-iv-16char');
var stringToSign = encrypted + timestamp;
var signature = hmacSha256('your-secret-key', stringToSign);
transformedHeaders = Object.assign({}, context.headers, {
  'X-Encrypted': encrypted,
  'X-Timestamp': timestamp,
  'X-Signature': signature
});
transformedBody = null;

// ========================================
// 示例6: RSA签名
// ========================================
var bodyStr = JSON.stringify(context.body || {});
var signature = rsaSign(bodyStr, 'your-private-key');
transformedHeaders = Object.assign({}, context.headers, {
  'X-Signature': signature
});
transformedBody = context.body;

// ========================================
// 示例7: 微信支付风格签名
// ========================================
var timestamp = String(Math.floor(context.time / 1000));
var nonceStr = timestamp + Math.random().toString(36).substr(2, 9);
var params = Object.assign({}, context.body, {
  mch_id: 'your-merchant-id',
  nonce_str: nonceStr,
  timestamp: timestamp
});
var sortedStr = sortAndJoin(params, '&', '=');
var signature = hmacSha256('your-secret-key', sortedStr).toUpperCase();
transformedBody = Object.assign({}, params, {
  sign: signature
});`;

const responseTransformExamples = `// ========================================
// 响应转换脚本说明
// ========================================
// 可用变量:
//   context.headers - 响应头 (Map)
//   context.body    - 响应体 (Map/Object)
//   context.time    - 当前时间戳 (毫秒)
//
// 可设置以下变量修改响应:
//   transformedHeaders - 新的响应头 (Map)
//   transformedBody    - 新的响应体 (Map/Object)
//
// ========================================
// 示例1: 统一响应格式
// ========================================
transformedBody = {
  code: context.body.code || 0,
  message: context.body.message || 'success',
  data: context.body.data || context.body
};

// ========================================
// 示例2: 提取data字段
// ========================================
if (context.body && context.body.data !== undefined) {
  transformedBody = context.body.data;
} else {
  transformedBody = context.body;
};

// ========================================
// 示例3: 解密响应数据
// ========================================
// if (context.body.encrypted) {
//   var decrypted = aesDecrypt(context.body.encrypted, 'your-key', 'your-iv');
//   transformedBody = JSON.parse(decrypted);
// } else {
//   transformedBody = context.body;
// };

// ========================================
// 示例4: 验证响应签名
// ========================================
// var receivedSign = context.headers['x-signature'];
// var bodyStr = JSON.stringify(context.body);
// var expectedSign = hmacSha256('your-secret-key', bodyStr);
// if (receivedSign !== expectedSign) {
//   transformedBody = { error: 'signature mismatch' };
// } else {
//   transformedBody = context.body;
// };`;

const defaultRequestTransformScript = '';
const defaultResponseTransformScript = '';


// 监听请求转换启用状态，自动填充默认模板
watch(requestTransformEnabled, (newVal) => {
  if (newVal && !form.request_transform_script) {
    form.request_transform_script = defaultRequestTransformScript
  }
})

// 监听响应转换启用状态，自动填充默认模板
watch(responseTransformEnabled, (newVal) => {
  if (newVal && !form.response_transform_script) {
    form.response_transform_script = defaultResponseTransformScript
  }
})

const headersToMap = () => {
  const map = {}
  headersList.value.forEach(h => { if (h.key && h.value) map[h.key] = h.value })
  return Object.keys(map).length ? map : undefined
}
const mapToHeaders = (map) => {
  if (!map || typeof map !== 'object') return [{ key: '', value: '' }]
  return Object.entries(map).map(([key, value]) => ({ key, value: String(value) }))
}

// ---- Modal Open/Close ----
const openCreateModal = () => {
  editingService.value = null
  Object.assign(form, {
    name: '', description: '', method: 'POST', target_url: '',
    timeout_seconds: 30, body_type: 'JSON', state: 1,
    request_transform_script: '', response_transform_script: '',
  })
  headersList.value = [{ key: '', value: '' }]
  inputSchemaStr.value = ''
  outputSchemaStr.value = ''
  inputSchemaError.value = ''
  outputSchemaError.value = ''
  inputFields.value = []
  outputFields.value = []
  inputSchemaMode.value = 'visual'
  outputSchemaMode.value = 'visual'
  requestTransformEnabled.value = false
  responseTransformEnabled.value = false
  showModal.value = true
}

const openEditModal = (service) => {
  editingService.value = service
  Object.assign(form, {
    name: service.name,
    description: service.description || '',
    method: service.method,
    target_url: service.target_url,
    timeout_seconds: service.timeout_seconds,
    body_type: service.body_type || 'JSON',
    state: service.state,
    request_transform_script: service.request_transform_script || '',
    response_transform_script: service.response_transform_script || '',
  })
  headersList.value = mapToHeaders(service.headers)

  // Parse schemas for both modes
  const inputSchema = service.input_schema || { type: 'object', properties: {} }
  const outputSchema = service.output_schema || { type: 'object', properties: {} }
  inputSchemaStr.value = JSON.stringify(inputSchema, null, 2)
  outputSchemaStr.value = JSON.stringify(outputSchema, null, 2)
  inputFields.value = schemaToFields(inputSchema)
  outputFields.value = schemaToFields(outputSchema)
  inputSchemaMode.value = 'visual'
  outputSchemaMode.value = 'visual'
  inputSchemaError.value = ''
  outputSchemaError.value = ''
  requestTransformEnabled.value = !!service.request_transform_script
  responseTransformEnabled.value = !!service.response_transform_script
  showModal.value = true
}

// ---- Submit ----
const handleSubmit = async () => {
  inputSchemaError.value = ''
  outputSchemaError.value = ''

  // Ensure visual mode schemas are synced
  if (inputSchemaMode.value === 'visual') syncInputFieldsToSchema()
  if (outputSchemaMode.value === 'visual') syncOutputFieldsToSchema()

  const inputSchema = parseJSON(inputSchemaStr.value)
  if (inputSchemaStr.value.trim() && inputSchema === undefined) {
    inputSchemaError.value = '入参 Schema 不是合法的 JSON'
    return
  }
  const outputSchema = parseJSON(outputSchemaStr.value)
  if (outputSchemaStr.value.trim() && outputSchema === undefined) {
    outputSchemaError.value = '出参 Schema 不是合法的 JSON'
    return
  }

  const payload = {
    ...form,
    headers: headersToMap(),
    request_transform_script: requestTransformEnabled.value ? form.request_transform_script : '',
    response_transform_script: responseTransformEnabled.value ? form.response_transform_script : '',
    input_schema: inputSchema || { type: 'object', properties: {} },
    output_schema: outputSchema || { type: 'object', properties: {} },
  }

  try {
    if (editingService.value) {
      await servicesStore.updateService(editingService.value.id, payload)
      showToast('服务已更新', 'success')
    } else {
      await servicesStore.createService(payload)
      showToast('服务已创建', 'success')
    }
    showModal.value = false
  } catch (e) {
    showToast('保存失败: ' + e.message, 'error')
  }
}

const parseJSON = (str) => {
  if (!str || !str.trim()) return null
  try { return JSON.parse(str) } catch { return undefined }
}

// ---- Delete ----
const handleDelete = async (service) => {
  if (!confirm('确定要删除此服务吗？')) return
  try {
    await servicesStore.deleteService(service.id)
    showToast('服务已删除', 'success')
  } catch (e) {
    showToast('删除失败: ' + e.message, 'error')
  }
}

// ---- Debug ----
const openDebugModal = (service) => {
  debugService.value = service
  debugResult.value = null
  debugLoading.value = false
  debugBodyError.value = ''

  // 初始化请求体类型
  debugBodyType.value = service.body_type || 'JSON'

  // Pre-fill headers from service config
  debugHeaders.value = service.headers
    ? Object.entries(service.headers).map(([k, v]) => ({ key: k, value: String(v) }))
    : [{ key: '', value: '' }]

  // 根据请求体类型自动添加 Content-Type（如果 headers 中没有的话）
  const ct = bodyTypeContentTypeMap[debugBodyType.value]
  if (ct && !debugHeaders.value.some(h => h.key.toLowerCase() === 'content-type')) {
    debugHeaders.value.push({ key: 'Content-Type', value: ct })
  }

  // Pre-fill body from input schema (using defaults if available)
  if (service.input_schema && service.input_schema.properties) {
    const sampleBody = {}
    Object.entries(service.input_schema.properties).forEach(([name, prop]) => {
      // Use default value if defined, otherwise use type-based placeholder
      if (prop.default !== undefined && prop.default !== null) {
        sampleBody[name] = prop.default
      } else {
        switch (prop.type) {
          case 'string': sampleBody[name] = ''; break
          case 'integer': sampleBody[name] = 0; break
          case 'number': sampleBody[name] = 0.0; break
          case 'boolean': sampleBody[name] = false; break
          case 'array': sampleBody[name] = []; break
          case 'object': sampleBody[name] = {}; break
          default: sampleBody[name] = ''
        }
      }
    })
    debugBodyStr.value = JSON.stringify(sampleBody, null, 2)

    // 如果是 form-data / urlencoded，也初始化键值对
    if (debugBodyType.value === 'form-data' || debugBodyType.value === 'urlencoded') {
      debugFormFields.value = Object.entries(sampleBody).map(([k, v]) => ({ key: k, value: String(v) }))
    } else {
      debugFormFields.value = [{ key: '', value: '' }]
    }
  } else {
    debugBodyStr.value = '{\n  \n}'
    debugFormFields.value = [{ key: '', value: '' }]
  }

  showDebugModal.value = true
}

const runDebug = async () => {
  debugBodyError.value = ''
  debugResult.value = null

  let body = null
  // body_type 为 none 时不需要请求体
  if (debugBodyType.value === 'none') {
    body = null
  } else if (debugBodyType.value === 'form-data' || debugBodyType.value === 'urlencoded') {
    // form-data / urlencoded: 从键值对构建对象
    const obj = {}
    debugFormFields.value.forEach(f => { if (f.key) obj[f.key] = f.value })
    body = obj
  } else if (debugBodyStr.value.trim()) {
    if (debugBodyType.value === 'JSON') {
      try {
        body = JSON.parse(debugBodyStr.value)
      } catch {
        debugBodyError.value = '请求体不是合法的 JSON'
        return
      }
    } else {
      // raw / binary / msgpack 等作为字符串传递
      body = debugBodyStr.value
    }
  }

  const headers = {}
  debugHeaders.value.forEach(h => { if (h.key && h.value) headers[h.key] = h.value })

  debugLoading.value = true
  try {
    const result = await servicesApi.debugService(debugService.value.id, {
      headers,
      body,
      body_type: debugBodyType.value,
    })
    debugResult.value = result
    if (result.success) {
      showToast('请求成功', 'success')
    } else {
      showToast('请求失败: ' + (result.error || `HTTP ${result.status_code}`), 'error')
    }
  } catch (e) {
    debugResult.value = { success: false, error: e.message || '请求异常' }
    showToast('调试失败: ' + e.message, 'error')
  } finally {
    debugLoading.value = false
  }
}

const formatJSON = (obj) => {
  if (obj === undefined || obj === null) return ''
  if (typeof obj === 'string') return obj
  try { return JSON.stringify(obj, null, 2) } catch { return String(obj) }
}

// 解析请求体 JSON 供 JsonViewer 使用
const parsedDebugBody = computed(() => {
  if (!debugBodyStr.value?.trim()) return null
  try {
    return JSON.parse(debugBodyStr.value)
  } catch {
    return null
  }
})

// 检查响应体是否存在
const hasResponseBody = computed(() => {
  if (!debugResult.value) return false
  const body = debugResult.value.response_body
  if (body === undefined || body === null) return false
  if (typeof body === 'object' && Object.keys(body).length === 0) return false
  return true
})

// 检查响应体是否为对象类型
const isResponseBodyObject = computed(() => {
  if (!debugResult.value) return false
  const body = debugResult.value.response_body
  return body !== null && typeof body === 'object'
})

// 格式化响应体为字符串
const formatResponseBody = computed(() => {
  if (!debugResult.value) return ''
  const body = debugResult.value.response_body
  if (body === null || body === undefined) return ''
  if (typeof body === 'object') {
    try {
      return JSON.stringify(body, null, 2)
    } catch {
      return String(body)
    }
  }
  return String(body)
})

// 格式化 JSON 值
const formatJsonValue = (value) => {
  if (value === null) return 'null'
  if (value === undefined) return 'undefined'
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value, null, 2)
    } catch {
      return String(value)
    }
  }
  if (typeof value === 'string') return `"${value}"`
  return String(value)
}

// 格式化请求体
const formatRequestBody = () => {
  debugBodyError.value = ''
  const str = debugBodyStr.value.trim()
  if (!str) return
  try {
    const parsed = JSON.parse(str)
    debugBodyStr.value = JSON.stringify(parsed, null, 2)
  } catch {
    debugBodyError.value = '请求体不是合法的 JSON'
  }
}

onMounted(() => {
  servicesStore.fetchServices()
})
</script>

<style scoped>
/* VueJsonPretty 自定义样式 */
:deep(.vjs-tree) {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}
:deep(.vjs-tree-node) {
  line-height: 1.5;
}
</style>
