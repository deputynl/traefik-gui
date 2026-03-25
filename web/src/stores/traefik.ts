import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface TraefikCounts {
  total: number
  warnings: number
  errors: number
}

export interface TraefikOverview {
  http?: {
    routers?: TraefikCounts
    services?: TraefikCounts
    middlewares?: TraefikCounts
  }
}

export interface TraefikRouter {
  name: string
  rule: string
  status: string       // "enabled" | "disabled" | "warning" | "error"
  service: string
  entryPoints: string[]
  tls?: { certResolver?: string }
  provider: string
  priority?: number
}

export interface TraefikService {
  name: string
  type: string
  status: string
  provider: string
  serverStatus?: Record<string, string>
}

export const useTraefikStore = defineStore('traefik', () => {
  const overview = ref<TraefikOverview | null>(null)
  const routers = ref<TraefikRouter[]>([])
  const services = ref<TraefikService[]>([])
  const loading = ref(false)
  const online = ref(false)
  const lastFetched = ref<Date | null>(null)

  const routerCounts = computed(() => ({
    total: overview.value?.http?.routers?.total ?? 0,
    warnings: overview.value?.http?.routers?.warnings ?? 0,
    errors: overview.value?.http?.routers?.errors ?? 0,
  }))
  const serviceCounts = computed(() => ({
    total: overview.value?.http?.services?.total ?? 0,
    errors: overview.value?.http?.services?.errors ?? 0,
  }))
  const middlewareCounts = computed(() => ({
    total: overview.value?.http?.middlewares?.total ?? 0,
  }))

  async function fetchAll() {
    loading.value = true
    try {
      const [ovRes, rtRes] = await Promise.all([
        fetch('/api/traefik/api/overview'),
        fetch('/api/traefik/api/http/routers?per_page=100'),
      ])
      if (!ovRes.ok) throw new Error('overview failed')

      overview.value = await ovRes.json()
      routers.value = rtRes.ok ? await rtRes.json() : []
      online.value = true
      lastFetched.value = new Date()
    } catch {
      online.value = false
      overview.value = null
      routers.value = []
    } finally {
      loading.value = false
    }
  }

  return {
    overview, routers, services, loading, online, lastFetched,
    routerCounts, serviceCounts, middlewareCounts,
    fetchAll,
  }
})
