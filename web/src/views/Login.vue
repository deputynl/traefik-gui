<template>
  <div class="min-h-screen bg-slate-900 flex items-center justify-center p-4">
    <div class="w-full max-w-sm">
      <!-- Logo -->
      <div class="flex items-center justify-center gap-3 mb-8">
        <div class="w-10 h-10 rounded-xl bg-sky-500 flex items-center justify-center">
          <svg class="w-5 h-5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"/>
          </svg>
        </div>
        <div>
          <div class="text-lg font-bold text-slate-100 leading-tight">Traefik GUI</div>
          <div class="text-xs text-slate-500 leading-tight">Configuration Manager</div>
        </div>
      </div>

      <!-- Card -->
      <div class="card">
        <h1 class="text-base font-semibold text-slate-200 mb-5">Sign in</h1>

        <form class="space-y-4" @submit.prevent="submit">
          <div>
            <label class="field-label">Username</label>
            <input
              v-model="username"
              type="text"
              class="input w-full"
              autocomplete="username"
              autofocus
              required
            />
          </div>

          <div>
            <label class="field-label">Password</label>
            <input
              v-model="password"
              type="password"
              class="input w-full"
              autocomplete="current-password"
              required
            />
          </div>

          <div v-if="auth.error"
            class="text-sm text-red-400 bg-red-900/20 border border-red-800/50 rounded-lg px-3 py-2">
            {{ auth.error }}
          </div>

          <button
            type="submit"
            class="btn btn-primary w-full justify-center mt-1"
            :disabled="auth.loading"
          >
            <svg v-if="auth.loading" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
            </svg>
            {{ auth.loading ? 'Signing in…' : 'Sign in' }}
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const emit = defineEmits<{ success: [] }>()
const auth = useAuthStore()
const username = ref('')
const password = ref('')

async function submit() {
  const ok = await auth.login(username.value, password.value)
  if (ok) emit('success')
}
</script>
