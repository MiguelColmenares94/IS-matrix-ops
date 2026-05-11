<template>
  <div class="min-h-screen bg-brand-gray">
    <!-- Header -->
    <header class="bg-white shadow px-6 py-4 flex items-center justify-between">
      <h1 class="text-xl font-bold text-brand-blue">Inter Seguros coding-challenge</h1>
      <button @click="handleLogout" class="text-sm text-brand-navy hover:text-brand-pink">Logout</button>
    </header>

    <main class="max-w-4xl mx-auto p-6 space-y-6">
      <h2 class="text-lg font-semibold text-brand-navy text-center">Matrix Operations</h2>

      <!-- Input card -->
      <div class="bg-white rounded-xl shadow p-6 space-y-4 max-w-sm mx-auto">
        <!-- Tabs: each occupies 50% -->
        <div class="flex">
          <button
            @click="mode = 'grid'"
            :class="['flex-1 py-1.5 rounded-l-lg text-sm font-medium', mode === 'grid' ? 'bg-brand-blue text-white' : 'bg-brand-lavender text-brand-navy']"
          >Grid</button>
          <button
            @click="mode = 'json'"
            :class="['flex-1 py-1.5 rounded-r-lg text-sm font-medium', mode === 'json' ? 'bg-brand-blue text-white' : 'bg-brand-lavender text-brand-navy']"
          >JSON</button>
        </div>

        <!-- Grid mode -->
        <div v-if="mode === 'grid'" class="space-y-3">
          <div class="flex gap-4 items-center justify-center">
            <label class="text-sm text-brand-navy">Rows
              <input v-model.number="rows" type="number" min="1" max="10" class="ml-2 w-16 border border-brand-lavender rounded px-2 py-1 text-sm" />
            </label>
            <label class="text-sm text-brand-navy">Columns
              <input v-model.number="cols" type="number" min="1" max="10" class="ml-2 w-16 border border-brand-lavender rounded px-2 py-1 text-sm" />
            </label>
          </div>
          <div class="overflow-x-auto flex justify-center">
            <table>
              <tbody>
                <tr v-for="i in rows" :key="i">
                  <td v-for="j in cols" :key="j" class="p-1">
                    <input
                      v-model.number="grid[i-1][j-1]"
                      type="number"
                      class="w-16 border border-brand-lavender rounded px-2 py-1 text-sm text-center"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- JSON mode -->
        <div v-else>
          <textarea
            v-model="jsonInput"
            rows="5"
            placeholder="[[1, 2, 3], [4, 5, 6], [7, 8, 9]]"
            class="w-full border border-brand-lavender rounded-lg px-3 py-2 font-mono text-sm focus:outline-none focus:ring-2 focus:ring-brand-lavender"
          />
          <p v-if="jsonError" class="text-red-600 text-sm mt-1">{{ jsonError }}</p>
        </div>

        <!-- Action buttons: each occupies 50% -->
        <div class="flex">
          <button
            @click="computeQR"
            :disabled="matrixStore.loading"
            class="flex-1 bg-brand-pink text-white py-2 rounded-l-lg font-semibold hover:opacity-90 disabled:opacity-50"
          >
            {{ matrixStore.loading ? 'Computing…' : 'Calculate' }}
          </button>
          <button
            @click="handleClear"
            class="flex-1 bg-brand-lavender text-brand-navy py-2 rounded-r-lg font-semibold hover:opacity-90"
          >
            Clear
          </button>
        </div>
      </div>

      <!-- QR Skeleton -->
      <div v-if="matrixStore.loading" class="bg-white rounded-xl shadow p-6">
        <h2 class="font-semibold text-brand-navy mb-4">Calculating QR…</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div v-for="n in 2" :key="n" class="space-y-2">
            <div class="h-4 bg-brand-lavender rounded animate-pulse w-1/3"></div>
            <div v-for="r in 3" :key="r" class="flex gap-2">
              <div v-for="c in 3" :key="c" class="h-6 bg-brand-lavender rounded animate-pulse flex-1"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- QR Result card: shown on success OR error -->
      <div v-else-if="matrixStore.qResult || matrixStore.error" class="bg-white rounded-xl shadow p-6">
        <h2 class="font-semibold text-brand-navy mb-4 text-center">QR Result</h2>
        <p v-if="matrixStore.error" class="text-red-600 text-sm text-center">{{ matrixStore.error }}</p>
        <div v-else class="flex items-center justify-center">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 w-full">
            <div class="flex flex-col items-center justify-center">
              <h3 class="font-semibold text-brand-navy mb-2">Q Matrix</h3>
              <MatrixTable :data="matrixStore.qResult.q" />
            </div>
            <div class="flex flex-col items-center justify-center">
              <h3 class="font-semibold text-brand-navy mb-2">R Matrix</h3>
              <MatrixTable :data="matrixStore.qResult.r" />
            </div>
          </div>
        </div>
      </div>

      <!-- Stats Skeleton -->
      <div v-if="statsStore.loading" class="bg-white rounded-xl shadow p-6">
        <h2 class="font-semibold text-brand-navy mb-4">Calculating Statistics…</h2>
        <div class="grid grid-cols-2 gap-3">
          <div v-for="n in 6" :key="n" class="bg-brand-gray rounded-lg px-4 py-3 space-y-2">
            <div class="h-3 bg-brand-lavender rounded animate-pulse w-2/3"></div>
            <div class="h-5 bg-brand-lavender rounded animate-pulse w-1/2"></div>
          </div>
        </div>
      </div>

      <!-- Stats Result card: shown on success OR error -->
      <div v-else-if="statsStore.statsResult || statsStore.error" class="bg-white rounded-xl shadow p-6">
        <h2 class="font-semibold text-brand-navy mb-4 text-center">Statistics</h2>
        <p v-if="statsStore.error" class="text-red-600 text-sm text-center">{{ statsStore.error }}</p>
        <dl v-else class="grid grid-cols-2 gap-3">
          <div v-for="[label, value] in statsFields" :key="label" class="bg-brand-gray rounded-lg px-4 py-3 flex flex-col items-center justify-center">
            <dt class="text-xs text-brand-navy font-medium">{{ label }}</dt>
            <dd class="text-brand-blue font-semibold mt-1">{{ value }}</dd>
          </div>
        </dl>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../auth/auth.store'
