import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { toolsApi } from '@/api/tools'

export const useToolsStore = defineStore('tools', () => {
  const tools = ref([])
  const loading = ref(false)
  const error = ref(null)

  const pagination = ref({
    page: 1,
    pageSize: 20,
    total: 0
  })

  const searchKeyword = ref('')

  const fetchTools = async () => {
    loading.value = true
    error.value = null
    try {
      const res = await toolsApi.getTools()
      const data = res.data || res
      tools.value = data.tools || []
    } catch (e) {
      error.value = e.message
      tools.value = []
    } finally {
      loading.value = false
    }
  }

  const fetchToolsForAdmin = async (params = {}) => {
    loading.value = true
    error.value = null
    try {
      const queryParams = {
        page: params.page || pagination.value.page,
        page_size: params.page_size || pagination.value.pageSize,
        keyword: params.keyword !== undefined ? params.keyword : searchKeyword.value
      }
      const res = await toolsApi.list(queryParams)
      const data = res.data || res
      const items = data.tools || data.items || []
      tools.value = items
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
      tools.value = []
    } finally {
      loading.value = false
    }
  }

  const setPage = (page) => {
    pagination.value.page = page
    fetchToolsForAdmin({ page })
  }

  const setKeyword = (keyword) => {
    searchKeyword.value = keyword
    pagination.value.page = 1
    fetchToolsForAdmin({ keyword, page: 1 })
  }

  const createTool = async (data) => {
    loading.value = true
    error.value = null
    try {
      await toolsApi.create(data)
      await fetchToolsForAdmin()
      return true
    } catch (e) {
      error.value = e.message
      return false
    } finally {
      loading.value = false
    }
  }

  const updateTool = async (id, data) => {
    loading.value = true
    error.value = null
    try {
      await toolsApi.update(id, data)
      await fetchToolsForAdmin()
      return true
    } catch (e) {
      error.value = e.message
      return false
    } finally {
      loading.value = false
    }
  }

  const deleteTool = async (id) => {
    loading.value = true
    error.value = null
    try {
      await toolsApi.delete(id)
      await fetchToolsForAdmin()
      return true
    } catch (e) {
      error.value = e.message
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    tools,
    loading,
    error,
    pagination,
    searchKeyword,
    fetchTools,
    fetchToolsForAdmin,
    setPage,
    setKeyword,
    createTool,
    updateTool,
    deleteTool
  }
})