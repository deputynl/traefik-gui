<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <div class="flex-shrink-0 px-8 pt-8 pb-4">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-100">Activity</h1>
          <p class="text-slate-400 mt-1 text-sm">
            Live access log stream
            <span v-if="configStore.appConfig?.paths.accessLogPath" class="font-mono text-slate-500">
              · {{ configStore.appConfig.paths.accessLogPath }}
            </span>
          </p>
        </div>
        <div class="flex items-center gap-2">
          <!-- Connection status -->
          <div class="flex items-center gap-1.5 text-xs text-slate-400 mr-2">
            <span class="w-2 h-2 rounded-full flex-shrink-0"
              :class="streamState === 'connected' ? 'bg-emerald-400 animate-pulse' : streamState === 'error' ? 'bg-red-500' : 'bg-slate-500'" />
            {{ streamState === 'connected' ? 'live' : streamState === 'error' ? 'error' : 'connecting…' }}
          </div>
          <button class="btn-secondary text-xs" @click="togglePause">
            {{ paused ? '▶ Resume' : '⏸ Pause' }}
          </button>
          <button class="btn-secondary text-xs" @click="clearEntries">Clear</button>
        </div>
      </div>

      <!-- Not available -->
      <div v-if="!available && streamState !== 'connecting'" class="card mb-4">
        <p class="text-slate-300 font-medium mb-2">{{ unavailableReason }}</p>
        <p class="text-slate-500 text-sm mb-3">
          Add a <code class="text-sky-400">filePath</code> to your access log configuration so Traefik writes to a file instead of stdout:
        </p>
        <pre class="bg-slate-900 rounded-lg px-4 py-3 text-xs font-mono text-slate-300">accessLog:
  filePath: /var/log/traefik/access.log
  format: json   # json gives richer filtering; omit for CLF</pre>
      </div>

      <!-- Filter bar -->
      <div v-if="available || entries.length" class="flex flex-wrap gap-2">
        <!-- Status filter -->
        <div class="flex rounded-lg overflow-hidden border border-slate-700 text-xs">
          <button v-for="f in statusFilters" :key="f.value"
            class="px-3 py-1.5 transition-colors"
            :class="statusFilter === f.value ? 'bg-sky-600 text-white' : 'bg-slate-800 text-slate-400 hover:bg-slate-700'"
            @click="statusFilter = f.value">
            {{ f.label }}
          </button>
        </div>
        <!-- Method filter -->
        <div class="flex rounded-lg overflow-hidden border border-slate-700 text-xs">
          <button v-for="m in ['ALL','GET','POST','PUT','DELETE','OTHER']" :key="m"
            class="px-3 py-1.5 transition-colors"
            :class="methodFilter === m ? 'bg-sky-600 text-white' : 'bg-slate-800 text-slate-400 hover:bg-slate-700'"
            @click="methodFilter = m">
            {{ m }}
          </button>
        </div>
        <!-- Search -->
        <input v-model="search" class="input text-xs flex-1 min-w-40"
          placeholder="Filter by host, path, router, IP…" />
        <span class="text-xs text-slate-500 self-center">
          {{ filtered.length }} / {{ entries.length }}
        </span>
      </div>
    </div>

    <!-- Log feed -->
    <div ref="feedEl" class="flex-1 overflow-y-auto px-8 pb-8 min-h-0">
      <div v-if="!entries.length && streamState !== 'error'" class="text-center py-16 text-slate-500 text-sm">
        Waiting for requests…
      </div>

      <!-- Table -->
      <table v-else class="w-full text-xs font-mono">
        <thead class="sticky top-0 bg-slate-900 z-10">
          <tr class="text-left text-slate-500 border-b border-slate-700">
            <th class="py-2 pr-4 font-medium">Time</th>
            <th class="py-2 pr-4 font-medium">Method</th>
            <th class="py-2 pr-4 font-medium w-1/3">Host + Path</th>
            <th class="py-2 pr-4 font-medium">Status</th>
            <th class="py-2 pr-4 font-medium">Duration</th>
            <th class="py-2 pr-4 font-medium hidden lg:table-cell">Router</th>
            <th class="py-2 font-medium hidden xl:table-cell">Client IP</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(entry, i) in filtered" :key="i"
            class="border-b border-slate-800 hover:bg-slate-800/50 transition-colors cursor-default"
            @click="selected = selected === i ? null : i">
            <td class="py-1.5 pr-4 text-slate-500 whitespace-nowrap">
              {{ formatTime(entry.time) }}
            </td>
            <td class="py-1.5 pr-4">
              <span class="font-bold" :class="methodColor(entry.method)">{{ entry.method }}</span>
            </td>
            <td class="py-1.5 pr-4 max-w-xs">
              <span class="text-slate-400">{{ entry.host }}</span>
              <span class="text-slate-200">{{ entry.path }}</span>
            </td>
            <td class="py-1.5 pr-4">
              <span class="px-1.5 py-0.5 rounded font-bold" :class="statusColor(entry.status)">
                {{ entry.status }}
              </span>
            </td>
            <td class="py-1.5 pr-4 text-slate-400 whitespace-nowrap">
              {{ formatDuration(entry.durationMs) }}
            </td>
            <td class="py-1.5 pr-4 text-slate-500 hidden lg:table-cell truncate max-w-xs">
              {{ entry.routerName || '—' }}
            </td>
            <td class="py-1.5 text-slate-500 hidden xl:table-cell">{{ entry.clientIp || '—' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useConfigStore } from '@/stores/config'

interface LogEntry {
  time: string
  method: string
  host: string
  path: string
  status: number
  durationMs: number
  routerName: string
  clientIp: string
}

const configStore = useConfigStore()
const entries = ref<LogEntry[]>([])
const paused = ref(false)
const search = ref('')
const statusFilter = ref('ALL')
const methodFilter = ref('ALL')
const selected = ref<number | null>(null)
const feedEl = ref<HTMLElement | null>(null)
const streamState = ref<'connecting' | 'connected' | 'error'>('connecting')
const available = ref(true)
const unavailableReason = ref('')

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
  if (statusFilter.value !== 'ALL') {
    const prefix = statusFilter.value
    list = list.filter(e => String(e.status).startsWith(prefix))
  }
  if (methodFilter.value !== 'ALL') {
    if (methodFilter.value === 'OTHER') {
      list = list.filter(e => !['GET','POST','PUT','DELETE'].includes(e.method))
    } else {
      list = list.filter(e => e.method === methodFilter.value)
    }
  }
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(e =>
      e.host?.toLowerCase().includes(q) ||
      e.path?.toLowerCase().includes(q) ||
      e.routerName?.toLowerCase().includes(q) ||
      e.clientIp?.includes(q)
    )
  }
  return list
})

