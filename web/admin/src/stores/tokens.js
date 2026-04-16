import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '@/api/auth'

export const useTokenStore = defineStore('tokens', () => {
  const tokens = ref([])
  const loading = ref(false)
  const error = ref(null)

  const fetchTokens = async () => {
    loading.value = true
    error.value = null
    try {
      const data = await authApi.getTokens()
      tokens.value = data.tokens || []
    } catch (e) {
      error.value = e.message
      tokens.value = []
    } finally {
      loading.value = false
    }
  }

  const createToken = async (params) => {
    loading.value = true
    error.value = null
    try {
      const result = await authApi.createToken(params)
      await fetchTokens()
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const deleteToken = async (token) => {
    loading.value = true
    error.value = null
    try {
      await authApi.deleteToken(token)
      await fetchTokens()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const refreshToken = async (token) => {
    loading.value = true
    error.value = null
    try {
      const result = await authApi.refreshToken(token)
      await fetchTokens()
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const toggleToken = async (token, enable) => {
    loading.value = true
    error.value = null
    try {
      if (enable) {
        await authApi.enableToken(token)
      } else {
        await authApi.disableToken(token)
      }
      await fetchTokens()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    tokens,
    loading,
    error,
    fetchTokens,
    createToken,
    deleteToken,
    refreshToken,
    toggleToken
  }
})
