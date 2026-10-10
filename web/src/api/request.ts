import axios from 'axios'
import { analyticsRequest, analyticsResponse } from '@/utils/analytics'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

request.interceptors.request.use((config) => {
  // 标记请求所属页面版本；迟到的旧页面响应不能触发当前页面的 PV。
  analyticsRequest(config)
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

request.interceptors.response.use(
  (resp) => {
    const body = resp.data
    if (body && typeof body === 'object' && 'code' in body && body.code !== 0) {
      ElMessage.error(body.msg || '请求失败')
      return Promise.reject(new Error(body.msg || '请求失败'))
    }
    // 业务成功且有数据才通知埋点；具体是否主接口，由采集器继续判断。
    if (body.data != null) analyticsResponse(resp.config)
    return body.data
  },
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      if (!location.pathname.startsWith('/admin/login')) {
        ElMessage.error('登录已过期，请重新登录')
        location.href = '/admin/login'
      }
    } else {
      ElMessage.error(err.response?.data?.msg || err.message || '网络错误')
    }
    return Promise.reject(err)
  },
)

// 响应拦截器已剥壳，直接返回 data 字段
export function get<T>(url: string, params?: object): Promise<T> {
  return request.get(url, { params }) as unknown as Promise<T>
}

export function post<T>(url: string, data?: unknown): Promise<T> {
  return request.post(url, data) as unknown as Promise<T>
}

export function put<T>(url: string, data?: unknown): Promise<T> {
  return request.put(url, data) as unknown as Promise<T>
}

export function del<T>(url: string): Promise<T> {
  return request.delete(url) as unknown as Promise<T>
}

export default request
