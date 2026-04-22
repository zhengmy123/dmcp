import { request } from './request'

export const systemConfigApi = {
  getConfig(key) {
    return request.get(`/api/v1/system/config/${key}`)
  },

  updateConfig(key, configValue) {
    return request.put(`/api/v1/system/config/${key}`, { config_value: configValue })
  }
}
