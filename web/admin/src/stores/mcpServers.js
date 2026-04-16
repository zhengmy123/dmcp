import { defineStore } from 'pinia'
import { ref } from 'vue'
import { mcpServerAPI } from '@/api/mcpServers'

export const useMCPServersStore = defineStore('mcpServers', () => {
  const servers = ref([])
  const currentServer = ref(null)
  const loading = ref(false)
  const error = ref(null)

  // 获取所有 Server
  const fetchServers = async () => {
    loading.value = true
    error.value = null
    try {
      const data = await mcpServerAPI.list()
      servers.value = data.servers || []
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
      const data = await mcpServerAPI.get(id)
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
    addToolsToServer,
    removeToolFromServer
  }
})
