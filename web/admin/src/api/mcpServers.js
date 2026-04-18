import { request } from './index'

export const mcpServerAPI = {
  // 获取所有 MCPServer
  list(params = {}) {
    return request.get('/api/admin/mcp-servers', { params })
  },

  // 获取单个 MCPServer
  get(id) {
    return request.get(`/api/admin/mcp-servers/${id}`)
  },

  // 创建 MCPServer
  create(data) {
    return request.post('/api/admin/mcp-servers', data)
  },

  // 更新 MCPServer
  update(id, data) {
    return request.put(`/api/admin/mcp-servers/${id}`, data)
  },

  // 删除 MCPServer
  delete(id) {
    return request.delete(`/api/admin/mcp-servers/${id}`)
  },

  // 恢复 MCPServer
  restore(id) {
    return request.post(`/api/admin/mcp-servers/${id}/restore`)
  },

  // 获取 Server 关联的工具
  getTools(id) {
    return request.get(`/api/admin/mcp-servers/${id}/tools`)
  },

  // 添加工具到 Server
  addTools(id, toolNames) {
    return request.post(`/api/admin/mcp-servers/${id}/tools`, { tool_names: toolNames })
  },

  // 从 Server 移除工具
  removeTool(id, toolName) {
    return request.delete(`/api/admin/mcp-servers/${id}/tools/${toolName}`)
  }
}
