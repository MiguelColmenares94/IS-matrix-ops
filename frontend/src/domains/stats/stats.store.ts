import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '../../api/client'

const GENERIC_ERROR = 'An unexpected error occurred. Please try again.'

interface StatsResult {
  max: number
  min: number
  avg: number
  sum: number
  q_diagonal: boolean
  r_diagonal: boolean
}

export const useStatsStore = defineStore('stats', () => {
  const statsResult = ref<StatsResult | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function computeStats(q: number[][], r: number[][]): Promise<void> {
    loading.value = true
    error.value = null
    statsResult.value = null
    try {
      const { data } = await client.post<StatsResult>('/api/v2/stats', { q, r })
      statsResult.value = data
    } catch (err: unknown) {
      const e = err as { response?: { status: number; data?: { error?: string; message?: string } } }
      const status = e.response?.status ?? 0
      error.value = (status >= 400 && status < 500)
        ? (e.response?.data?.error || e.response?.data?.message || GENERIC_ERROR)
        : GENERIC_ERROR
    } finally {
      loading.value = false
    }
  }

  function $reset(): void {
    statsResult.value = null
    loading.value = false
    error.value = null
  }

  return { statsResult, loading, error, computeStats, $reset }
})
