import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface FileSummary {
  name: string
  active: boolean
  hostnames: string[]
  backends: string[]
  certResolver: string
  insecureSkipVerify: boolean
  routerCount: number
  serviceCount: number
  middlewareCount: number
}

export interface DynRouter {
  rule?: string
  entryPoints?: string[]
  service?: string
  tls?: { certResolver?: string; domains?: { main?: string; sans?: string[] }[] }
  priority?: number
  middlewares?: string[]
}

export interface DynService {
  loadBalancer?: {
    servers?: { url?: string }[]
    passHostHeader?: boolean
    serversTransport?: string
  }
}

export interface ParsedDynamic {
  http?: {
    routers?: Record<string, DynRouter>
    services?: Record<string, DynService>
    middlewares?: Record<string, unknown>
    serversTransports?: Record<string, { insecureSkipVerify?: boolean }>
  }
}

export interface FileDetail {
  name: string
  raw: string
  parsed: ParsedDynamic | null
}

export interface ServiceSpec {
  name: string
  hostname: string
  backendUrl: string
  insecureBackend: boolean
  certResolver: string
  entryPoints: string[]
}

export const useDynamicStore = defineStore('dynamic', () => {
  const files = ref<FileSummary[]>([])
  const currentFile = ref<FileDetail | null>(null)
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<string | null>(null)

  async function fetchFiles() {
    loading.value = true
    error.value = null
    try {
      const res = await fetch('/api/dynamic')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      files.value = await res.json()
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  async function fetchFile(name: string) {
    loading.value = true
    error.value = null
    currentFile.value = null
    try {
      const res = await fetch(`/api/dynamic/${encodeURIComponent(name)}`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      currentFile.value = await res.json()
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  async function saveFile(name: string, content: string): Promise<boolean> {
    saving.value = true
    error.value = null
    try {
      const res = await fetch(`/api/dynamic/${encodeURIComponent(name)}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'text/plain' },
        body: content,
      })
      if (!res.ok) {
        const j = await res.json().catch(() => ({ error: res.statusText }))
        throw new Error(j.error ?? res.statusText)
      }
      return true
    } catch (e) {
      error.value = String(e)
      return false
    } finally {
      saving.value = false
    }
  }

  async function deleteFile(name: string): Promise<boolean> {
    saving.value = true
    error.value = null
    try {
      const res = await fetch(`/api/dynamic/${encodeURIComponent(name)}`, {
        method: 'DELETE',
      })
      if (!res.ok) {
        const j = await res.json().catch(() => ({ error: res.statusText }))
        throw new Error(j.error ?? res.statusText)
      }
      files.value = files.value.filter(f => f.name !== name)
      return true
    } catch (e) {
      error.value = String(e)
      return false
    } finally {
      saving.value = false
    }
  }

  async function createService(spec: ServiceSpec): Promise<string | null> {
    saving.value = true
    error.value = null
    try {
      const res = await fetch('/api/dynamic', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(spec),
      })
      const j = await res.json()
      if (!res.ok) throw new Error(j.error ?? res.statusText)
      return j.name as string
    } catch (e) {
      error.value = String(e)
      return null
    } finally {
      saving.value = false
    }
  }

  return {
    files, currentFile, loading, saving, error,
    fetchFiles, fetchFile, saveFile, deleteFile, createService,
  }
})
