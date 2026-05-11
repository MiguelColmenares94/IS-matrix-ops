<template>
  <div class="min-h-screen bg-brand-gray flex items-center justify-center">
    <div class="bg-white rounded-xl shadow-md p-8 w-full max-w-sm">
      <h1 class="text-2xl font-bold text-brand-blue mb-6 text-center">Sign in</h1>
      <form @submit.prevent="submit" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-brand-navy mb-1">Email</label>
          <input
            v-model="email"
            type="email"
            required
            class="w-full border border-brand-lavender rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-brand-lavender"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-brand-navy mb-1">Password</label>
          <input
            v-model="password"
            type="password"
            required
            class="w-full border border-brand-lavender rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-brand-lavender"
          />
        </div>
        <p v-if="error" class="text-red-600 text-sm">{{ error }}</p>
        <button
          type="submit"
          :disabled="loading"
          class="w-full bg-brand-pink text-white font-semibold py-2 rounded-lg hover:opacity-90 disabled:opacity-50"
        >
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from './auth.store'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    router.push('/')
  } catch {
    error.value = 'Invalid email or password.'
  } finally {
    loading.value = false
  }
}
</script>
