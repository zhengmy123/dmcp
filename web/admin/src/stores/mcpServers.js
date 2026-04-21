import { defineStore } from 'pinia'
import { ref } from 'vue'
import { mcpServerAPI } from '@/api/mcpServers'

export const useMCPServersStore = defineStore('mcpServers', () => {
  const servers = ref([])
  const currentServer = ref(null)
  const loading = ref(false)
  const error = ref(null)

  const pagination = ref({
    page: 1,
    pageSize: 20,
    total: 0
  })

  // 获取所有 Server
  const fetchServers = async (params = {}) => {
    loading.value = true
    error.value = null
    try {
      const res = await mcpServerAPI.list(params)
      const data = res.data || res
      const items = data.servers || data.items || []
      servers.value = items
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
      servers.value = []
    } finally {
      loading.value = false
    }
  }

  // 获取单个 Server
  const fetchServer = async (id) => {
    loading.value = true
    error.value = null
    try {
      const res = await mcpServerAPI.get(id)
      const data = res.data || res
      currentServer.value = data.server
      return data.server
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 创建 Server
  const createServer = async (params) => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpServerAPI.create(params)
      await fetchServers()
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 更新 Server
  const updateServer = async (id, params) => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpServerAPI.update(id, params)
      await fetchServers()
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 删除 Server
  const deleteServer = async (id) => {
    loading.value = true
    error.value = null
    try {
      await mcpServerAPI.delete(id)
      await fetchServers()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 恢复 Server
  const restoreServer = async (id) => {
    loading.value = true
    error.value = null
    try {
      await mcpServerAPI.restore(id)
      await fetchServers()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 添加工具到 Server
  const addToolsToServer = async (id, toolNames) => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpServerAPI.addTools(id, toolNames)
      await fetchServers()
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 从 Server 移除工具
  const removeToolFromServer = async (id, toolName) => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpServerAPI.removeTool(id, toolName)
      await fetchServers()
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  // 同步构建信息
  const syncBuild = async (id) => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpServerAPI.syncBuild(id)
      await fetchServers()
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    servers,
    currentServer,
    loading,
    error,
    fetchServers,
    fetchServer,
    createServer,
    updateServer,
    deleteServer,
    restoreServer,
    addToolsToServer,
    removeToolFromServer,
    syncBuild
  }
})
