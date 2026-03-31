<template>
  <div class="flex flex-col flex-1 min-h-0">
    <!-- Header -->
    <div class="flex-shrink-0 px-8 pt-8 pb-4">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-100">Logs</h1>
          <p class="text-slate-400 mt-1 text-sm">General Traefik log output · refreshes every 3 seconds</p>
        </div>
        <div class="flex items-center gap-2">
          <!-- Connection indicator -->
          <div class="flex items-center gap-1.5 text-xs text-slate-400 mr-2">
            <span class="w-2 h-2 rounded-full flex-shrink-0"
              :class="streamState === 'connected' ? 'bg-emerald-400 animate-pulse' : streamState === 'error' ? 'bg-red-500' : 'bg-slate-500'" />
            {{ streamState === 'connected' ? 'live' : streamState === 'error' ? 'error' : 'connecting…' }}
          </div>

          <!-- Log level control -->
          <div class="flex items-center gap-1.5">
            <select v-model="pendingLevel" class="input text-xs py-1 pr-7">
              <option v-for="lvl in LOG_LEVELS" :key="lvl" :value="lvl">{{ lvl }}</option>
            </select>
            <button v-if="pendingLevel !== currentLevel" class="btn btn-primary text-xs py-1"
              :disabled="applyingLevel" @click="applyLogLevel">
              {{ applyingLevel ? 'Restarting…' : 'Apply & Restart' }}
            </button>
            <span v-else class="text-xs text-slate-500">log level</span>
          </div>

          <button class="btn-secondary text-xs" @click="togglePause">
            {{ paused ? '▶ Resume' : '⏸ Pause' }}
          </button>
          <button class="btn-secondary text-xs" @click="clearEntries">Clear</button>
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
            Set <span class="font-mono text-slate-400">TRAEFIK_CONTAINER_NAME</span> to enable log streaming.
          </p>
        </div>
      </div>

      <!-- Filter bar -->
      <div v-if="available || lines.length" class="flex flex-wrap gap-2">
        <div class="flex rounded-lg overflow-hidden border border-slate-700 text-xs">
          <button v-for="lvl in ['ALL', ...LOG_LEVELS]" :key="lvl"
            class="px-3 py-1.5 transition-colors"
            :class="levelFilter === lvl ? 'bg-sky-600 text-white' : 'bg-slate-800 text-slate-400 hover:bg-slate-700'"
            @click="levelFilter = lvl">
            {{ lvl }}
          </button>
        </div>
        <input v-model="search" class="input text-xs flex-1 min-w-40"
          placeholder="Filter by message…" />
        <span class="text-xs text-slate-500 self-center">{{ filtered.length }} / {{ lines.length }}</span>
      </div>
    </div>

    <!-- Log feed -->
    <div class="flex-1 overflow-y-auto px-8 pb-8 min-h-0">
      <div v-if="!lines.length && streamState !== 'error'" class="text-center py-16 text-slate-500 text-sm">
        Waiting for log output…
      </div>
      <table v-else class="w-full text-xs font-mono">
        <thead class="sticky top-0 bg-slate-900 z-10">
          <tr class="text-left text-slate-500 border-b border-slate-700">
            <th class="py-2 pr-4 font-medium w-28">Time</th>
            <th class="py-2 pr-4 font-medium w-20">Level</th>
            <th class="py-2 font-medium">Message</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(line, i) in filtered" :key="i"
            class="border-b border-slate-800 hover:bg-slate-800/40 transition-colors">
            <td class="py-1.5 pr-4 text-slate-500 whitespace-nowrap align-top">{{ formatTime(line.time) }}</td>
            <td class="py-1.5 pr-4 align-top">
              <span v-if="line.level" class="px-1.5 py-0.5 rounded text-xs font-bold" :class="levelColor(line.level)">
                {{ line.level }}
              </span>
            </td>
            <td class="py-1.5 text-slate-300 break-all align-top">{{ line.msg || line.raw }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Feedback toast -->
    <div v-if="toast" class="fixed bottom-6 right-6 px-4 py-2.5 rounded-lg border text-sm shadow-xl z-50"
      :class="toast.ok
        ? 'bg-emerald-900/90 border-emerald-700 text-emerald-300'
        : 'bg-red-900/90 border-red-700 text-red-300'">
      {{ toast.text }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface LogLine {
  time: string
  level: string
  msg: string
  raw: string
}

const LOG_LEVELS = ['TRACE', 'DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL'] as const

const lines = ref<LogLine[]>([])
const paused = ref(false)
const search = ref('')
const levelFilter = ref('ALL')
const streamState = ref<'connecting' | 'connected' | 'error'>('connecting')
const available = ref(true)
const unavailableReason = ref('')
const currentLevel = ref('INFO')
const pendingLevel = ref('INFO')
const applyingLevel = ref(false)
const toast = ref<{ ok: boolean; text: string } | null>(null)
const MAX_LINES = 500

let pollTimer: ReturnType<typeof setInterval> | null = null

const filtered = computed(() => {
  let list = lines.value
  if (levelFilter.value !== 'ALL')
    list = list.filter(l => l.level === levelFilter.value)
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(l =>
      l.msg?.toLowerCase().includes(q) ||
      l.raw?.toLowerCase().includes(q)
    )
  }
  return list
})

async function loadRecent() {
  try {
    const res = await fetch('/api/tracinglog')
    if (!res.ok) return
    const data = await res.json()
    available.value = data.available ?? false
    if (!data.available) {
      unavailableReason.value = data.reason ?? 'Log streaming not available.'
      streamState.value = 'error'
      return
    }
    lines.value = data.lines ?? []
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
      const res = await fetch('/api/tracinglog')
      if (!res.ok) return
      const data = await res.json()
      if (!data.available || paused.value) return
      const incoming: LogLine[] = data.lines ?? []
      if (!incoming.length) return
      const newestKnown = lines.value[0]?.time ?? null
      const newLines = newestKnown
        ? incoming.filter(l => l.time > newestKnown)
        : incoming
      if (!newLines.length) return
      lines.value.unshift(...newLines)
      if (lines.value.length > MAX_LINES) lines.value.length = MAX_LINES
    } catch { /* ignore */ }
  }, 3000)
}

