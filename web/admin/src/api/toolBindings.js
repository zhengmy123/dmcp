import { request } from './request'

export const toolBindingsApi = {
  getToolBindings(toolId) {
    return request.get(`/api/admin/tool-bindings/${toolId}`)
  },

  bindTool(data) {
    return request.post('/api/admin/tool-bindings', data)
  },

  unbindTool(toolId, serverId) {
    return request.delete(`/api/admin/tool-bindings/${toolId}/${serverId}`)
  },

  batchBind(data) {
    return request.post('/api/admin/tool-bindings/batch-bind', data)
  },

  batchUnbind(bindingIds) {
    return request.delete('/api/admin/tool-bindings/batch-unbind', { binding_ids: bindingIds })
  },

  getServerBindings(serverId) {
    return request.get(`/api/admin/server-bindings/${serverId}`)
  }
}