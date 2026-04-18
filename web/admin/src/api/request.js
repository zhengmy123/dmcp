import axios from 'axios'
import { useAuthStore } from '@/stores/auth'
import { API_BASE_URL, SERVER_URL_KEY, JWT_TOKEN_KEY } from '@/types'

const getBaseURL = () => {
  return localStorage.getItem(SERVER_URL_KEY) || API_BASE_URL
}

const getJWTToken = () => {
  return localStorage.getItem(JWT_TOKEN_KEY) || ''
}

const createRequest = () => {
  const instance = axios.create({
    baseURL: getBaseURL(),
    timeout: 30000
  })

  instance.interceptors.request.use(
    (config) => {
      const jwtToken = getJWTToken()
      if (jwtToken) {
        config.headers['Authorization'] = `Bearer ${jwtToken}`
      }
      return config
    },
    (error) => Promise.reject(error)
  )

  instance.interceptors.response.use(
    (response) => {
      const resData = response.data
      if (resData && typeof resData === 'object' && 'code' in resData) {
        if (resData.code === 0) {
          return resData
        }
        if (resData.code === 401) {
          const authStore = useAuthStore()
          authStore.logout()
          if (!window.location.pathname.includes('/login')) {
            window.location.href = '/login'
          }
          window.dispatchEvent(new CustomEvent('show-toast', {
            detail: { message: resData.message || 'unauthorized', type: 'error' }
          }))
          return Promise.reject(new Error(resData.message || 'unauthorized'))
        }
        if (resData.message) {
          window.dispatchEvent(new CustomEvent('show-toast', {
            detail: { message: resData.message, type: 'error' }
          }))
          return Promise.reject(new Error(resData.message))
        }
        return Promise.reject(new Error('request failed'))
      }
      return response.data
    },
    (error) => {
      if (error.response?.status === 401) {
        const authStore = useAuthStore()
        authStore.logout()
        if (!window.location.pathname.includes('/login')) {
          window.location.href = '/login'
        }
        window.dispatchEvent(new CustomEvent('show-toast', {
          detail: { message: error.response?.data?.message || 'unauthorized', type: 'error' }
        }))
      } else if (error.message) {
        window.dispatchEvent(new CustomEvent('show-toast', {
          detail: { message: error.message, type: 'error' }
        }))
      }
      return Promise.reject(error)
    }
  )

  return instance
}

export const request = createRequest()