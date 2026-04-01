<template>
  <div class="flex flex-col flex-1 min-h-0">
    <!-- Header (fixed, doesn't scroll) -->
    <div class="flex-shrink-0 px-8 pt-8 pb-4">
    <div class="flex items-start justify-between mb-4 gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-100">Static Config</h1>
        <p class="text-slate-400 mt-1 text-sm">
          <span class="path-chip">{{ store.appConfig?.paths.staticConfig ?? '…' }}</span>
        </p>
      </div>
      <div class="flex gap-2 flex-shrink-0">
        <button class="btn btn-secondary flex items-center gap-1.5" :disabled="restarting" @click="restart"
          title="Restart Traefik container">
          <svg v-if="restarting" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
          </svg>
          <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round"
              d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0f01-15.357-2m15.357 2H15"/>
          </svg>
          {{ restarting ? 'Restarting…' : 'Restart Traefik' }}
        </button>
        <button class="btn btn-primary" :disabled="store.saving || store.loading" @click="save">
          <svg v-if="store.saving" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
          </svg>
          {{ store.saving ? 'Saving…' : 'Save' }}
        </button>
      </div>
    </div>

    <!-- Not found notice -->
    <div v-if="store.appConfig && !store.appConfig.paths.staticConfigFound"
      class="card border-yellow-800/50 bg-yellow-900/10 flex items-start gap-3 mb-6">
      <svg class="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
      </svg>
      <div>
        <p class="text-sm text-yellow-300 font-medium">Config file not found</p>
        <p class="text-xs text-slate-400 mt-0.5">
          No file at <span class="font-mono">{{ store.appConfig.paths.staticConfig }}</span>.
          Saving will create it.
        </p>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="store.loading" class="flex items-center gap-3 text-slate-400">
      <svg class="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
      </svg>
      Loading…
    </div>

    <template v-if="!store.loading && store.appConfig">
      <!-- Tabs -->
      <div class="flex gap-1 bg-slate-800/60 border border-slate-700 rounded-lg p-1 mb-4 w-fit">
        <button v-for="t in ['form', 'yaml']" :key="t"
          class="px-4 py-1.5 text-sm rounded-md capitalize transition-colors"
          :class="tab === t ? 'bg-slate-700 text-slate-100 font-medium' : 'text-slate-400 hover:text-slate-200'"
          @click="tab = t as 'form' | 'yaml'"
        >{{ t }}</button>
      </div>

      <!-- Save feedback -->
      <div v-if="saveMsg" class="mb-4 text-sm px-4 py-2.5 rounded-lg border"
        :class="saveMsg.ok
          ? 'bg-emerald-900/30 border-emerald-800 text-emerald-400'
          : 'bg-red-900/30 border-red-800 text-red-400'">
        {{ saveMsg.text }}
      </div>
      <!-- Validation warnings -->
      <div v-if="warnings.length" class="mb-4 p-4 rounded-lg border bg-yellow-900/20 border-yellow-700/50 text-yellow-300 text-sm space-y-1">
        <div class="font-semibold mb-1">Saved with warnings:</div>
        <div v-for="w in warnings" :key="w.field" class="flex gap-2">
          <span class="font-mono text-yellow-400 flex-shrink-0">{{ w.field }}</span>
          <span>{{ w.message }}</span>
        </div>
      </div>
    </template>
    </div><!-- end header -->

    <!-- ── FORM TAB: scrollable content area ── -->
    <div v-if="!store.loading && store.appConfig && tab === 'form'"
      class="flex-1 min-h-0 overflow-y-auto px-8 pb-8">
      <div class="space-y-5">

        <!-- API & Dashboard -->
        <Section title="API & Dashboard">
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Toggle v-model="form.apiDashboard" label="Enable dashboard" />
            <Toggle v-model="form.apiInsecure" label="Insecure mode (no auth)" />
          </div>
        </Section>

        <!-- Entry Points (fixed: web / websecure / websecuremtls) -->
        <Section title="Entry Points">
          <div class="space-y-3">

            <!-- web -->
            <div class="bg-slate-900/60 border border-slate-700 rounded-xl p-4 space-y-3">
              <div class="flex items-center justify-between gap-3">
                <span class="text-sm font-medium text-slate-200 font-mono">web</span>
                <div class="flex items-center gap-2">
                  <label class="field-label mb-0">Address</label>
                  <input v-model="form.webAddress" type="text" class="input text-xs w-24 text-right" />
                </div>
              </div>
              <Toggle v-model="form.httpRedirect" label="Redirect HTTP → HTTPS (websecure)" small />
            </div>

            <!-- websecure -->
            <div class="bg-slate-900/60 border border-slate-700 rounded-xl p-4 space-y-3">
              <div class="flex items-center justify-between gap-3">
                <span class="text-sm font-medium text-slate-200 font-mono">websecure</span>
                <div class="flex items-center gap-2">
                  <label class="field-label mb-0">Address</label>
                  <input v-model="form.websecureAddress" type="text" class="input text-xs w-24 text-right" />
                </div>
              </div>
              <p class="text-xs text-slate-500">
                TLS via cert resolver
                <span v-if="form.crName" class="font-mono text-slate-400">{{ form.crName }}</span>
                <span v-else class="text-slate-600">(configure a resolver below)</span>.
                Docker containers expose services here.
              </p>
            </div>

            <!-- websecuremtls -->
            <div class="bg-slate-900/60 border border-slate-700 rounded-xl p-4 space-y-3">
              <div class="flex items-center justify-between gap-3">
                <span class="text-sm font-medium text-slate-200 font-mono">websecuremtls</span>
                <div class="flex items-center gap-2">
                  <label class="field-label mb-0">Address</label>
                  <input v-model="form.websecuremtlsAddress" type="text" class="input text-xs w-24 text-right" />
                </div>
              </div>
              <p class="text-xs text-slate-500">
                External mTLS entrypoint. A TCP router (<span class="font-mono text-slate-400">HostSNI(`*`)</span>)
                enforces client certificate verification at the TLS handshake level, then forwards
                plaintext traffic to <span class="font-mono text-slate-400">websecure-internal</span>.
                Managed via the
                <router-link to="/mtls" class="text-sky-500 hover:text-sky-400 transition-colors">mTLS page</router-link>.
              </p>
            </div>

          </div>
        </Section>

        <!-- Providers -->
        <Section title="Providers">
          <!-- Docker -->
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-slate-300">Docker</span>
              <Toggle v-model="form.dockerEnabled" label="" small />
            </div>
            <div v-if="form.dockerEnabled" class="bg-slate-900/60 border border-slate-700 rounded-xl p-4 space-y-3">
              <div>
                <label class="field-label">Socket endpoint</label>
                <input v-model="form.dockerEndpoint" type="text" class="input w-full"
                  placeholder="unix:///var/run/docker.sock" />
              </div>
              <div>
                <label class="field-label">Default network <span class="text-slate-600">(optional)</span></label>
                <input v-model="form.dockerNetwork" type="text" class="input w-full" />
              </div>
              <Toggle v-model="form.dockerExposedByDefault" label="Expose all containers by default" small />
            </div>

            <div class="border-t border-slate-700/50 pt-3 mt-1">
              <div class="flex items-center justify-between">
                <span class="text-sm font-medium text-slate-300">File provider</span>
                <Toggle v-model="form.fileEnabled" label="" small />
              </div>
            </div>
            <div v-if="form.fileEnabled" class="bg-slate-900/60 border border-slate-700 rounded-xl p-4 space-y-3">
              <div>
                <label class="field-label">Dynamic config directory</label>
                <input v-model="form.fileDirectory" type="text" class="input w-full"
                  placeholder="/etc/traefik/dynamic" />
              </div>
              <Toggle v-model="form.fileWatch" label="Watch for changes" small />
            </div>
          </div>
        </Section>

        <!-- Certificate Resolver (single) -->
        <Section title="Certificate Resolver">
          <div class="space-y-3">
            <Toggle v-model="form.hasCertResolver" label="Enable ACME certificate resolver" small />

            <div v-if="form.hasCertResolver" class="bg-slate-900/60 border border-slate-700 rounded-xl p-4 space-y-4">
              <div>
                <label class="field-label">Resolver name</label>
                <input v-model="form.crName" type="text" class="input w-full" placeholder="myresolver" />
                <p class="text-xs text-slate-600 mt-1">
                  Referenced by <span class="font-mono">websecure</span> and the mTLS TCP router.
                </p>
              </div>

              <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <label class="field-label">ACME email</label>
                  <input v-model="form.crEmail" type="email" class="input w-full" placeholder="you@example.com" />
                </div>
                <div>
                  <label class="field-label">Storage path</label>
                  <input v-model="form.crStorage" type="text" class="input w-full" placeholder="/acme.json" />
                </div>
              </div>

              <div>
                <label class="field-label">Challenge type</label>
                <div class="flex gap-2 mt-1.5">
                  <label v-for="ct in ['http', 'tls', 'dns']" :key="ct"
                    class="flex items-center gap-2 cursor-pointer px-3 py-2 rounded-lg border text-sm transition-colors"
                    :class="form.crChallengeType === ct
                      ? 'border-sky-600 bg-sky-900/20 text-sky-300'
                      : 'border-slate-700 text-slate-400 hover:border-slate-600'">
                    <input v-model="form.crChallengeType" type="radio" :value="ct" class="sr-only" />
                    {{ ct.toUpperCase() }}
                  </label>
                </div>
              </div>

              <div v-if="form.crChallengeType === 'http'">
                <label class="field-label">Entry point</label>
                <input v-model="form.crHttpEntryPoint" type="text" class="input w-full" placeholder="web" />
              </div>

              <div v-if="form.crChallengeType === 'dns'" class="space-y-3">
                <div>
                  <label class="field-label">DNS provider</label>
                  <select v-model="form.crDnsProvider" class="input w-full">
                    <option value="">— select a provider —</option>
                    <option v-for="p in DNS_PROVIDERS" :key="p.id" :value="p.id">{{ p.name }}</option>
                  </select>
                </div>

                <div v-if="form.crDnsProvider" class="rounded-lg border border-sky-800/40 bg-sky-900/20 p-3 space-y-2 text-xs">
                  <p class="font-medium text-sky-300">Set on your Traefik container:</p>
                  <div v-if="findProvider(form.crDnsProvider)" class="space-y-1 font-mono">
                    <div v-for="env in findProvider(form.crDnsProvider)!.envVars" :key="env" class="text-slate-300">
                      {{ env }}<span class="text-slate-500">=your-value</span>
                    </div>
                  </div>
                  <a :href="`https://go-acme.github.io/lego/dns/${form.crDnsProvider}/`"
                    target="_blank" rel="noopener"
                    class="inline-flex items-center gap-1 text-sky-400 hover:text-sky-300 transition-colors">
                    Full documentation
                    <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
                    </svg>
                  </a>
                </div>

                <div>
                  <label class="field-label">Resolvers <span class="text-slate-600">(comma-separated, optional)</span></label>
                  <input v-model="form.crDnsResolversStr" type="text" class="input w-full"
                    placeholder="1.1.1.1:53, 8.8.8.8:53" />
                </div>
              </div>
            </div>
          </div>
        </Section>

        <!-- Logging -->
        <Section title="Logging">
          <div class="space-y-4">
            <div>
              <label class="field-label">Log level</label>
              <select v-model="form.logLevel" class="input">
                <option v-for="l in ['DEBUG', 'INFO', 'WARN', 'ERROR']" :key="l" :value="l">{{ l }}</option>
              </select>
            </div>

            <div class="border-t border-slate-700 pt-4">
              <Toggle v-model="form.accessLogEnabled" label="Enable access log" small />
              <p v-if="form.accessLogEnabled" class="text-xs text-slate-500 mt-2 pl-1">
                Traefik will log to stdout in JSON format. The Activity view streams
                logs directly from the container via the Docker socket.
              </p>
            </div>
          </div>
        </Section>

        <!-- Global -->
        <Section title="Global">
          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <Toggle v-model="form.checkNewVersion" label="Check for new Traefik versions" small />
            <Toggle v-model="form.sendAnonymousUsage" label="Send anonymous usage statistics" small />
          </div>
        </Section>

      </div>
    </div>

    <!-- ── YAML TAB: fills remaining height, no outer scroll ── -->
    <div v-if="!store.loading && store.appConfig && tab === 'yaml'"
      class="flex-1 min-h-0 px-8 pb-8">
      <textarea
        v-model="rawContent"
        class="w-full h-full font-mono text-sm bg-slate-900 border border-slate-700 rounded-xl p-4 text-slate-200 resize-none focus:outline-none focus:ring-2 focus:ring-sky-500/50 focus:border-sky-700 leading-relaxed"
        spellcheck="false"
        autocomplete="off"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { useConfigStore } from '@/stores/config'
import type { StaticConfig } from '@/stores/config'
import Section from '@/components/Section.vue'
import Toggle from '@/components/Toggle.vue'
import { DNS_PROVIDERS, findProvider } from '@/data/dnsProviders'

const store = useConfigStore()
const tab = ref<'form' | 'yaml'>('form')
const rawContent = ref('')
const saveMsg = ref<{ ok: boolean; text: string } | null>(null)
const restarting = ref(false)
const warnings = ref<Array<{ field: string; message: string }>>([])

// ── Form state ──────────────────────────────────────────────────────────────

const form = reactive({
  apiDashboard: false,
  apiInsecure: false,
  // Fixed entrypoints
  webAddress: ':80',
  httpRedirect: true,
  websecureAddress: ':443',
  websecuremtlsAddress: ':8443',
  // Single cert resolver
  hasCertResolver: false,
  crName: '',
  crEmail: '',
  crStorage: '/acme.json',
  crChallengeType: 'dns' as 'http' | 'tls' | 'dns',
  crHttpEntryPoint: 'web',
  crDnsProvider: '',
  crDnsResolversStr: '',
  // Providers
  dockerEnabled: false,
  dockerEndpoint: 'unix:///var/run/docker.sock',
  dockerNetwork: '',
  dockerExposedByDefault: false,
  fileEnabled: false,
  fileDirectory: '',
  fileWatch: true,
  // Logging
  logLevel: 'INFO',
  accessLogEnabled: false,
  // Global
  checkNewVersion: false,
  sendAnonymousUsage: false,
})

function populateForm(cfg: StaticConfig) {
  form.apiDashboard = cfg.api?.dashboard ?? false
  form.apiInsecure = cfg.api?.insecure ?? false

  const web = cfg.entryPoints?.['web']
  const websecure = cfg.entryPoints?.['websecure']
  const websecuremtls = cfg.entryPoints?.['websecuremtls']

  form.webAddress = web?.address ?? ':80'
  form.httpRedirect = !!web?.http?.redirections?.entryPoint
  form.websecureAddress = websecure?.address ?? ':443'
  form.websecuremtlsAddress = websecuremtls?.address ?? ':8443'

  // Single cert resolver — take the first one defined
  const resolvers = Object.entries(cfg.certificatesResolvers ?? {})
  form.hasCertResolver = resolvers.length > 0
  if (resolvers.length > 0) {
    const [name, cr] = resolvers[0]
    const acme = cr.acme
    form.crName = name
    form.crEmail = acme?.email ?? ''
    form.crStorage = acme?.storage ?? '/acme.json'
    if (acme?.dnsChallenge) form.crChallengeType = 'dns'
    else if (acme?.tlsChallenge) form.crChallengeType = 'tls'
    else form.crChallengeType = 'http'
    form.crHttpEntryPoint = acme?.httpChallenge?.entryPoint ?? 'web'
    form.crDnsProvider = acme?.dnsChallenge?.provider ?? ''
    form.crDnsResolversStr = (acme?.dnsChallenge?.resolvers ?? []).join(', ')
  } else {
    form.crName = ''
    form.crEmail = ''
    form.crStorage = '/acme.json'
    form.crChallengeType = 'dns'
    form.crHttpEntryPoint = 'web'
    form.crDnsProvider = ''
    form.crDnsResolversStr = ''
  }

  const docker = cfg.providers?.docker
  form.dockerEnabled = !!docker
  form.dockerEndpoint = docker?.endpoint ?? 'unix:///var/run/docker.sock'
  form.dockerNetwork = docker?.network ?? ''
  form.dockerExposedByDefault = docker?.exposedByDefault ?? false

  const file = cfg.providers?.file
  form.fileEnabled = !!file
  form.fileDirectory = file?.directory ?? ''
  form.fileWatch = file?.watch ?? true

  form.logLevel = cfg.log?.level ?? 'INFO'
  form.accessLogEnabled = cfg.accessLog !== undefined && cfg.accessLog !== null

  form.checkNewVersion = cfg.global?.checkNewVersion ?? false
  form.sendAnonymousUsage = cfg.global?.sendAnonymousUsage ?? false
}

function buildConfig(): StaticConfig {
  const cfg: StaticConfig = {}

  cfg.api = { dashboard: form.apiDashboard, insecure: form.apiInsecure }

  // Always emit the three standard entrypoints
  cfg.entryPoints = {
    web: {
      address: form.webAddress,
      ...(form.httpRedirect ? {
        http: {
          redirections: { entryPoint: { to: 'websecure', scheme: 'https' } }
        }
      } : {}),
    },
    websecure: {
      address: form.websecureAddress,
      ...(form.hasCertResolver && form.crName ? {
        http: { tls: { certResolver: form.crName } }
      } : {}),
    },
    websecuremtls: {
      address: form.websecuremtlsAddress,
      // No http.tls — mTLS is enforced by the TCP router in mtls.yml (HostSNI(`*`)).
    },
    'websecure-internal': {
      address: '127.0.0.1:8444',
      // Plain HTTP, loopback only. Receives traffic forwarded from websecuremtls
      // after the TCP router has terminated TLS and verified the client certificate.
    },
  }

  if (form.dockerEnabled || form.fileEnabled) {
    cfg.providers = {}
    if (form.dockerEnabled) {
      cfg.providers.docker = {
        endpoint: form.dockerEndpoint,
        exposedByDefault: form.dockerExposedByDefault,
        ...(form.dockerNetwork ? { network: form.dockerNetwork } : {}),
      }
    }
    if (form.fileEnabled) {
      cfg.providers.file = {
        directory: form.fileDirectory,
        watch: form.fileWatch,
      }
    }
  }

  if (form.hasCertResolver && form.crName) {
    const acme: StaticConfig['certificatesResolvers'][string]['acme'] = {
      email: form.crEmail,
      storage: form.crStorage,
    }
    if (form.crChallengeType === 'http') {
      acme.httpChallenge = { entryPoint: form.crHttpEntryPoint }
    } else if (form.crChallengeType === 'tls') {
      acme.tlsChallenge = {}
    } else {
      acme.dnsChallenge = {
        provider: form.crDnsProvider,
        resolvers: form.crDnsResolversStr.split(',').map(s => s.trim()).filter(Boolean),
      }
    }
    cfg.certificatesResolvers = { [form.crName]: { acme } }
  }

  cfg.log = { level: form.logLevel }
  if (form.accessLogEnabled) {
    cfg.accessLog = { format: 'json' }
  }

  if (form.checkNewVersion || form.sendAnonymousUsage) {
    cfg.global = {
      checkNewVersion: form.checkNewVersion,
      sendAnonymousUsage: form.sendAnonymousUsage,
    }
  }

  return cfg
}

async function restart() {
  restarting.value = true
  saveMsg.value = { ok: true, text: 'Restarting Traefik…' }
  try {
    const res = await fetch('/api/traefik/restart', { method: 'POST' })
    if (!res.ok) {
      if (res.status < 500 || res.status === 500) {
        const j = await res.json().catch(() => ({}))
        if (j.error) {
          saveMsg.value = { ok: false, text: j.error }
          restarting.value = false
          return
        }
      }
    }
  } catch {
    // Connection dropped because Traefik restarted mid-request — expected.
  }
  saveMsg.value = { ok: true, text: 'Waiting for Traefik to come back…' }
  await new Promise(r => setTimeout(r, 4_000))
  try {
    const ctrl = new AbortController()
    setTimeout(() => ctrl.abort(), 3000)
    const probe = await fetch('/api/status', { signal: ctrl.signal })
    if (probe.ok) {
      saveMsg.value = { ok: true, text: 'Traefik restarted successfully.' }
      setTimeout(() => { saveMsg.value = null }, 4000)
    } else {
      saveMsg.value = { ok: false, text: 'Traefik may still be starting up — please check.' }
    }
  } catch {
    saveMsg.value = { ok: false, text: 'Traefik did not respond after 20 s — please check.' }
  }
  restarting.value = false
}

async function save() {
  saveMsg.value = null
  warnings.value = []
  try {
    const warns = tab.value === 'yaml'
      ? await store.saveConfigRaw(rawContent.value)
      : await store.saveConfigJSON(buildConfig())
    warnings.value = warns
    saveMsg.value = { ok: true, text: 'Saved successfully.' }
    setTimeout(() => { saveMsg.value = null }, 3000)
  } catch {
    saveMsg.value = { ok: false, text: store.error ?? 'Save failed.' }
  }
}

// Re-populate form whenever the store data changes (after save or initial load).
watch(() => store.appConfig, (ac) => {
  if (!ac) return
  rawContent.value = ac.rawConfig ?? ''
  if (ac.staticConfig) populateForm(ac.staticConfig)
}, { immediate: true })

onMounted(async () => {
  if (!store.appConfig) await store.fetchConfig()
})
</script>
