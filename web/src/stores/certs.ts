import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export interface CertInfo {
  resolver: string
  domain: string
  sans: string[]
  expiry: string
  daysLeft: number
}

export const useCertStore = defineStore('certs', () => {
  const certs = ref<CertInfo[]>([])
  const available = ref(false)
  const loading = ref(false)
  const error = ref('')

  const expiringSoon = computed(() => certs.value.filter(c => c.daysLeft >= 0 && c.daysLeft <= 30))
  const expired = computed(() => certs.value.filter(c => c.daysLeft < 0))

  async function fetchCerts() {
    loading.value = true
    error.value = ''
    try {
      const res = await fetch('/api/certificates')
      if (!res.ok) throw new Error(await res.text())
      const data = await res.json()
      certs.value = data.certs ?? []
      available.value = data.available ?? false
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  return { certs, available, loading, error, expiringSoon, expired, fetchCerts }
})