let es: EventSource | null = null
let atBottom = true

function startStream() {
  streamState.value = 'connecting'
  es = new EventSource('/api/accesslog/stream')

  es.onopen = () => {
    streamState.value = 'connected'
  }

  es.onmessage = (event) => {
    if (paused.value) return
    try {
      const entry: LogEntry = JSON.parse(event.data)
      entries.value.unshift(entry)
      if (entries.value.length > MAX_ENTRIES) {
        entries.value.length = MAX_ENTRIES
      }
    } catch { /* ignore malformed */ }
  }

  es.onerror = () => {
    streamState.value = 'error'
    available.value = false
    unavailableReason.value = 'Could not connect to access log stream.'
    es?.close()
  }
}

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
    startStream()
  } catch {
    streamState.value = 'error'
  }
}

function togglePause() { paused.value = !paused.value }
function clearEntries() { entries.value = []; selected.value = null }

function formatTime(iso: string): string {
  if (!iso) return '—'
  const d = new Date(iso)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function formatDuration(ms: number): string {
  if (ms === undefined || ms === null) return '—'
  if (ms < 1) return `${(ms * 1000).toFixed(0)}µs`
  if (ms < 1000) return `${ms.toFixed(1)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function statusColor(status: number): string {
  if (status >= 500) return 'bg-red-900/50 text-red-400'
  if (status >= 400) return 'bg-orange-900/50 text-orange-400'
  if (status >= 300) return 'bg-yellow-900/50 text-yellow-400'
  if (status >= 200) return 'bg-emerald-900/50 text-emerald-400'
  return 'bg-slate-700 text-slate-400'
}

function methodColor(method: string): string {
  switch (method) {
    case 'GET':    return 'text-sky-400'
    case 'POST':   return 'text-emerald-400'
    case 'PUT':    return 'text-yellow-400'
    case 'DELETE': return 'text-red-400'
    default:       return 'text-slate-400'
  }
}

onMounted(() => loadRecent())
onUnmounted(() => es?.close())
</script>
