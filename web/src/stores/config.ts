import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface ResolvedPaths {
  staticConfig: string
  dynamicDir: string
  acmePath: string
  staticConfigFound: boolean
}

export interface EntryPointRedirect {
  to?: string
  scheme?: string
  permanent?: boolean
}

export interface EntryPointHTTP {
  redirections?: {
    entryPoint?: EntryPointRedirect
  }
  tls?: {
    certResolver?: string
  }
}

export interface EntryPoint {
  address?: string
  http?: EntryPointHTTP
}

export interface DockerProvider {
  endpoint?: string
  exposedByDefault?: boolean
  network?: string
  watch?: boolean
}

export interface FileProvider {
  directory?: string
  filename?: string
  watch?: boolean
}

export interface ACMEConfig {
  email?: string
  storage?: string
  caServer?: string
  httpChallenge?: { entryPoint?: string }
  tlsChallenge?: Record<string, never>
  dnsChallenge?: {
    provider?: string
    resolvers?: string[]
    delayBeforeCheck?: number
    disablePropagationCheck?: boolean
  }
}

export interface CertResolver {
  acme?: ACMEConfig
}

export interface StaticConfig {
  api?: { dashboard?: boolean; insecure?: boolean; debug?: boolean }
  entryPoints?: Record<string, EntryPoint>
  providers?: {
    docker?: DockerProvider
    file?: FileProvider
  }
  certificatesResolvers?: Record<string, CertResolver>
  log?: { level?: string; filePath?: string; format?: string }
  accessLog?: Record<string, unknown> | null
  global?: { checkNewVersion?: boolean; sendAnonymousUsage?: boolean }
}

export interface AppConfig {
  paths: ResolvedPaths
  staticConfig?: StaticConfig
  rawConfig?: string
  traefikApiUrl: string
}

export interface Status {
  gui: string
  traefik: boolean
}

export const useConfigStore = defineStore('config', () => {
  const appConfig = ref<AppConfig | null>(null)
  const status = ref<Status | null>(null)
  const loading = ref(false)
  const saving = ref(false)
  const error = ref<string | null>(null)

  async function fetchConfig() {
    loading.value = true
    error.value = null
    try {
      const res = await fetch('/api/config')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      appConfig.value = await res.json()
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  async function fetchStatus() {
    try {
      const res = await fetch('/api/status')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      status.value = await res.json()
    } catch {
      status.value = { gui: 'ok', traefik: false }
    }
  }

  /** Save static config as JSON (from form editor). */
  async function saveConfigJSON(cfg: StaticConfig): Promise<boolean> {
    saving.value = true
    error.value = null
    try {
      const res = await fetch('/api/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(cfg),
      })
      if (!res.ok) {
        const j = await res.json().catch(() => ({ error: res.statusText }))
        throw new Error(j.error ?? res.statusText)
      }
      await fetchConfig()
      return true
    } catch (e) {
      error.value = String(e)
      return false
    } finally {
      saving.value = false
    }
  }

  /** Save static config as raw YAML (from YAML editor). */
  async function saveConfigRaw(yaml: string): Promise<boolean> {
    saving.value = true
    error.value = null
    try {
      const res = await fetch('/api/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'text/plain' },
        body: yaml,
      })
      if (!res.ok) {
        const j = await res.json().catch(() => ({ error: res.statusText }))
        throw new Error(j.error ?? res.statusText)
      }
      await fetchConfig()
      return true
    } catch (e) {
      error.value = String(e)
      return false
    } finally {
      saving.value = false
    }
  }

  return {
    appConfig, status, loading, saving, error,
    fetchConfig, fetchStatus, saveConfigJSON, saveConfigRaw,
  }
})