async function loadCurrentLevel() {
  try {
    const res = await fetch('/api/config')
    if (!res.ok) return
    const data = await res.json()
    const level = data.staticConfig?.log?.level ?? 'INFO'
    currentLevel.value = level.toUpperCase()
    pendingLevel.value = currentLevel.value
  } catch { /* ignore */ }
}

async function applyLogLevel() {
  if (!confirm(`This will update the log level to ${pendingLevel.value} and restart Traefik. Continue?`)) return
  applyingLevel.value = true
  try {
    const res = await fetch('/api/traefik/loglevel', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ level: pendingLevel.value }),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error)
    currentLevel.value = pendingLevel.value
    const msg = data.restarted
      ? `Log level set to ${pendingLevel.value} — Traefik restarted.`
      : `Log level set to ${pendingLevel.value} — restart Traefik to apply.`
    showToast(true, msg)
  } catch (e) {
    showToast(false, String(e))
  } finally {
    applyingLevel.value = false
  }
}

function togglePause() { paused.value = !paused.value }
function clearEntries() { lines.value = [] }

function showToast(ok: boolean, text: string) {
  toast.value = { ok, text }
  setTimeout(() => { toast.value = null }, 4000)
}

function formatTime(iso: string) {
  if (!iso) return '—'
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function levelColor(level: string) {
  switch (level) {
    case 'TRACE': return 'bg-slate-700 text-slate-300'
    case 'DEBUG': return 'bg-cyan-900/40 text-cyan-400'
    case 'INFO':  return 'bg-emerald-900/40 text-emerald-400'
    case 'WARN':  return 'bg-yellow-900/40 text-yellow-400'
    case 'ERROR': return 'bg-red-900/40 text-red-400'
    case 'FATAL': return 'bg-red-900/60 text-red-300'
    default:      return 'bg-slate-700 text-slate-400'
  }
}

onMounted(() => {
  loadRecent()
  loadCurrentLevel()
})
onUnmounted(() => { if (pollTimer !== null) clearInterval(pollTimer) })
</script>
