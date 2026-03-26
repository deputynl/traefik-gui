<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex-shrink-0 px-8 pt-8 pb-4">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-100">Activity</h1>
          <p class="text-slate-400 mt-1 text-sm">Access log · refreshes every 3 seconds</p>
        </div>
        <div class="flex items-center gap-2">
          <div class="flex items-center gap-1.5 text-xs text-slate-400 mr-2">
            <span class="w-2 h-2 rounded-full flex-shrink-0"
              :class="streamState === 'connected' ? 'bg-emerald-400 animate-pulse' : streamState === 'error' ? 'bg-red-500' : 'bg-slate-500'" />
            {{ streamState === 'connected' ? 'live' : streamState === 'error' ? 'error' : 'connecting…' }}
          </div>
          <button class="btn-secondary text-xs" @click="togglePause">
            {{ paused ? '▶ Resume' : '⏸ Pause' }}
          </button>
          <button class="btn-secondary text-xs" @click="clearEntries">Clear</button>
          <button class="btn-secondary text-xs flex items-center gap-1.5 transition-colors"
            :class="showSelf ? 'border-sky-600 bg-sky-600/20 text-sky-300' : ''"
            @click="showSelf = !showSelf"
            :title="'Toggle traffic from ' + selfHost">
            <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="showSelf ? 'bg-sky-400' : 'bg-slate-600'" />
            GUI traffic
          </button>
        </div>
      </div>

      <!-- Not available -->
      <div v-if="!available && streamState !== 'connecting'" class="card mb-4 flex items-start gap-4">
        <svg class="w-5 h-5 text-yellow-500 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <div>
          <p class="text-slate-300 font-medium">{{ unavailableReason }}</p>
          <p class="text-slate-500 text-sm mt-1">
            Enable the access log and set a file path in
            <router-link to="/static" class="text-sky-400 hover:text-sky-300 transition-colors">Static Config → Logging</router-link>.
          </p>
        </div>
      </div>

      <!-- Filter bar -->
      <div v-if="available || entries.length" class="flex flex-wrap gap-2">
        <div class="flex rounded-lg overflow-hidden border border-slate-700 text-xs">
          <button v-for="f in statusFilters" :key="f.value"
            class="px-3 py-1.5 transition-colors"
            :class="statusFilter === f.value ? 'bg-sky-600 text-white' : 'bg-slate-800 text-slate-400 hover:bg-slate-700'"
            @click="statusFilter = f.value">
            {{ f.label }}
          </button>
        </div>
        <div class="flex rounded-lg overflow-hidden border border-slate-700 text-xs">
          <button v-for="m in ['ALL','GET','POST','PUT','DELETE','OTHER']" :key="m"
            class="px-3 py-1.5 transition-colors"
            :class="methodFilter === m ? 'bg-sky-600 text-white' : 'bg-slate-800 text-slate-400 hover:bg-slate-700'"
            @click="methodFilter = m">
            {{ m }}
          </button>
        </div>
        <input v-model="search" class="input text-xs flex-1 min-w-40"
          placeholder="Filter by host, path, router, IP…" />
        <span class="text-xs text-slate-500 self-center">{{ filtered.length }} / {{ entries.length }}</span>
      </div>
    </div>

    <!-- Main area: log table + optional detail panel -->
    <div class="flex flex-1 min-h-0">
      <!-- Log feed -->
      <div class="flex-1 overflow-y-auto px-8 pb-8 min-h-0">
        <div v-if="!entries.length && streamState !== 'error'" class="text-center py-16 text-slate-500 text-sm">
          Waiting for requests…
        </div>
        <table v-else class="w-full text-xs font-mono">
          <thead class="sticky top-0 bg-slate-900 z-10">
            <tr class="text-left text-slate-500 border-b border-slate-700">
              <th class="py-2 pr-3 font-medium">Time</th>
              <th class="py-2 pr-3 font-medium">Method</th>
              <th class="py-2 pr-3 font-medium">Host + Path</th>
              <th class="py-2 pr-3 font-medium">Status</th>
              <th class="py-2 pr-3 font-medium">Duration</th>
              <th class="py-2 pr-3 font-medium hidden lg:table-cell">Router</th>
              <th class="py-2 font-medium hidden xl:table-cell">Client IP</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(entry, i) in filtered" :key="i"
              class="border-b border-slate-800 hover:bg-slate-800/50 transition-colors cursor-pointer"
              :class="selected === i ? 'bg-sky-900/20 border-sky-800/50' : ''"
              @click="selected = selected === i ? null : i">
              <td class="py-1.5 pr-3 text-slate-500 whitespace-nowrap">{{ formatTime(entry.time) }}</td>
              <td class="py-1.5 pr-3 font-bold" :class="methodColor(entry.method)">{{ entry.method }}</td>
              <td class="py-1.5 pr-3 max-w-xs">
                <span class="text-slate-400">{{ entry.host }}</span><span class="text-slate-200">{{ entry.path }}</span>
              </td>
              <td class="py-1.5 pr-3">
                <span class="px-1.5 py-0.5 rounded font-bold" :class="statusColor(entry.status)">{{ entry.status }}</span>
              </td>
              <td class="py-1.5 pr-3 text-slate-400 whitespace-nowrap">{{ formatDuration(entry.durationMs) }}</td>
              <td class="py-1.5 pr-3 text-slate-500 hidden lg:table-cell truncate max-w-[12rem]">{{ entry.routerName || '—' }}</td>
              <td class="py-1.5 text-slate-500 hidden xl:table-cell">{{ entry.clientIp || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Detail panel -->
      <transition name="slide">
        <div v-if="selectedEntry" class="w-80 flex-shrink-0 border-l border-slate-700 overflow-y-auto bg-slate-800/50">
          <div class="sticky top-0 bg-slate-800 border-b border-slate-700 px-4 py-3 flex items-center justify-between">
            <span class="text-xs font-semibold text-slate-300 uppercase tracking-wider">Request Detail</span>
            <button class="text-slate-500 hover:text-slate-300 transition-colors" @click="selected = null">✕</button>
          </div>
          <div class="p-4 space-y-3 text-xs">
            <!-- Status line -->
            <div class="flex items-center gap-2 pb-3 border-b border-slate-700">
              <span class="px-2 py-1 rounded font-bold text-sm" :class="statusColor(selectedEntry.status)">
                {{ selectedEntry.status }}
              </span>
              <span class="font-bold" :class="methodColor(selectedEntry.method)">{{ selectedEntry.method }}</span>
              <span class="text-slate-300 truncate">{{ selectedEntry.path }}</span>
            </div>

            <DetailRow label="Time"       :value="new Date(selectedEntry.time).toLocaleString()" />
            <DetailRow label="Host"       :value="selectedEntry.host" />
            <DetailRow label="Protocol"   :value="selectedEntry.protocol" v-if="selectedEntry.protocol" />
            <DetailRow label="Scheme"     :value="selectedEntry.scheme" v-if="selectedEntry.scheme" />
            <DetailRow label="Client IP"  :value="selectedEntry.clientIp" />

            <div class="border-t border-slate-700 pt-3 space-y-3">
              <DetailRow label="Entry Point"  :value="selectedEntry.entryPoint" v-if="selectedEntry.entryPoint" />
              <DetailRow label="Router"       :value="selectedEntry.routerName" v-if="selectedEntry.routerName" />
              <DetailRow label="Service"      :value="selectedEntry.serviceName" v-if="selectedEntry.serviceName" />
              <DetailRow label="Backend addr" :value="selectedEntry.serviceAddr" v-if="selectedEntry.serviceAddr" />
            </div>

            <div class="border-t border-slate-700 pt-3 space-y-3">
              <DetailRow label="Total duration"    :value="formatDuration(selectedEntry.durationMs)" />
              <DetailRow label="Origin duration"   :value="formatDuration(selectedEntry.originDurationMs)"
                v-if="selectedEntry.originDurationMs" />
              <DetailRow label="Origin status"     :value="String(selectedEntry.originStatus)"
                v-if="selectedEntry.originStatus && selectedEntry.originStatus !== selectedEntry.status" />
              <DetailRow label="Retries"           :value="String(selectedEntry.retryAttempts)"
                v-if="selectedEntry.retryAttempts" />
            </div>

            <div class="border-t border-slate-700 pt-3 space-y-3">
              <DetailRow label="Response size" :value="formatBytes(selectedEntry.responseSize)"
                v-if="selectedEntry.responseSize" />
              <DetailRow label="Request size"  :value="formatBytes(selectedEntry.requestSize)"
                v-if="selectedEntry.requestSize" />
            </div>

            <div v-if="selectedEntry.tlsVersion" class="border-t border-slate-700 pt-3 space-y-3">
              <DetailRow label="TLS version" :value="selectedEntry.tlsVersion" />
              <DetailRow label="Cipher"      :value="selectedEntry.tlsCipher" v-if="selectedEntry.tlsCipher" />
            </div>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, defineComponent, h } from 'vue'

interface LogEntry {
  time: string
  method: string
  host: string
  path: string
  protocol: string
  scheme: string
  status: number
  durationMs: number
  clientIp: string
  routerName: string
  serviceName: string
  serviceAddr: string
  entryPoint: string
  originStatus: number
  originDurationMs: number
  retryAttempts: number
  responseSize: number
  requestSize: number
  tlsVersion: string
  tlsCipher: string
}

// Tiny helper component to avoid repetition in the detail panel.
const DetailRow = defineComponent({
  props: { label: String, value: String },
  setup(props) {
    return () => h('div', { class: 'flex gap-2' }, [
      h('span', { class: 'text-slate-500 w-28 flex-shrink-0' }, props.label),
      h('span', { class: 'text-slate-300 break-all font-mono' }, props.value || '—'),
    ])
  },
})

const entries = ref<LogEntry[]>([])
const paused = ref(false)
const search = ref('')
const statusFilter = ref('ALL')
const methodFilter = ref('ALL')
const selected = ref<number | null>(null)
const streamState = ref<'connecting' | 'connected' | 'error'>('connecting')
const available = ref(true)
const unavailableReason = ref('')
const showSelf = ref(false)
const selfHost = window.location.hostname
const MAX_ENTRIES = 500

const statusFilters = [
  { label: 'All', value: 'ALL' },
  { label: '2xx', value: '2' },
  { label: '3xx', value: '3' },
  { label: '4xx', value: '4' },
  { label: '5xx', value: '5' },
]

const filtered = computed(() => {
  let list = entries.value
  if (!showSelf.value)
    list = list.filter(e => e.host !== selfHost)
  if (statusFilter.value !== 'ALL')
    list = list.filter(e => String(e.status).startsWith(statusFilter.value))
  if (methodFilter.value !== 'ALL') {
    if (methodFilter.value === 'OTHER')
      list = list.filter(e => !['GET','POST','PUT','DELETE'].includes(e.method))
    else
      list = list.filter(e => e.method === methodFilter.value)
  }
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(e =>
      e.host?.toLowerCase().includes(q) ||
      e.path?.toLowerCase().includes(q) ||
      e.routerName?.toLowerCase().includes(q) ||
      e.serviceName?.toLowerCase().includes(q) ||
      e.clientIp?.includes(q)
    )
  }
  return list
})

