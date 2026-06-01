import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { report } from '@/utils/clog'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

request.interceptors.request.use((config) => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers['Authorization'] = `Bearer ${auth.token}`
  }
  if (auth.csrfToken) {
    config.headers['X-CSRF-Token'] = auth.csrfToken
  }
  config.headers['X-Request-Id'] = crypto.randomUUID()
  return config
})

request.interceptors.response.use(
  (response) => {
    // Blob 响应（文件下载）直接返回，不走 envelope 解包
    if (response.config.responseType === 'blob') {
      return response.data
    }
    const { code, message, data } = response.data
    if (code === 0) return data
    if (code === 10002) {
      useAuthStore().logout()
      window.location.href = '/login'
      return Promise.reject(new Error(message))
    }
    if (code === 10003) {
      window.location.href = '/403'
      return Promise.reject(new Error(message))
    }
    ElMessage.error(message || '请求失败')
    return Promise.reject(new Error(message))
  },
  (error) => {
    const status: number | undefined = error.response?.status
    const reqUrl: string | undefined = error.config?.url
    // 仅 5xx 与网络错误（无 response）上报，4xx 业务错不上报
    if (!error.response || (typeof status === 'number' && status >= 500)) {
      try {
        report('warn', 'API_FAIL', {
          extra: {
            url: reqUrl,
            status: status ?? 0,
            code: error.response?.data?.code,
            message: error.response?.data?.message ?? error.message,
          },
        })
      } catch {
        // 静默
      }
    }
    ElMessage.error(error.response?.data?.message || '网络错误')
    return Promise.reject(error)
  }
)

export default request
