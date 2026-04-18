import { defineStore } from 'pinia'
import { ref } from 'vue'
import { toolBindingsApi } from '@/api/toolBindings'

export const useToolBindingsStore = defineStore('toolBindings', () => {
  const bindings = ref([])
  const loading = ref(false)
  const error = ref(null)

  const getToolBindings = async (toolId) => {
    loading.value = true
    error.value = null
    try {
      const res = await toolBindingsApi.getToolBindings(toolId)
      const data = res.data || res
      bindings.value = data.bindings || []
      return bindings.value
    } catch (e) {
      error.value = e.message
      return []
    } finally {
      loading.value = false
    }
  }

  const bindTool = async (toolId, serverId) => {
    loading.value = true
    error.value = null
    try {
      await toolBindingsApi.bindTool({ tool_id: toolId, server_id: serverId })
      await getToolBindings(toolId)
      return true
    } catch (e) {
      error.value = e.message
      return false
    } finally {
      loading.value = false
    }
  }

  const unbindTool = async (toolId, serverId) => {
    loading.value = true
    error.value = null
    try {
      await toolBindingsApi.unbindTool(toolId, serverId)
      await getToolBindings(toolId)
      return true
    } catch (e) {
      error.value = e.message
      return false
    } finally {
      loading.value = false
    }
  }

  const batchBind = async (toolIds, serverIds) => {
    loading.value = true
    error.value = null
    try {
      const res = await toolBindingsApi.batchBind({
        tool_ids: toolIds,
        server_ids: serverIds
      })
      return res.data || res
    } catch (e) {
      error.value = e.message
      return null
    } finally {
      loading.value = false
    }
  }

  const batchUnbind = async (bindingIds) => {
    loading.value = true
    error.value = null
    try {
      const res = await toolBindingsApi.batchUnbind(bindingIds)
      return res.data || res
    } catch (e) {
      error.value = e.message
      return null
    } finally {
      loading.value = false
    }
  }

  const getServerBindings = async (serverId) => {
    loading.value = true
    error.value = null
    try {
      const res = await toolBindingsApi.getServerBindings(serverId)
      const data = res.data || res
      bindings.value = data.bindings || []
      return bindings.value
    } catch (e) {
      error.value = e.message
      return []
    } finally {
      loading.value = false
    }
  }

  return {
    bindings,
    loading,
    error,
    getToolBindings,
    bindTool,
    unbindTool,
    batchBind,
    batchUnbind,
    getServerBindings
  }
})