const selectedEntry = computed(() =>
  selected.value !== null ? filtered.value[selected.value] ?? null : null
)

let pollTimer: ReturnType<typeof setInterval> | null = null

async function loadRecent() {
  try {
    const res = await fetch('/api/accesslog')
    if (!res.ok) return
    const data = await res.json()
    available.value = data.available ?? false
    if (!data.available) {
      unavailableReason.value = data.reason ?? 'Access log not available.'
      streamState.value = 'error'
      return
    }
    entries.value = data.entries ?? []
    streamState.value = 'connected'
    startPolling()
  } catch {
    streamState.value = 'error'
  }
}

function startPolling() {
  if (pollTimer !== null) return
  pollTimer = setInterval(async () => {
    try {
      const res = await fetch('/api/accesslog')
      if (!res.ok) return
      const data = await res.json()
      if (!data.available) return
      const incoming: LogEntry[] = data.entries ?? []
      if (!incoming.length || paused.value) return
      // Find entries newer than what we already have.
      const newestKnown = entries.value[0]?.time ?? null
      const newEntries = newestKnown
        ? incoming.filter(e => e.time > newestKnown)
        : incoming
      if (!newEntries.length) return
      entries.value.unshift(...newEntries)
      if (entries.value.length > MAX_ENTRIES) entries.value.length = MAX_ENTRIES
      if (selected.value !== null) selected.value += newEntries.length
    } catch { /* ignore */ }
  }, 3000)
}

