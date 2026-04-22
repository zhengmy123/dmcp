<template>
  <div class="min-h-screen flex">
    <!-- Left Panel - Branding -->
    <div class="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-primary-600 via-primary-700 to-primary-800 relative overflow-hidden">
      <!-- Decorative circles -->
      <div class="absolute top-20 left-20 w-72 h-72 bg-white/10 rounded-full blur-3xl"></div>
      <div class="absolute bottom-20 right-10 w-96 h-96 bg-primary-400/20 rounded-full blur-3xl"></div>
      <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] border border-white/10 rounded-full"></div>
      <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[400px] h-[400px] border border-white/10 rounded-full"></div>
      <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[300px] h-[300px] border border-white/10 rounded-full"></div>

      <!-- Content -->
      <div class="relative z-10 flex flex-col justify-center px-16 text-white">
        <div class="flex items-center gap-4 mb-8">
          <div class="w-16 h-16 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center">
            <svg class="w-10 h-10 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"/>
            </svg>
          </div>
          <h1 class="text-4xl font-bold">DMCP</h1>
        </div>
        <p class="text-xl text-white/80 mb-12 leading-relaxed">
          Dynamic MCP Server<br/>
          高效管理 MCP 工具与服务
        </p>

        <!-- Feature list -->
        <div class="space-y-4">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
              </svg>
            </div>
            <span class="text-white/90">灵活的 MCP Server 配置</span>
          </div>
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
              </svg>
            </div>
            <span class="text-white/90">工具定义与绑定管理</span>
          </div>
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
              </svg>
            </div>
            <span class="text-white/90">HTTP 服务集成</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Right Panel - Login Form -->
    <div class="flex-1 flex items-center justify-center p-8 bg-gray-50">
      <div class="w-full max-w-md">
        <!-- Mobile Logo -->
        <div class="lg:hidden flex items-center justify-center gap-3 mb-8">
          <div class="w-12 h-12 bg-primary-600 rounded-xl flex items-center justify-center">
            <svg class="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"/>
            </svg>
          </div>
          <h1 class="text-2xl font-bold text-gray-900">DMCP</h1>
        </div>

        <!-- Form Card -->
        <div class="bg-white rounded-2xl shadow-xl p-8">
          <div class="text-center mb-8">
            <h2 class="text-2xl font-bold text-gray-900 mb-2">欢迎回来</h2>
            <p class="text-gray-500">请登录到管理后台</p>
          </div>

          <form @submit.prevent="handleLogin" class="space-y-5">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">用户名</label>
              <div class="relative">
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
                  </svg>
                </span>
                <input
                  v-model="form.username"
                  type="text"
                  placeholder="请输入用户名"
                  class="w-full pl-12 pr-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all bg-gray-50 focus:bg-white"
                  :disabled="loading"
                />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">密码</label>
              <div class="relative">
                <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
                  </svg>
                </span>
                <input
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  placeholder="请输入密码"
                  class="w-full pl-12 pr-12 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition-all bg-gray-50 focus:bg-white"
                  :disabled="loading"
                />
                <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 transition-colors"
                >
                  <svg v-if="showPassword" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/>
                  </svg>
                  <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                  </svg>
                </button>
              </div>
            </div>

            <!-- Error Alert -->
            <div v-if="error" class="bg-red-50 border border-red-100 text-red-600 px-4 py-3 rounded-xl text-sm flex items-center gap-2">
              <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              {{ error }}
            </div>

            <!-- Login Button -->
            <button
              type="submit"
              :disabled="loading || !form.username || !form.password"
              class="w-full py-3 px-4 bg-primary-600 text-white font-medium rounded-xl hover:bg-primary-700 focus:ring-4 focus:ring-primary-100 disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary-600/30"
            >
              <svg v-if="loading" class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span>{{ loading ? '登录中...' : '登 录' }}</span>
            </button>
          </form>
        </div>

        <!-- Footer -->
        <p class="text-center text-sm text-gray-400 mt-8">
          © 2024 DMCP. All rights reserved.
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'

const router = useRouter()
const authStore = useAuthStore()

const form = ref({
  username: '',
  password: ''
})

const loading = ref(false)
const error = ref('')
const showPassword = ref(false)

const handleLogin = async () => {
  error.value = ''
  loading.value = true

  try {
    const response = await authApi.login(form.value.username, form.value.password)

    authStore.login(response.data.token, response.data.user)

    router.push('/dashboard')
  } catch (e) {
    error.value = e.message || e.response?.data?.message || e.response?.data?.error || '登录失败，请检查用户名和密码'
  } finally {
    loading.value = false
  }
}
</script>
