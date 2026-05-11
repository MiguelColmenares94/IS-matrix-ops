import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '../../api/client'

const GENERIC_ERROR = 'An unexpected error occurred. Please try again.'

export const useStatsStore = defineStore('stats', () => {
  const statsResult = ref(null)
  const loading = ref(false)
  const error = ref(null)

  async function computeStats(q, r) {
    loading.value = true
    error.value = null
    statsResult.value = null
    try {
      const { data } = await client.post('/api/v2/stats', { q, r })
      statsResult.value = data
    } catch (err) {
      const status = err.response?.status
      error.value = (status >= 400 && status < 500)
        ? (err.response.data?.error || err.response.data?.message || GENERIC_ERROR)
        : GENERIC_ERROR
    } finally {
      loading.value = false
    }
  }

  function $reset() {
    statsResult.value = null
    loading.value = false
    error.value = null
  }

  return { statsResult, loading, error, computeStats, $reset }
})
