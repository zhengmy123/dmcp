import { request } from './request'

export const toolsApi = {
  // 获取所有工具定义（现有端点，用于展示）
  getTools() {
    return request.get('/mcp')
  },

  // 获取特定服务的工具
  getServiceTools(vauthKey) {
    return request.get(`/mcp/${vauthKey}`)
  },

  // ========== 工具管理端点 ==========

  // 获取工具列表（管理）
  list(params = {}) {
    return request.get('/api/admin/tools', { params })
  },

  // 获取单个工具
  get(id) {
    return request.get(`/api/admin/tools/${id}`)
  },

  // 创建工具
  create(data) {
    return request.post('/api/admin/tools', data)
  },

  // 更新工具
  update(id, data) {
    return request.put(`/api/admin/tools/${id}`, data)
  },

  // 删除工具
  delete(id) {
    return request.delete(`/api/admin/tools/${id}`)
  }
}
