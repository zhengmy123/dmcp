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

  return {
    tools,
    loading,
    error,
    groupedTools,
    serviceCount,
    fetchTools
  }
})