import { useMatrixStore } from './matrix.store'
import { useStatsStore } from '../stats/stats.store'
import MatrixTable from './MatrixTable.vue'

const router = useRouter()
const auth = useAuthStore()
const matrixStore = useMatrixStore()
const statsStore = useStatsStore()

const mode = ref('grid')
const jsonInput = ref('')
const jsonError = ref('')
const rows = ref(3)
const cols = ref(3)

function makeGrid(r, c, old) {
  return Array.from({ length: r }, (_, i) =>
    Array.from({ length: c }, (_, j) => old?.[i]?.[j] ?? 0)
  )
}
const grid = ref(makeGrid(3, 3))
watch([rows, cols], ([r, c]) => { grid.value = makeGrid(r, c, grid.value) })

function parseInput() {
  if (mode.value === 'json') {
    try {
      const parsed = JSON.parse(jsonInput.value)
      if (!Array.isArray(parsed) || !parsed.every(r => Array.isArray(r))) throw new Error()
      jsonError.value = ''
      return parsed
    } catch {
      jsonError.value = 'Invalid JSON matrix format.'
      return null
    }
  }
  return grid.value.map(row => [...row])
}

async function computeQR() {
  const matrix = parseInput()
  if (!matrix) return
  statsStore.$reset()
  await matrixStore.computeQR(matrix)
  if (matrixStore.qResult) {
    const { q, r } = matrixStore.qResult
    await statsStore.computeStats(q, r)
  }
}

function handleClear() {
  matrixStore.$reset()
  statsStore.$reset()
  jsonInput.value = ''
  jsonError.value = ''
  grid.value = makeGrid(rows.value, cols.value)
}

async function handleLogout() {
  await auth.logout()
  router.push('/login')
}

const statsFields = computed(() => {
  const s = statsStore.statsResult
  if (!s) return []
  return [
    ['Maximum value', s.max.toFixed(6)],
    ['Minimum value', s.min.toFixed(6)],
    ['Average', s.avg.toFixed(6)],
    ['Total sum', s.sum.toFixed(6)],
    ['Q is diagonal', s.q_diagonal ? 'Yes' : 'No'],
    ['R is diagonal', s.r_diagonal ? 'Yes' : 'No'],
  ]
})
</script>
