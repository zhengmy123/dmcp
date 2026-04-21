<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold text-gray-900">Token 管理</h2>
        <p class="text-sm text-gray-500 mt-1">管理 MCP Server 的访问凭证</p>
      </div>
      <button
        @click="showCreateModal = true"
        class="inline-flex items-center px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 btn-transition"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
        </svg>
        创建 Token
      </button>
    </div>

    <!-- Search -->
    <div class="bg-white rounded-xl border border-gray-200 p-4">
      <div class="relative max-w-md">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="搜索 Key ID、Token 或名称..."
          class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
        >
        <svg class="w-5 h-5 text-gray-400 absolute left-3 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
        </svg>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="bg-gray-50 border-b border-gray-200">
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Key ID</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Token</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">名称</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">创建时间</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-if="tokenStore.loading">
              <td colspan="7" class="px-6 py-12 text-center">
                <div class="loading-spinner mx-auto"></div>
                <p class="text-gray-500 mt-2">加载中...</p>
              </td>
            </tr>
            <tr v-else-if="filteredTokens.length === 0">
              <td colspan="7" class="px-6 py-12 text-center text-gray-500">
                暂无数据
              </td>
            </tr>
            <tr
              v-else
              v-for="token in filteredTokens"
              :key="token.id"
              class="table-row-hover"
            >
              <td class="px-6 py-4">
                <span class="font-mono text-sm font-medium text-gray-900">{{ token.key_id }}</span>
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center space-x-2">
                  <button
                    @click="copyText(token.token, token.token)"
                    class="p-1 rounded transition-colors"
                    :class="copiedTokenId === token.token ? 'text-green-600' : 'text-gray-400 hover:text-primary-600 hover:bg-primary-50'"
                    title="复制 Token"
                  >
                    <svg v-if="copiedTokenId !== token.token" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                    </svg>
                    <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                    </svg>
                  </button>
                  <span class="font-mono text-sm text-gray-600 truncate max-w-[200px]">{{ token.token }}</span>
                </div>
              </td>
              <td class="px-6 py-4">
                <span class="text-sm text-gray-900">{{ token.name || '-' }}</span>
              </td>
              <td class="px-6 py-4">
                <span
                  class="px-2.5 py-1 text-xs font-medium rounded-full"
                  :class="token.state === 1 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'"
                >
                  {{ token.state === 1 ? '启用' : '禁用' }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">
                {{ formatDate(token.created_at) }}
              </td>
              <td class="px-6 py-4 text-right">
                <div class="flex items-center justify-end space-x-2">
                  <button
                    @click="handleRefresh(token)"
                    class="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                    title="刷新 Token"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
                    </svg>
                  </button>
                  <button
                    @click="handleToggle(token)"
                    class="px-3 py-1 text-xs font-medium rounded-lg transition-colors"
                    :class="token.state === 1 
                      ? 'bg-red-100 text-red-700 hover:bg-red-200' 
                      : 'bg-green-100 text-green-700 hover:bg-green-200'"
                    :title="token.state === 1 ? '禁用' : '启用'"
                  >
                    {{ token.state === 1 ? '禁用' : '启用' }}
                  </button>
                  <button
                    @click="handleDelete(token)"
                    class="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                    title="删除"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="tokenStore.pagination.total > 0" class="px-6 py-4 border-t border-gray-200 flex items-center justify-between">
      <div class="text-sm text-gray-500">
        共 <span class="font-medium">{{ tokenStore.pagination.total }}</span> 条记录，第
        <span class="font-medium">{{ tokenStore.pagination.page }}</span> /
        <span class="font-medium">{{ totalPages }}</span> 页
      </div>
      <div class="flex items-center space-x-2">
        <button
          @click="goToPage(tokenStore.pagination.page - 1)"
          :disabled="tokenStore.pagination.page <= 1"
          class="px-3 py-1.5 text-sm border border-gray-300 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 transition-colors"
        >
          上一页
        </button>
        <div class="flex items-center space-x-1">
          <button
            v-for="page in visiblePages"
            :key="page"
            @click="page !== '...' && goToPage(page)"
            class="w-8 h-8 text-sm rounded-lg transition-colors"
            :class="page === tokenStore.pagination.page
              ? 'bg-primary-600 text-white'
              : 'border border-gray-300 hover:bg-gray-50'"
            :disabled="page === '...'"
          >
            {{ page }}
          </button>
        </div>
        <button
          @click="goToPage(tokenStore.pagination.page + 1)"
          :disabled="tokenStore.pagination.page >= totalPages"
          class="px-3 py-1.5 text-sm border border-gray-300 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 transition-colors"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- Create Modal -->
    <teleport to="body">
      <transition name="fade">
        <div v-if="showCreateModal" class="fixed inset-0 z-50 overflow-y-auto">
          <div class="flex items-center justify-center min-h-screen px-4">
            <div class="fixed inset-0 bg-black bg-opacity-50" @click="showCreateModal = false"></div>
            <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md p-6 fade-in">
              <div class="flex items-center justify-between mb-6">
                <h3 class="text-lg font-semibold text-gray-900">创建新 Token</h3>
                <button @click="showCreateModal = false" class="text-gray-400 hover:text-gray-600">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
              <form @submit.prevent="handleCreate">
                <div class="space-y-4">
                  <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">名称</label>
                    <input
                      v-model="createForm.name"
                      type="text"
                      class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                      placeholder="例如: 我的应用密钥"
                    >
                  </div>
                  <div class="bg-blue-50 rounded-lg p-3 text-sm text-blue-700">
                    <p>Key ID、Token 和 Secret 将在创建时自动生成。</p>
                  </div>
                </div>
                <div class="flex justify-end space-x-3 mt-6">
                  <button
                    type="button"
                    @click="showCreateModal = false"
                    class="px-4 py-2 border border-gray-300 text-gray-700 text-sm font-medium rounded-lg hover:bg-gray-50"
                  >
                    取消
                  </button>
                  <button
                    type="submit"
                    :disabled="tokenStore.loading"
                    class="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 disabled:opacity-50"
                  >
                    {{ tokenStore.loading ? '创建中...' : '创建' }}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      </transition>
    </teleport>

    <!-- Result Modal -->
    <teleport to="body">
      <transition name="fade">
        <div v-if="showResultModal" class="fixed inset-0 z-50 overflow-y-auto">
          <div class="flex items-center justify-center min-h-screen px-4">
            <div class="fixed inset-0 bg-black bg-opacity-50" @click="showResultModal = false"></div>
            <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md p-6 fade-in">
              <div class="text-center">
                <div class="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                  <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                  </svg>
                </div>
                <h3 class="text-lg font-semibold text-gray-900 mb-2">Token 创建成功</h3>
                <p class="text-sm text-gray-500 mb-4">请妥善保管以下信息，关闭后将无法再次查看</p>
                <div class="bg-gray-50 rounded-lg p-4 text-left space-y-3">
                  <div>
                    <label class="block text-xs font-medium text-gray-500 mb-1">Token</label>
                    <div class="flex items-center space-x-2">
                      <code class="flex-1 text-sm text-gray-900 bg-white px-2 py-1 border rounded break-all">{{ newToken.token }}</code>
                      <button @click="copyText(newToken.token, 'token')" class="p-1 rounded transition-colors" :class="copiedTokenId === 'token' ? 'text-green-600' : 'text-gray-400 hover:text-primary-600 hover:bg-primary-50'">
                        <svg v-if="copiedTokenId !== 'token'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                        </svg>
                        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                        </svg>
                      </button>
                    </div>
                  </div>
                  <div>
                    <label class="block text-xs font-medium text-gray-500 mb-1">Secret</label>
                    <div class="flex items-center space-x-2">
                      <code class="flex-1 text-sm text-gray-900 bg-white px-2 py-1 border rounded break-all">{{ newToken.secret }}</code>
                      <button @click="copyText(newToken.secret, 'secret', 'secret')" class="p-1 rounded transition-colors" :class="copiedSecretId === 'secret' ? 'text-green-600' : 'text-gray-400 hover:text-primary-600 hover:bg-primary-50'">
                        <svg v-if="copiedSecretId !== 'secret'" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                        </svg>
                        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>
                <button
                  @click="showResultModal = false"
                  class="mt-6 w-full px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700"
                >
                  我已保存
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
import { ref, computed, inject, onMounted } from 'vue'
import { useTokenStore } from '@/stores/tokens'
import dayjs from 'dayjs'

const tokenStore = useTokenStore()
const showToast = inject('showToast')

const searchQuery = ref('')
const showCreateModal = ref(false)
const showResultModal = ref(false)
const newToken = ref({ token: '', secret: '' })
const createForm = ref({ name: '' })
const copiedTokenId = ref(null)
const copiedSecretId = ref('secret')

const filteredTokens = computed(() => {
  if (!searchQuery.value) return tokenStore.tokens
  const q = searchQuery.value.toLowerCase()
  return tokenStore.tokens.filter(t =>
    (t.key_id || '').toLowerCase().includes(q) ||
    (t.token || '').toLowerCase().includes(q) ||
    (t.name || '').toLowerCase().includes(q)
  )
})

const formatDate = (date) => {
  if (!date) return '-'
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

const handleCreate = async () => {
  try {
    const result = await tokenStore.createToken(createForm.value)
    newToken.value = { token: result.token, secret: result.secret }
    showResultModal.value = true
    showCreateModal.value = false
    createForm.value = { name: '' }
    showToast('Token 创建成功', 'success')
  } catch (e) {
    showToast('创建失败: ' + e.message, 'error')
  }
}

const handleRefresh = async (token) => {
  if (!confirm('确定要刷新此 Token 吗？刷新后将生成新的 Token 和 Secret。')) return
  try {
    const result = await tokenStore.refreshToken(token.token)
    newToken.value = { token: result.new_token, secret: result.new_secret }
    showResultModal.value = true
    showToast('Token 已刷新', 'success')
  } catch (e) {
    showToast('刷新失败: ' + e.message, 'error')
  }
}

const handleToggle = async (token) => {
  try {
    await tokenStore.toggleToken(token.token, token.state !== 1)
    showToast(`Token 已${token.state === 1 ? '禁用' : '启用'}`, 'success')
  } catch (e) {
    showToast('操作失败: ' + e.message, 'error')
  }
}

const handleDelete = async (token) => {
  if (!confirm('确定要删除此 Token 吗？此操作不可撤销。')) return
  try {
    await tokenStore.deleteToken(token.token)
    showToast('Token 已删除', 'success')
  } catch (e) {
    showToast('删除失败: ' + e.message, 'error')
  }
}

const copyText = async (text, tokenId, secretId = null) => {
  try {
    await navigator.clipboard.writeText(text)
    if (secretId) {
      copiedSecretId.value = secretId
    } else {
      copiedTokenId.value = tokenId
    }
    setTimeout(() => {
      if (secretId) {
        copiedSecretId.value = null
      } else {
        copiedTokenId.value = null
      }
    }, 1500)
  } catch (e) {
    console.error('复制失败:', e)
  }
}

onMounted(() => {
  tokenStore.fetchTokens({
    page: 1,
    pageSize: 10
  })
})

const handlePageSizeChange = () => {
  tokenStore.fetchTokens({
    page: 1,
    pageSize: tokenStore.pagination.pageSize
  })
}

const goToPage = (page) => {
  tokenStore.fetchTokens({
    page: page
  })
}

const totalPages = computed(() => {
  const total = tokenStore.pagination.total
  const size = tokenStore.pagination.pageSize
  return Math.ceil(total / size) || 1
})

const visiblePages = computed(() => {
  const current = tokenStore.pagination.page
  const total = totalPages.value
  const pages = []

  if (total <= 7) {
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    if (current <= 4) {
      for (let i = 1; i <= 5; i++) pages.push(i)
      pages.push('...')
      pages.push(total)
    } else if (current >= total - 3) {
      pages.push(1)
      pages.push('...')
      for (let i = total - 4; i <= total; i++) pages.push(i)
    } else {
      pages.push(1)
      pages.push('...')
      for (let i = current - 1; i <= current + 1; i++) pages.push(i)
      pages.push('...')
      pages.push(total)
    }
  }
  return pages
})
</script>