function togglePause() { paused.value = !paused.value }
function clearEntries() { entries.value = []; selected.value = null }

function formatTime(iso: string) {
  if (!iso) return '—'
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
function formatDuration(ms: number) {
  if (!ms) return '—'
  if (ms < 1) return `${(ms * 1000).toFixed(0)}µs`
  if (ms < 1000) return `${ms.toFixed(1)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}
function formatBytes(b: number) {
  if (!b) return '—'
  if (b < 1024) return `${b} B`
  if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`
  return `${(b / 1024 / 1024).toFixed(2)} MB`
}
function statusColor(status: number) {
  if (status >= 500) return 'bg-red-900/50 text-red-400'
  if (status >= 400) return 'bg-orange-900/50 text-orange-400'
  if (status >= 300) return 'bg-yellow-900/50 text-yellow-400'
  if (status >= 200) return 'bg-emerald-900/50 text-emerald-400'
  return 'bg-slate-700 text-slate-400'
}
function methodColor(method: string) {
  switch (method) {
    case 'GET':    return 'text-sky-400'
    case 'POST':   return 'text-emerald-400'
    case 'PUT':    return 'text-yellow-400'
    case 'DELETE': return 'text-red-400'
    default:       return 'text-slate-400'
  }
}

onMounted(() => loadRecent())
onUnmounted(() => { if (pollTimer !== null) clearInterval(pollTimer) })
</script>

<style scoped>
.slide-enter-active, .slide-leave-active { transition: width 0.2s ease, opacity 0.2s ease; }
.slide-enter-from, .slide-leave-to { width: 0; opacity: 0; overflow: hidden; }
</style>
