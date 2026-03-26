<template>
  <div class="p-8 max-w-4xl">
    <!-- Header -->
    <div class="flex items-start justify-between mb-8 gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-100">Static Config</h1>
        <p class="text-slate-400 mt-1 text-sm">
          <span class="path-chip">{{ store.appConfig?.paths.staticConfig ?? '…' }}</span>
        </p>
      </div>
      <button class="btn btn-primary flex-shrink-0" :disabled="store.saving || store.loading" @click="save">
        <svg v-if="store.saving" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
        </svg>
        {{ store.saving ? 'Saving…' : 'Save' }}
      </button>
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

    <template v-else-if="store.appConfig">
      <!-- Tabs -->
      <div class="flex gap-1 bg-slate-800/60 border border-slate-700 rounded-lg p-1 mb-6 w-fit">
        <button v-for="t in ['form', 'yaml']" :key="t"
          class="px-4 py-1.5 text-sm rounded-md capitalize transition-colors"
          :class="tab === t ? 'bg-slate-700 text-slate-100 font-medium' : 'text-slate-400 hover:text-slate-200'"
          @click="tab = t as 'form' | 'yaml'"
        >{{ t }}</button>
      </div>

      <!-- Save feedback -->
      <div v-if="saveMsg" class="mb-5 text-sm px-4 py-2.5 rounded-lg border"
        :class="saveMsg.ok
          ? 'bg-emerald-900/30 border-emerald-800 text-emerald-400'
          : 'bg-red-900/30 border-red-800 text-red-400'">
        {{ saveMsg.text }}
      </div>
      <!-- Validation warnings -->
      <div v-if="warnings.length" class="mb-5 p-4 rounded-lg border bg-yellow-900/20 border-yellow-700/50 text-yellow-300 text-sm space-y-1">
        <div class="font-semibold mb-1">Saved with warnings:</div>
        <div v-for="w in warnings" :key="w.field" class="flex gap-2">
          <span class="font-mono text-yellow-400 flex-shrink-0">{{ w.field }}</span>
          <span>{{ w.message }}</span>
        </div>
      </div>

      <!-- ── FORM TAB ── -->
      <div v-if="tab === 'form'" class="space-y-5">

        <!-- API & Dashboard -->
        <Section title="API & Dashboard">
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Toggle v-model="form.apiDashboard" label="Enable dashboard" />
            <Toggle v-model="form.apiInsecure" label="Insecure mode (no auth)" />
          </div>
        </Section>

        <!-- Entry Points -->
        <Section title="Entry Points">
          <div class="space-y-3">
            <div v-for="(ep, idx) in form.entryPoints" :key="ep._id"
              class="bg-slate-900/60 border border-slate-700 rounded-xl p-4 space-y-3">
              <!-- Name + address row -->
              <div class="flex items-start gap-3">
                <div class="flex-1">
                  <label class="field-label">Name</label>
                  <input v-model="ep.name" type="text" class="input w-full" placeholder="websecure" />
                </div>
                <div class="flex-1">
                  <label class="field-label">Address</label>
                  <input v-model="ep.address" type="text" class="input w-full" placeholder=":443" />
                </div>
                <button class="mt-5 text-slate-500 hover:text-red-400 transition-colors"
                  @click="form.entryPoints.splice(idx, 1)">
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>

              <!-- HTTP redirect -->
              <div>
                <Toggle v-model="ep.redirect" label="HTTP redirect" small />
                <div v-if="ep.redirect" class="mt-3 grid grid-cols-2 gap-3 pl-1">
                  <div>
                    <label class="field-label">Redirect to</label>
                    <input v-model="ep.redirectTo" type="text" class="input w-full" placeholder="websecure" />
                  </div>
                  <div>
                    <label class="field-label">Scheme</label>
                    <select v-model="ep.redirectScheme" class="input w-full">
                      <option value="https">https</option>
                      <option value="http">http</option>
                    </select>
                  </div>
                  <Toggle v-model="ep.redirectPermanent" label="Permanent (308)" small />
                </div>
              </div>
            </div>

            <button class="btn btn-secondary w-full justify-center" @click="addEntryPoint">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
              </svg>
              Add entry point
            </button>
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

        <!-- Certificate Resolvers -->
        <Section title="Certificate Resolvers">
          <div class="space-y-3">
            <div v-for="(cr, idx) in form.certResolvers" :key="cr._id"
              class="bg-slate-900/60 border border-slate-700 rounded-xl p-4 space-y-4">
              <!-- Resolver name -->
              <div class="flex items-center gap-3">
                <div class="flex-1">
                  <label class="field-label">Resolver name</label>
                  <input v-model="cr.name" type="text" class="input w-full" placeholder="myresolver" />
                </div>
                <button class="mt-5 text-slate-500 hover:text-red-400 transition-colors"
                  @click="form.certResolvers.splice(idx, 1)">
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>

              <!-- ACME -->
              <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <div>
                  <label class="field-label">ACME email</label>
                  <input v-model="cr.email" type="email" class="input w-full" placeholder="you@example.com" />
                </div>
                <div>
                  <label class="field-label">Storage path</label>
                  <input v-model="cr.storage" type="text" class="input w-full" placeholder="/acme.json" />
                </div>
              </div>

              <!-- Challenge type -->
              <div>
                <label class="field-label">Challenge type</label>
                <div class="flex gap-2 mt-1.5">
                  <label v-for="ct in ['http', 'tls', 'dns']" :key="ct"
                    class="flex items-center gap-2 cursor-pointer px-3 py-2 rounded-lg border text-sm transition-colors"
                    :class="cr.challengeType === ct
                      ? 'border-sky-600 bg-sky-900/20 text-sky-300'
                      : 'border-slate-700 text-slate-400 hover:border-slate-600'">
                    <input v-model="cr.challengeType" type="radio" :value="ct" class="sr-only" />
                    {{ ct.toUpperCase() }}
                  </label>
                </div>
              </div>

              <!-- HTTP challenge fields -->
              <div v-if="cr.challengeType === 'http'">
                <label class="field-label">Entry point</label>
                <input v-model="cr.httpEntryPoint" type="text" class="input w-full" placeholder="web" />
              </div>

              <!-- DNS challenge fields -->
              <div v-if="cr.challengeType === 'dns'" class="space-y-3">
                <div>
                  <label class="field-label">DNS provider</label>
                  <input v-model="cr.dnsProvider" type="text" class="input w-full"
                    placeholder="cloudflare, route53, digitalocean…" />
                </div>
                <div>
                  <label class="field-label">Resolvers <span class="text-slate-600">(comma-separated)</span></label>
                  <input v-model="cr.dnsResolversStr" type="text" class="input w-full"
                    placeholder="1.1.1.1:53, 8.8.8.8:53" />
                </div>
              </div>
            </div>

            <button class="btn btn-secondary w-full justify-center" @click="addCertResolver">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
              </svg>
              Add resolver
            </button>
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
                Traefik will log to stdout. The Activity view streams logs directly
                from the container via the Docker socket.
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

      <!-- ── YAML TAB ── -->
      <div v-if="tab === 'yaml'">
        <textarea
          v-model="rawContent"
          class="w-full h-[calc(100vh-22rem)] font-mono text-sm bg-slate-900 border border-slate-700 rounded-xl p-4 text-slate-200 resize-none focus:outline-none focus:ring-2 focus:ring-sky-500/50 focus:border-sky-700 leading-relaxed"
          spellcheck="false"
          autocomplete="off"
        />
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted } from 'vue'
import { useConfigStore } from '@/stores/config'
import type { StaticConfig } from '@/stores/config'
import Section from '@/components/Section.vue'
import Toggle from '@/components/Toggle.vue'

