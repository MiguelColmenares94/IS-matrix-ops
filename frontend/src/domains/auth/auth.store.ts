import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

const BASE_URL = import.meta.env.VITE_API_BASE_URL as string

interface TokenPair {
  access_token: string
  refresh_token: string
  expires_at: string
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const expiresAt = ref<string | null>(null)

  const isAuthenticated = computed(
    () => !!accessToken.value && new Date(expiresAt.value as string) > new Date()
  )

  function setTokens(pair: TokenPair): void {
    accessToken.value = pair.access_token
    refreshToken.value = pair.refresh_token
    expiresAt.value = pair.expires_at
    localStorage.setItem('auth', JSON.stringify(pair))
  }

  function clearSession(): void {
    accessToken.value = null
    refreshToken.value = null
    expiresAt.value = null
    localStorage.removeItem('auth')
  }

  async function login(email: string, password: string): Promise<void> {
    const { data } = await axios.post<TokenPair>(`${BASE_URL}/api/v1/auth/login`, { email, password })
    setTokens(data)
  }

  async function logout(): Promise<void> {
    try {
      await axios.post(
        `${BASE_URL}/api/v1/auth/logout`,
        { refresh_token: refreshToken.value },
        { headers: { Authorization: `Bearer ${accessToken.value}` } }
      )
    } finally {
      clearSession()
    }
  }

  async function refresh(): Promise<void> {
    const { data } = await axios.post<TokenPair>(`${BASE_URL}/api/v1/auth/refresh`, {
      refresh_token: refreshToken.value,
    })
    setTokens(data)
  }

  function restoreSession(): void {
    const stored = localStorage.getItem('auth')
    if (stored) {
      const pair: TokenPair = JSON.parse(stored)
      accessToken.value = pair.access_token
      refreshToken.value = pair.refresh_token
      expiresAt.value = pair.expires_at
    }
  }

  return { accessToken, refreshToken, expiresAt, isAuthenticated, login, logout, refresh, restoreSession, clearSession }
})
