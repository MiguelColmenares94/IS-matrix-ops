import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '../../api/client'

const GENERIC_ERROR = 'An unexpected error occurred. Please try again.'

interface QRResult {
  q: number[][]
  r: number[][]
}

export const useMatrixStore = defineStore('matrix', () => {
  const qResult = ref<QRResult | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function computeQR(matrix: number[][]): Promise<void> {
    loading.value = true
    error.value = null
    qResult.value = null
    try {
      const { data } = await client.post<QRResult>('/api/v1/matrix/qr', { matrix })
      qResult.value = data
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
    qResult.value = null
    loading.value = false
    error.value = null
  }

  return { qResult, loading, error, computeQR, $reset }
})
