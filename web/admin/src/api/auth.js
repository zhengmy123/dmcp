import { request } from './request'

export const authApi = {
  // 用户登录
  login(username, password) {
    return request.post('/auth/login', { username, password })
  },

  // 获取当前用户
  getCurrentUser() {
    return request.get('/auth/me')
  },

  // 修改密码
  changePassword(data) {
    return request.post('/auth/change-password', data)
  },

  // 获取所有用户
  getUsers() {
    return request.get('/api/v1/users')
  },

  // 创建用户
  createUser(data) {
    return request.post('/api/v1/users', data)
  },

  // 更新用户
  updateUser(id, data) {
    return request.put(`/api/v1/users/${id}`, data)
  },

  // 删除用户
  deleteUser(id) {
    return request.delete(`/api/v1/users/${id}`)
  },

  // 获取所有 Token
  getTokens() {
    return request.get('/api/v1/auth/tokens')
  },

  // 创建 Token
  createToken(data) {
    return request.post('/api/v1/auth/tokens', data)
  },

  // 删除 Token
  deleteToken(token) {
    return request.delete(`/api/v1/auth/tokens/${token}`)
  },

  // 刷新 Token
  refreshToken(token) {
    return request.post(`/api/v1/auth/tokens/${token}/refresh`)
  },

  // 启用 Token
  enableToken(token) {
    return request.post(`/api/v1/auth/tokens/${token}/enable`)
  },

  // 禁用 Token
  disableToken(token) {
    return request.post(`/api/v1/auth/tokens/${token}/disable`)
  }
}
