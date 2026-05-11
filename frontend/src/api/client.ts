import axios, { type InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '../domains/auth/auth.store'
import router from '../router'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL as string,
})

client.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
  const auth = useAuthStore()
  if (auth.accessToken) {
    const expiresAt = new Date(auth.expiresAt as string).getTime()
    const now = Date.now()
    if (expiresAt - now < 60_000 && expiresAt - now > 0) {
      const stay = window.confirm('Your session is about to expire. Stay logged in?')
      if (stay) {
        await auth.refresh()
      } else {
        await auth.logout()
        router.push('/login')
        return Promise.reject(new Error('Session expired'))
      }
    }
    config.headers.Authorization = `Bearer ${auth.accessToken}`
  }
  return config
})

client.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      const auth = useAuthStore()
      auth.clearSession()
      router.push('/login')
    }
    return Promise.reject(err)
  }
)

export default client
