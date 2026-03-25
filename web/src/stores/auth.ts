import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<string | null>(null)
  const checked = ref(false)   // true once we've made at least one check
  const loading = ref(false)
  const error = ref<string | null>(null)

  /** Check whether the current session cookie is still valid. */
  async function check(): Promise<boolean> {
    loading.value = true
    try {
      const res = await fetch('/auth/check')
      if (res.ok) {
        const j = await res.json()
        user.value = j.user
        return true
      }
      user.value = null
      return false
    } catch {
      user.value = null
      return false
    } finally {
      loading.value = false
      checked.value = true
    }
  }

  async function login(username: string, password: string): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      const res = await fetch('/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })
      if (res.ok) {
        const j = await res.json()
        user.value = j.user
        return true
      }
      const j = await res.json().catch(() => ({ error: 'Login failed' }))
      error.value = j.error ?? 'Login failed'
      return false
    } catch (e) {
      error.value = String(e)
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    await fetch('/auth/logout', { method: 'POST' })
    user.value = null
    checked.value = false
  }

  return { user, checked, loading, error, check, login, logout }
})
