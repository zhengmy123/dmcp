import { request } from './request'

export const toolsApi = {
  // 获取所有工具定义
  getTools() {
    return request.get('/mcp')
  },

  // 获取特定服务的工具
  getServiceTools(vauthKey) {
    return request.get(`/mcp/${vauthKey}`)
  }
}
