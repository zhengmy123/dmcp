import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { toolsApi } from '@/api/tools'

export const useToolsStore = defineStore('tools', () => {
  const tools = ref([])
  const loading = ref(false)
  const error = ref(null)

  const groupedTools = computed(() => {
    const groups = {}
    tools.value.forEach(tool => {
      const key = tool.vauth_key || 'default'
      if (!groups[key]) {
        groups[key] = []
      }
      groups[key].push(tool)
    })
    return groups
  })

  const serviceCount = computed(() => Object.keys(groupedTools.value).length)

  const fetchTools = async () => {
    loading.value = true
    error.value = null
    try {
      const data = await toolsApi.getTools()
      tools.value = data.tools || []
    } catch (e) {
      error.value = e.message
      tools.value = []
    } finally {
      loading.value = false
    }
  }

  // ========== 工具管理功能 ==========

  // 获取工具列表（管理）
  const fetchToolsForAdmin = async () => {
    loading.value = true
    error.value = null
    try {
      const data = await toolsApi.list()
      tools.value = data.tools || []
    } catch (e) {
      error.value = e.message
      tools.value = []
    } finally {
      loading.value = false
    }
  }

  // 创建工具
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

  // 更新工具
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

  // 删除工具
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
    groupedTools,
    serviceCount,
    fetchTools,
    fetchToolsForAdmin,
    createTool,
    updateTool,
    deleteTool
  }
})
