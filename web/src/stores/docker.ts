import { ref } from 'vue'
import { defineStore } from 'pinia'

export interface DockerContainer {
  id: string
  name: string
  image: string
  state: string
  traefikLabels: Record<string, string>
  enabled: boolean
}

export const useDockerStore = defineStore('docker', () => {
  const containers = ref<DockerContainer[]>([])
  const available = ref(false)
  const loading = ref(false)
  const error = ref('')

  async function fetchContainers() {
    loading.value = true
    error.value = ''
    try {
      const res = await fetch('/api/docker')
      if (!res.ok) throw new Error(await res.text())
      const data = await res.json()
      containers.value = data.containers ?? []
      available.value = data.available ?? false
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  return { containers, available, loading, error, fetchContainers }
})
