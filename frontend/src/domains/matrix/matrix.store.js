import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '../../api/client'

const GENERIC_ERROR = 'An unexpected error occurred. Please try again.'

export const useMatrixStore = defineStore('matrix', () => {
  const qResult = ref(null)
  const loading = ref(false)
  const error = ref(null)

  async function computeQR(matrix) {
    loading.value = true
    error.value = null
    qResult.value = null
    try {
      const { data } = await client.post('/api/v1/matrix/qr', { matrix })
      qResult.value = data
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
    qResult.value = null
    loading.value = false
    error.value = null
  }

  return { qResult, loading, error, computeQR, $reset }
})
