import { defineStore } from 'pinia'
import { ref } from 'vue'
import { servicesApi } from '@/api/services'

export const useServicesStore = defineStore('services', () => {
  const services = ref([])
  const loading = ref(false)
  const error = ref(null)

  const fetchServices = async () => {
    loading.value = true
    error.value = null
    try {
      const data = await servicesApi.getServices()
      services.value = data.services || []
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
