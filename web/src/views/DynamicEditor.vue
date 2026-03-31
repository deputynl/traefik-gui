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
        <h1 class="text-xl font-bold text-slate-100 truncate">{{ filename }}</h1>
        <p class="text-xs text-slate-500 mt-0.5">{{ store.currentFile ? 'Dynamic config file' : '…' }}</p>
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
          {{ store.saving ? 'Saving…' : 'Save' }}
        </button>
        <button class="btn btn-secondary text-red-400 hover:text-red-300 hover:bg-red-900/20" @click="confirmDelete = true">
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

    <template v-else-if="store.currentFile">
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
              <input v-model="formData.hostname" type="text" class="input w-full" />
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1.5">Entry points</label>
              <input v-model="formData.entryPointsStr" type="text" class="input w-full" placeholder="websecure" />
              <p class="text-xs text-slate-500 mt-1">Comma-separated</p>
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1.5">Certificate resolver</label>
              <input v-model="formData.certResolver" type="text" class="input w-full" />
            </div>
          </div>

          <div class="card space-y-4">
            <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wider">Service</h2>

            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1.5">Backend URL</label>
              <input v-model="formData.backendUrl" type="text" class="input w-full" />
            </div>

            <label class="flex items-center gap-3 cursor-pointer select-none">
              <div class="relative">
                <input v-model="formData.insecureBackend" type="checkbox" class="sr-only peer" />
                <div class="w-9 h-5 bg-slate-700 peer-checked:bg-sky-500 rounded-full transition-colors"></div>
                <div class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform peer-checked:translate-x-4"></div>
              </div>
              <div>
                <span class="text-sm text-slate-300">Skip TLS verification</span>
                <p class="text-xs text-slate-500">Adds <span class="font-mono">insecureSkipVerify: true</span></p>
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
  certResolver: string
  entryPointsStr: string
  insecureBackend: boolean
} | null>(null)

// Regex to extract Host(`...`) from a router rule.
const hostRe = /Host\(`([^`]+)`\)/

// Detect simple pattern: single Host() router + loadBalancer service.
const isSimplePattern = computed(() => {
  const p = store.currentFile?.parsed
  if (!p?.http) return false
  const routers = Object.values(p.http.routers ?? {})
  const services = Object.values(p.http.services ?? {})
  if (routers.length !== 1) return false
  const r = routers[0]
  if (!hostRe.test(r.rule ?? '')) return false
  // Allow api@internal (no services section) or 1 LB service.
  if (services.length > 1) return false
  return true
})

// Populate form when file loads.
watch(() => store.currentFile, (file) => {
  if (!file) return
  rawContent.value = file.raw

  const p = file.parsed
  if (!p?.http) return

  const router_ = Object.values(p.http.routers ?? {})[0]
  const service_ = Object.values(p.http.services ?? {})[0]

  const hostMatch = hostRe.exec(router_?.rule ?? '')
  const hasInsecure = Object.values(p.http.serversTransports ?? {}).some(t => t.insecureSkipVerify)

  formData.value = {
    hostname: hostMatch?.[1] ?? '',
    backendUrl: service_?.loadBalancer?.servers?.[0]?.url ?? '',
    certResolver: router_?.tls?.certResolver ?? '',
    entryPointsStr: (router_?.entryPoints ?? []).join(', '),
    insecureBackend: hasInsecure,
  }
}, { immediate: true })

// Sync YAML tab edits back so save always uses rawContent.
// (Form save rebuilds rawContent via the API.)

async function save() {
  saveMsg.value = null
  let content = rawContent.value

  // If saving from form tab and simple pattern — rebuild from form fields.
  if (tab.value === 'form' && isSimplePattern.value && formData.value) {
    const fd = formData.value
    const name = filename.value.replace(/\.yml(\.bak)?$/, '')
    const svcName = name + '-svc'
    const t = true

    const spec = {
      name,
      hostname: fd.hostname,
      backendUrl: fd.backendUrl,
      certResolver: fd.certResolver,
      insecureBackend: fd.insecureBackend,
      entryPoints: fd.entryPointsStr.split(',').map(s => s.trim()).filter(Boolean),
    }

    // Ask the server to generate clean YAML for us by creating a temp spec.
    // Instead, regenerate via a PUT with a JSON spec — but our PUT endpoint
    // expects raw YAML. So we build the YAML here client-side using a
    // template that matches the established file pattern.
    content = buildYaml(spec, svcName, t)
  }

  const ok = await store.saveFile(filename.value, content)
  if (ok) {
    // Refresh raw content from server.
    await store.fetchFile(filename.value)
    saveMsg.value = { ok: true, text: 'Saved successfully.' }
    setTimeout(() => { saveMsg.value = null }, 3000)
  } else {
    saveMsg.value = { ok: false, text: store.error ?? 'Save failed.' }
  }
}

function buildYaml(spec: {
  name: string; hostname: string; backendUrl: string
  certResolver: string; insecureBackend: boolean; entryPoints: string[]
}, svcName: string, _passHost: boolean): string {
  const eps = spec.entryPoints.map(e => `        - ${e}`).join('\n')
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
${eps}
      service: ${svcName}
      tls:
        certResolver: ${spec.certResolver}

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
