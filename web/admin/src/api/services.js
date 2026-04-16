import { request } from './request'

export const servicesApi = {
  // 获取所有 HTTP 服务
  getServices() {
    return request.get('/api/v1/services')
  },

  // 获取单个服务
  getService(id) {
    return request.get(`/api/v1/services/${id}`)
  },

  // 创建服务
  createService(data) {
    return request.post('/api/v1/services', data)
  },

  // 更新服务
  updateService(id, data) {
    return request.put(`/api/v1/services/${id}`, data)
  },

  // 删除服务
  deleteService(id) {
    return request.delete(`/api/v1/services/${id}`)
  },

  // 执行服务
  executeService(id, data) {
    return request.post(`/api/v1/execute/${id}`, data)
  },

  // 调试服务
  debugService(id, data) {
    return request.post(`/api/v1/services/${id}/debug`, data)
  }
}
