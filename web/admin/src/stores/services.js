import { defineStore } from 'pinia'
import { ref } from 'vue'
import { servicesApi } from '@/api/services'

export const useServicesStore = defineStore('services', () => {
  const services = ref([])
  const loading = ref(false)
  const error = ref(null)

  const pagination = ref({
    page: 1,
    pageSize: 20,
    total: 0
  })

  const fetchServices = async () => {
    loading.value = true
    error.value = null
    try {
      const res = await servicesApi.getServices()
      const data = res.data || res
      const items = data.services || data.items || []
      services.value = items
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
      services.value = []
    } finally {
      loading.value = false
    }
  }

  const createService = async (params) => {
    loading.value = true
    error.value = null
    try {
      const result = await servicesApi.createService(params)
      await fetchServices()
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const updateService = async (id, params) => {
    loading.value = true
    error.value = null
    try {
      const result = await servicesApi.updateService(id, params)
      await fetchServices()
      return result
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const deleteService = async (id) => {
    loading.value = true
    error.value = null
    try {
      await servicesApi.deleteService(id)
      await fetchServices()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    services,
    loading,
    error,
    fetchServices,
    createService,
    updateService,
    deleteService
  }
})
