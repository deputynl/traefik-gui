<template>
  <div class="p-8 flex-1 overflow-y-auto min-h-0">
    <!-- Back + title -->
    <div class="flex items-center gap-3 mb-6">
      <button class="btn btn-secondary px-2.5 py-2" @click="router.push({ name: 'dynamic' })">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7"/>
        </svg>
      </button>
      <div class="min-w-0">
        <h1 class="text-xl font-bold text-slate-100 truncate">{{ store.isNewFile && filename === '_new_' ? 'New service' : filename }}</h1>
        <p class="text-xs text-slate-500 mt-0.5">{{ store.isNewFile ? 'New dynamic config file' : store.currentFile ? 'Dynamic config file' : '…' }}</p>
      </div>
      <div class="ml-auto flex items-center gap-2 flex-shrink-0">
        <button
          class="btn btn-primary"
          :disabled="store.saving || store.loading"
          @click="save"
        >
          <svg v-if="store.saving" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
          </svg>
          {{ store.saving ? (store.isNewFile ? 'Creating…' : 'Saving…') : (store.isNewFile ? 'Create' : 'Save') }}
        </button>
        <button v-if="!store.isNewFile"
          class="btn btn-secondary text-red-400 hover:text-red-300 hover:bg-red-900/20"
          @click="confirmDelete = true">
          Delete
        </button>
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

    <template v-else-if="store.currentFile || store.isNewFile">
      <!-- mTLS managed file warning -->
      <div v-if="filename === 'mtls.yml'"
        class="mb-4 p-3 bg-yellow-900/30 border border-yellow-800 rounded-lg flex items-start gap-3">
        <svg class="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <div>
          <p class="text-sm text-yellow-300 font-medium">Managed by mTLS configuration</p>
          <p class="text-xs text-slate-400 mt-0.5">
            This file is automatically regenerated whenever mTLS settings change.
            Manual edits will be overwritten. Use the
            <router-link to="/mtls" class="text-sky-400 hover:text-sky-300 transition-colors">mTLS page</router-link>
            to manage routers, public exceptions, and TLS options.
          </p>
        </div>
      </div>

      <!-- Tabs -->
      <div class="flex gap-1 bg-slate-800/60 border border-slate-700 rounded-lg p-1 mb-5 w-fit">
        <button
          class="px-4 py-1.5 text-sm rounded-md transition-colors"
          :class="tab === 'form' ? 'bg-slate-700 text-slate-100 font-medium' : 'text-slate-400 hover:text-slate-200'"
          @click="tab = 'form'"
        >Form</button>
        <button
          class="px-4 py-1.5 text-sm rounded-md transition-colors"
          :class="tab === 'yaml' ? 'bg-slate-700 text-slate-100 font-medium' : 'text-slate-400 hover:text-slate-200'"
          @click="tab = 'yaml'"
        >YAML</button>
      </div>

      <!-- Save feedback -->
      <div v-if="saveMsg" class="mb-4 flex items-center gap-2 text-sm px-4 py-2.5 rounded-lg border"
        :class="saveMsg.ok
          ? 'bg-emerald-900/30 border-emerald-800 text-emerald-400'
          : 'bg-red-900/30 border-red-800 text-red-400'"
      >
        {{ saveMsg.text }}
      </div>

      <!-- Form tab -->
      <div v-if="tab === 'form'" class="space-y-5">
        <!-- Advanced notice -->
        <div v-if="!isSimplePattern" class="card border-yellow-800/50 bg-yellow-900/10 flex items-start gap-3">
          <svg class="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
          <div>
            <p class="text-sm text-yellow-300 font-medium">Advanced configuration</p>
            <p class="text-xs text-slate-400 mt-0.5">This file uses rules or settings beyond the simple pattern. Use the YAML tab to edit it.</p>
          </div>
        </div>

        <template v-if="isSimplePattern && formData">
          <div class="card space-y-4">
            <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wider">Router</h2>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1.5">Hostname</label>
              <input v-model="formData.hostname" type="text" class="input w-full" placeholder="myapp.example.com" />
            </div>
          </div>

          <div class="card space-y-4">
            <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wider">Service</h2>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1.5">Backend URL</label>
              <input v-model="formData.backendUrl" type="text" class="input w-full" placeholder="http://192.168.1.10:8080" />
            </div>

            <label class="flex items-center gap-3 cursor-pointer select-none">
              <div class="relative">
                <input v-model="formData.insecureBackend" type="checkbox" class="sr-only peer" />
                <div class="w-9 h-5 bg-slate-700 peer-checked:bg-sky-500 rounded-full transition-colors"></div>
                <div class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform peer-checked:translate-x-4"></div>
              </div>
              <div>
                <span class="text-sm text-slate-300">Skip TLS verification</span>
                <p class="text-xs text-slate-500">For backends with self-signed certificates</p>
              </div>
            </label>
          </div>

          <p class="text-xs text-slate-500">
            Saving from this form will regenerate the YAML. Use the YAML tab to preserve advanced settings.
          </p>
        </template>
      </div>

      <!-- YAML tab -->
      <div v-if="tab === 'yaml'">
        <textarea
          v-model="rawContent"
          class="w-full h-[calc(100vh-22rem)] font-mono text-sm bg-slate-900 border border-slate-700 rounded-xl p-4 text-slate-200 resize-none focus:outline-none focus:ring-2 focus:ring-sky-500/50 focus:border-sky-700 leading-relaxed"
          spellcheck="false"
          autocomplete="off"
        />
      </div>
    </template>

    <!-- Load error (not a 404 new-file case) -->
    <div v-else-if="store.error" class="card border-red-800/50 bg-red-900/10 text-red-400 text-sm">
      {{ store.error }}
    </div>

    <!-- Delete confirmation modal -->
    <div v-if="confirmDelete" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div class="w-full max-w-sm bg-slate-800 border border-slate-700 rounded-2xl p-6 shadow-2xl">
        <h3 class="text-base font-semibold text-slate-100 mb-2">Delete {{ filename }}?</h3>
        <p class="text-sm text-slate-400 mb-6">This will permanently remove the file. Traefik will stop routing traffic through this config.</p>
        <div class="flex gap-3 justify-end">
          <button class="btn btn-secondary" @click="confirmDelete = false">Cancel</button>
          <button class="btn bg-red-600 hover:bg-red-500 text-white focus:ring-red-500" @click="doDelete">
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDynamicStore } from '@/stores/dynamic'

