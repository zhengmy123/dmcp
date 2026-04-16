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
      // 使用 JWT token 认证
      const jwtToken = getJWTToken()
      if (jwtToken) {
        config.headers['Authorization'] = `Bearer ${jwtToken}`
      }
      return config
    },
    (error) => Promise.reject(error)
  )

  instance.interceptors.response.use(
    (response) => response.data,
    (error) => {
      if (error.response?.status === 401) {
        const authStore = useAuthStore()
        authStore.logout()
        // 如果不是在登录页，跳转到登录页
        if (!window.location.pathname.includes('/login')) {
          window.location.href = '/login'
        }
      }
      return Promise.reject(error)
    }
  )

  return instance
}

export const request = createRequest()
