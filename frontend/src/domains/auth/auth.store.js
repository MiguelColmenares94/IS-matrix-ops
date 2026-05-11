import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

const BASE_URL = import.meta.env.VITE_API_BASE_URL

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref(null)
  const refreshToken = ref(null)
  const expiresAt = ref(null)

  const isAuthenticated = computed(
    () => !!accessToken.value && new Date(expiresAt.value) > new Date()
  )

  function setTokens(pair) {
    accessToken.value = pair.access_token
    refreshToken.value = pair.refresh_token
    expiresAt.value = pair.expires_at
    localStorage.setItem('auth', JSON.stringify(pair))
  }

  function clearSession() {
    accessToken.value = null
    refreshToken.value = null
    expiresAt.value = null
    localStorage.removeItem('auth')
  }

  async function login(email, password) {
    const { data } = await axios.post(`${BASE_URL}/api/v1/auth/login`, { email, password })
    setTokens(data)
  }

  async function logout() {
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

  async function refresh() {
    const { data } = await axios.post(`${BASE_URL}/api/v1/auth/refresh`, {
      refresh_token: refreshToken.value,
    })
    setTokens(data)
  }

  function restoreSession() {
    const stored = localStorage.getItem('auth')
    if (stored) {
      const pair = JSON.parse(stored)
      accessToken.value = pair.access_token
      refreshToken.value = pair.refresh_token
      expiresAt.value = pair.expires_at
    }
  }

  return { accessToken, refreshToken, expiresAt, isAuthenticated, login, logout, refresh, restoreSession, clearSession }
})