const store = useConfigStore()
const tab = ref<'form' | 'yaml'>('form')
const rawContent = ref('')
const saveMsg = ref<{ ok: boolean; text: string } | null>(null)
const warnings = ref<Array<{ field: string; message: string }>>([])

let _idCounter = 0
const uid = () => String(++_idCounter)

// ── Form state ──────────────────────────────────────────────────────────────

interface EPForm {
  _id: string
  name: string
  address: string
  redirect: boolean
  redirectTo: string
  redirectScheme: string
  redirectPermanent: boolean
}

interface CRForm {
  _id: string
  name: string
  email: string
  storage: string
  challengeType: 'http' | 'tls' | 'dns'
  httpEntryPoint: string
  dnsProvider: string
  dnsResolversStr: string
}

const form = reactive({
  apiDashboard: false,
  apiInsecure: false,
  entryPoints: [] as EPForm[],
  dockerEnabled: false,
  dockerEndpoint: 'unix:///var/run/docker.sock',
  dockerNetwork: '',
  dockerExposedByDefault: false,
  fileEnabled: false,
  fileDirectory: '',
  fileWatch: true,
  certResolvers: [] as CRForm[],
  logLevel: 'INFO',
  accessLogEnabled: false,
  checkNewVersion: false,
  sendAnonymousUsage: false,
})

