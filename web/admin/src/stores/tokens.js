import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '@/api/auth'

export const useTokenStore = defineStore('tokens', () => {
  const tokens = ref([])
  const loading = ref(false)
  const error = ref(null)

  const pagination = ref({
    page: 1,
    pageSize: 20,
    total: 0
  })

  const fetchTokens = async () => {
    loading.value = true
    error.value = null
    try {
      const res = await authApi.getTokens()
      const data = res.data || res
      const items = data.items || data.tokens || []
      tokens.value = items.map(item => ({
        ...item,
        enabled: item.state === 1
      }))
      if (data.total !== undefined) {
        pagination.value.total = data.total
      }
      if (data.page !== undefined) {
        pagination.value.page = data.page
      }
      if (data.page_size !== undefined) {
        pagination.value.pageSize = data.page_size
      }
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
