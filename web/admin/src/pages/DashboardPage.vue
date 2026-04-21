<template>
  <div class="space-y-6">
    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <div
        v-for="(stat, index) in stats"
        :key="index"
        class="bg-white rounded-xl p-6 border border-gray-200 card-hover cursor-pointer"
      >
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-gray-500">{{ stat.title }}</p>
            <p class="text-3xl font-bold text-gray-900 mt-1">{{ stat.value }}</p>
            <p class="text-sm text-gray-500 mt-1">{{ stat.subtitle }}</p>
          </div>
          <div
            class="w-12 h-12 rounded-xl flex items-center justify-center"
            :class="stat.bgClass"
          >
            <component :is="stat.icon" :class="stat.iconClass" />
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Recent Tokens -->
      <div class="lg:col-span-2 bg-white rounded-xl border border-gray-200">
        <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <h3 class="font-semibold text-gray-900">最近 Token</h3>
          <router-link to="/tokens" class="text-sm text-primary-600 hover:text-primary-700">
            查看全部 →
          </router-link>
        </div>
        <div class="p-6">
          <div v-if="tokenStore.loading" class="text-center py-8">
            <div class="loading-spinner mx-auto"></div>
            <p class="text-gray-500 mt-2">加载中...</p>
          </div>
          <div v-else-if="tokenStore.tokens.length === 0" class="text-center py-8 text-gray-500">
            暂无 Token
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="token in tokenStore.tokens.slice(0, 5)"
              :key="token.id"
              class="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
            >
              <div class="flex items-center space-x-3">
                <div class="w-10 h-10 bg-primary-100 rounded-lg flex items-center justify-center">
                  <svg class="w-5 h-5 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"/>
                  </svg>
                </div>
                <div>
                  <p class="font-medium text-gray-900">{{ token.name || token.key_id }}</p>
                  <p class="text-xs text-gray-500 font-mono">{{ token.key_id }}</p>
                </div>
              </div>
              <span
                class="px-2.5 py-1 text-xs font-medium rounded-full"
                :class="token.state === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'"
              >
                {{ token.state === 1 ? '启用' : '禁用' }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Service Groups -->
      <div class="bg-white rounded-xl border border-gray-200">
        <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <h3 class="font-semibold text-gray-900">服务分组</h3>
          <router-link to="/tools" class="text-sm text-primary-600 hover:text-primary-700">
            查看全部 →
          </router-link>
        </div>
        <div class="p-6">
          <div v-if="toolsStore.loading" class="text-center py-8">
            <div class="loading-spinner mx-auto"></div>
            <p class="text-gray-500 mt-2">加载中...</p>
          </div>
          <div v-else-if="Object.keys(toolsStore.groupedTools).length === 0" class="text-center py-8 text-gray-500">
            暂无服务分组
          </div>
          <div v-else class="space-y-3">
            <div
              v-for="(tools, group) in toolsStore.groupedTools"
              :key="group"
              class="p-3 bg-gray-50 rounded-lg"
            >
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900">{{ group }}</span>
                <span class="text-xs text-gray-500">{{ tools.length }} 个工具</span>
              </div>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="tool in tools.slice(0, 3)"
                  :key="tool.name"
                  class="px-2 py-0.5 text-xs bg-primary-50 text-primary-700 rounded"
                >
                  {{ tool.name }}
                </span>
                <span v-if="tools.length > 3" class="text-xs text-gray-500">
                  +{{ tools.length - 3 }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="bg-white rounded-xl border border-gray-200">
      <div class="px-6 py-4 border-b border-gray-200">
        <h3 class="font-semibold text-gray-900">快捷操作</h3>
      </div>
      <div class="p-6 grid grid-cols-1 md:grid-cols-3 gap-4">
        <button
          @click="$router.push('/tokens')"
          class="flex items-center p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors"
        >
          <div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mr-4">
            <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
            </svg>
          </div>
          <div class="text-left">
            <p class="font-medium text-gray-900">创建 Token</p>
            <p class="text-sm text-gray-500">生成新的访问凭证</p>
          </div>
        </button>
        <button
          @click="$router.push('/services')"
          class="flex items-center p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors"
        >
          <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center mr-4">
            <svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
            </svg>
          </div>
          <div class="text-left">
            <p class="font-medium text-gray-900">管理服务</p>
            <p class="text-sm text-gray-500">配置 HTTP 服务</p>
          </div>
        </button>
        <button
          @click="$router.push('/settings')"
          class="flex items-center p-4 bg-gray-50 rounded-xl hover:bg-gray-100 transition-colors"
        >
          <div class="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center mr-4">
            <svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
            </svg>
          </div>
          <div class="text-left">
            <p class="font-medium text-gray-900">系统设置</p>
            <p class="text-sm text-gray-500">调整配置参数</p>
          </div>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, h, onMounted } from 'vue'
import { useTokenStore } from '@/stores/tokens'
import { useToolsStore } from '@/stores/tools'

const tokenStore = useTokenStore()
const toolsStore = useToolsStore()

const stats = computed(() => [
  {
    title: 'Token 总数',
    value: tokenStore.tokens.length,
    subtitle: '访问凭证',
    bgClass: 'bg-blue-100',
    iconClass: 'w-6 h-6 text-blue-600',
    icon: h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z' })
    ])
  },
  {
    title: '活跃 Token',
    value: tokenStore.tokens.filter(t => t.state === 1).length,
    subtitle: '正在使用',
    bgClass: 'bg-green-100',
    iconClass: 'w-6 h-6 text-green-600',
    icon: h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z' })
    ])
  },
  {
    title: '工具定义',
    value: toolsStore.tools.length,
    subtitle: 'MCP 工具',
    bgClass: 'bg-purple-100',
    iconClass: 'w-6 h-6 text-purple-600',
    icon: h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z' }),
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M15 12a3 3 0 11-6 0 3 3 0 016 0z' })
    ])
  },
  {
    title: '服务分组',
    value: toolsStore.serviceCount,
    subtitle: 'MCP Server',
    bgClass: 'bg-orange-100',
    iconClass: 'w-6 h-6 text-orange-600',
    icon: h('svg', { fill: 'none', stroke: 'currentColor', viewBox: '0 0 24 24' }, [
      h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10' })
    ])
  }
])

onMounted(() => {
  tokenStore.fetchTokens()
  toolsStore.fetchTools()
})
</script>