function populateForm(cfg: StaticConfig) {
  form.apiDashboard = cfg.api?.dashboard ?? false
  form.apiInsecure = cfg.api?.insecure ?? false

  form.entryPoints = Object.entries(cfg.entryPoints ?? {}).map(([name, ep]) => ({
    _id: uid(),
    name,
    address: ep.address ?? '',
    redirect: !!ep.http?.redirections?.entryPoint,
    redirectTo: ep.http?.redirections?.entryPoint?.to ?? '',
    redirectScheme: ep.http?.redirections?.entryPoint?.scheme ?? 'https',
    redirectPermanent: ep.http?.redirections?.entryPoint?.permanent ?? true,
  }))

  const docker = cfg.providers?.docker
  form.dockerEnabled = !!docker
  form.dockerEndpoint = docker?.endpoint ?? 'unix:///var/run/docker.sock'
  form.dockerNetwork = docker?.network ?? ''
  form.dockerExposedByDefault = docker?.exposedByDefault ?? false

  const file = cfg.providers?.file
  form.fileEnabled = !!file
  form.fileDirectory = file?.directory ?? ''
  form.fileWatch = file?.watch ?? true

  form.certResolvers = Object.entries(cfg.certificatesResolvers ?? {}).map(([name, cr]) => {
    const acme = cr.acme
    let challengeType: 'http' | 'tls' | 'dns' = 'http'
    if (acme?.dnsChallenge) challengeType = 'dns'
    else if (acme?.tlsChallenge) challengeType = 'tls'
    return {
      _id: uid(),
      name,
      email: acme?.email ?? '',
      storage: acme?.storage ?? '',
      challengeType,
      httpEntryPoint: acme?.httpChallenge?.entryPoint ?? '',
      dnsProvider: acme?.dnsChallenge?.provider ?? '',
      dnsResolversStr: (acme?.dnsChallenge?.resolvers ?? []).join(', '),
    }
  })

  form.logLevel = cfg.log?.level ?? 'INFO'
  form.accessLogEnabled = cfg.accessLog !== undefined && cfg.accessLog !== null

  form.checkNewVersion = cfg.global?.checkNewVersion ?? false
  form.sendAnonymousUsage = cfg.global?.sendAnonymousUsage ?? false
}

function buildConfig(): StaticConfig {
  const cfg: StaticConfig = {}

  cfg.api = { dashboard: form.apiDashboard, insecure: form.apiInsecure }

  if (form.entryPoints.length) {
    cfg.entryPoints = {}
    for (const ep of form.entryPoints) {
      if (!ep.name) continue
      const entry: StaticConfig['entryPoints'][string] = { address: ep.address }
      if (ep.redirect) {
        entry.http = {
          redirections: {
            entryPoint: {
              to: ep.redirectTo,
              scheme: ep.redirectScheme,
              permanent: ep.redirectPermanent,
            },
          },
        }
      }
      cfg.entryPoints[ep.name] = entry
    }
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

  if (form.certResolvers.length) {
    cfg.certificatesResolvers = {}
    for (const cr of form.certResolvers) {
      if (!cr.name) continue
      const acme: StaticConfig['certificatesResolvers'][string]['acme'] = {
        email: cr.email,
        storage: cr.storage,
      }
      if (cr.challengeType === 'http') {
        acme.httpChallenge = { entryPoint: cr.httpEntryPoint }
      } else if (cr.challengeType === 'tls') {
        acme.tlsChallenge = {}
      } else {
        acme.dnsChallenge = {
          provider: cr.dnsProvider,
          resolvers: cr.dnsResolversStr.split(',').map(s => s.trim()).filter(Boolean),
        }
      }
      cfg.certificatesResolvers[cr.name] = { acme }
    }
  }

  cfg.log = { level: form.logLevel }
  if (form.accessLogEnabled) {
    cfg.accessLog = {}
  }

  if (form.checkNewVersion || form.sendAnonymousUsage) {
    cfg.global = {
      checkNewVersion: form.checkNewVersion,
      sendAnonymousUsage: form.sendAnonymousUsage,
    }
  }

  return cfg
}

function addEntryPoint() {
  form.entryPoints.push({
    _id: uid(), name: '', address: ':80',
    redirect: false, redirectTo: 'websecure', redirectScheme: 'https', redirectPermanent: true,
  })
}

function addCertResolver() {
  form.certResolvers.push({
    _id: uid(), name: '', email: '', storage: '/acme.json',
    challengeType: 'dns', httpEntryPoint: 'web', dnsProvider: '', dnsResolversStr: '',
  })
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
