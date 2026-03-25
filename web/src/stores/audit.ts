import { ref } from 'vue'
import { defineStore } from 'pinia'

export interface AuditEntry {
  time: string
  user: string
  action: string
  detail: string
}

export const useAuditStore = defineStore('audit', () => {
  const entries = ref<AuditEntry[]>([])
  const loading = ref(false)
  const error = ref('')

  async function fetchAudit() {
    loading.value = true
    error.value = ''
    try {
      const res = await fetch('/api/audit')
      if (!res.ok) throw new Error(await res.text())
      const data = await res.json()
      entries.value = data.entries ?? []
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  return { entries, loading, error, fetchAudit }
})