const route = useRoute()
const router = useRouter()
const store = useDynamicStore()

const filename = computed(() => route.params.file as string)
const tab = ref<'form' | 'yaml'>('form')
const rawContent = ref('')
const confirmDelete = ref(false)
const saveMsg = ref<{ ok: boolean; text: string } | null>(null)

// Editable form fields for the simple pattern.
const formData = ref<{
  hostname: string
  backendUrl: string
  insecureBackend: boolean
} | null>(null)

// Regex to extract Host(`...`) from a router rule.
const hostRe = /Host\(`([^`]+)`\)/

// Detect simple pattern: 1 or 2 Host() routers (websecure + websecure-internal) + 1 LB service.
// New files are always considered simple.
const isSimplePattern = computed(() => {
  if (store.isNewFile) return true
  const p = store.currentFile?.parsed
  if (!p?.http) return false
  const routers = Object.values(p.http.routers ?? {})
  const services = Object.values(p.http.services ?? {})
  if (routers.length === 0 || routers.length > 2) return false
  // All routers must use Host() and the same hostname.
  if (!routers.every(r => hostRe.test(r.rule ?? ''))) return false
  const hosts = routers.map(r => hostRe.exec(r.rule ?? '')?.[1])
  if (new Set(hosts).size !== 1) return false
  // Allow api@internal (no services section) or 1 LB service.
  if (services.length > 1) return false
  return true
})

// Populate form when file loads or when entering create mode.
watch([() => store.currentFile, () => store.isNewFile], ([file, isNew]) => {
  if (isNew) {
    rawContent.value = ''
    formData.value = { hostname: '', backendUrl: '', insecureBackend: false }
    return
  }
  if (!file) return
  rawContent.value = file.raw

  const p = file.parsed
  if (!p?.http) return

  // Prefer the router with TLS (websecure) for reading certResolver; fall back to first.
  const allRouters = Object.values(p.http.routers ?? {})
  const tlsRouter = allRouters.find(r => r.tls) ?? allRouters[0]
  const service_ = Object.values(p.http.services ?? {})[0]

  const hostMatch = hostRe.exec(tlsRouter?.rule ?? '')
  const hasInsecure = Object.values(p.http.serversTransports ?? {}).some(t => t.insecureSkipVerify)

  formData.value = {
    hostname: hostMatch?.[1] ?? '',
    backendUrl: service_?.loadBalancer?.servers?.[0]?.url ?? '',
    insecureBackend: hasInsecure,
  }
}, { immediate: true })

async function save() {
  saveMsg.value = null
  let content = rawContent.value

  // Derive the target filename: for new files, use first hostname segment.
  let targetFilename = filename.value
  if (store.isNewFile) {
    const hostname = formData.value?.hostname ?? ''
    const base = hostname.split('.')[0].replace(/[^a-zA-Z0-9\-_]/g, '-').replace(/^-+|-+$/g, '') || 'service'
    targetFilename = base + '.yml'
  }

  // If saving from form tab and simple pattern — rebuild from form fields.
  if (tab.value === 'form' && isSimplePattern.value && formData.value) {
    const fd = formData.value
    const name = targetFilename.replace(/\.yml(\.bak)?$/, '')
    const svcName = name + '-svc'
    content = buildYaml({ name, hostname: fd.hostname, backendUrl: fd.backendUrl, insecureBackend: fd.insecureBackend }, svcName)
  }

  const wasNew = store.isNewFile
  const ok = await store.saveFile(targetFilename, content)
  if (ok) {
    if (wasNew) {
      // Update the URL to the real filename without adding a history entry.
      router.replace({ name: 'dynamic-editor', params: { file: targetFilename } })
    }
    await store.fetchFile(targetFilename)
    saveMsg.value = { ok: true, text: wasNew ? 'Created successfully.' : 'Saved successfully.' }
    setTimeout(() => { saveMsg.value = null }, 3000)
  } else {
    saveMsg.value = { ok: false, text: store.error ?? 'Save failed.' }
  }
}

function buildYaml(spec: {
  name: string; hostname: string; backendUrl: string; insecureBackend: boolean
}, svcName: string): string {
  const insecureTransport = spec.insecureBackend
    ? `  serversTransports:\n    insecure-skip:\n      insecureSkipVerify: true\n\n`
    : ''
  const transport = spec.insecureBackend
    ? `        serversTransport: insecure-skip\n`
    : ''

  return `http:
${insecureTransport}  routers:
    ${spec.name}:
      rule: "Host(\`${spec.hostname}\`)"
      entryPoints:
        - websecure
      service: ${svcName}
      tls: {}

    ${spec.name}-internal:
      rule: "Host(\`${spec.hostname}\`)"
      entryPoints:
        - websecure-internal
      service: ${svcName}

  services:
    ${svcName}:
      loadBalancer:
${transport}        servers:
          - url: "${spec.backendUrl}"
        passHostHeader: true
`
}

async function doDelete() {
  const ok = await store.deleteFile(filename.value)
  if (ok) {
    router.push({ name: 'dynamic' })
  }
  confirmDelete.value = false
}

onMounted(() => store.fetchFile(filename.value))
</script>
