import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { JWT_TOKEN_KEY, USER_INFO_KEY } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const jwtToken = ref(localStorage.getItem(JWT_TOKEN_KEY) || '')
  const userInfo = ref(JSON.parse(localStorage.getItem(USER_INFO_KEY) || 'null'))
  const isAuthenticated = computed(() => !!jwtToken.value)

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

  return {
    jwtToken,
    userInfo,
    isAuthenticated,
    login,
    logout
  }
})
