import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { SERVER_URL_KEY, JWT_TOKEN_KEY, USER_INFO_KEY } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  // JWT token
  const jwtToken = ref(localStorage.getItem(JWT_TOKEN_KEY) || '')
  // 服务器地址
  const serverUrl = ref(localStorage.getItem(SERVER_URL_KEY) || 'http://localhost:18080')
  // 用户信息
  const userInfo = ref(JSON.parse(localStorage.getItem(USER_INFO_KEY) || 'null'))
  // 是否已登录
  const isAuthenticated = computed(() => !!jwtToken.value)

  const saveSettings = () => {
    localStorage.setItem(SERVER_URL_KEY, serverUrl.value)
  }

  const login = (token, user) => {
    jwtToken.value = token
    userInfo.value = user
    localStorage.setItem(JWT_TOKEN_KEY, token)
    localStorage.setItem(USER_INFO_KEY, JSON.stringify(user))
  }

  const logout = () => {
    jwtToken.value = ''
    userInfo.value = null
    localStorage.removeItem(JWT_TOKEN_KEY)
    localStorage.removeItem(USER_INFO_KEY)
  }

  const updateServerUrl = (url) => {
    serverUrl.value = url
    saveSettings()
  }

  return {
    jwtToken,
    serverUrl,
    userInfo,
    isAuthenticated,
    login,
    logout,
    saveSettings,
    updateServerUrl
  }
})
