<template>
  <!-- Splash while checking session -->
  <div v-if="!auth.checked" class="min-h-screen bg-slate-900 flex items-center justify-center">
    <svg class="animate-spin w-8 h-8 text-sky-500" fill="none" viewBox="0 0 24 24">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
    </svg>
  </div>

  <!-- Login -->
  <Login v-else-if="!auth.user" @success="onLogin" />

  <!-- Main app shell -->
  <div v-else class="flex h-screen bg-slate-900 text-slate-100 overflow-hidden">
    <!-- Sidebar -->
    <aside class="flex-shrink-0 bg-slate-800 border-r border-slate-700 flex flex-col transition-all duration-200"
      :class="sidebarCollapsed ? 'w-14' : 'w-56'">
      <!-- Logo -->
      <div class="border-b border-slate-700 flex items-center"
        :class="sidebarCollapsed ? 'px-2 py-3 justify-center' : 'px-5 py-5'">
        <div class="flex items-center gap-2.5 min-w-0">
          <div class="w-7 h-7 rounded-lg bg-sky-500 flex items-center justify-center flex-shrink-0">
            <svg class="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"/>
            </svg>
          </div>
          <div v-if="!sidebarCollapsed" class="min-w-0">
            <div class="text-sm font-semibold text-slate-100 leading-tight">Traefik GUI</div>
            <div class="text-xs text-slate-500 leading-tight">v1.1.3</div>
          </div>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 py-4 space-y-0.5 overflow-y-auto" :class="sidebarCollapsed ? 'px-1.5' : 'px-3'">
        <NavItem to="/" :icon="IconDashboard" label="Dashboard" :collapsed="sidebarCollapsed" />
        <NavItem to="/static" :icon="IconFile" label="Static Config" :collapsed="sidebarCollapsed" />
        <NavItem to="/dynamic" :icon="IconLayers" label="Dynamic Config" :collapsed="sidebarCollapsed" />
        <NavItem to="/certificates" :icon="IconCert" label="Certificates" :badge="certBadge" :collapsed="sidebarCollapsed" />
        <NavItem to="/docker" :icon="IconDocker" label="Docker Labels" :collapsed="sidebarCollapsed" />
        <NavItem to="/activity" :icon="IconActivity" label="Activity" :collapsed="sidebarCollapsed" />
        <NavItem to="/mtls" :icon="IconMTLS" label="mTLS" :collapsed="sidebarCollapsed" />
        <NavItem to="/audit" :icon="IconAudit" label="Audit Log" :collapsed="sidebarCollapsed" />

        <!-- Collapse toggle -->
        <div class="pt-2 mt-2 border-t border-slate-700/50">
          <button
            class="w-full flex items-center rounded-lg text-sm text-slate-500 hover:text-slate-300 hover:bg-slate-700/60 transition-colors"
            :class="sidebarCollapsed ? 'justify-center px-2 py-2' : 'gap-2.5 px-3 py-2'"
            :title="sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
            @click="sidebarCollapsed = !sidebarCollapsed">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="w-4 h-4 flex-shrink-0 transition-transform duration-200"
              :class="sidebarCollapsed ? 'rotate-180' : ''">
              <path stroke-linecap="round" stroke-linejoin="round" d="M11 19l-7-7 7-7M18 19l-7-7 7-7"/>
            </svg>
            <span v-if="!sidebarCollapsed" class="text-xs">Collapse</span>
          </button>
        </div>
      </nav>

      <!-- Footer: status + user -->
      <div class="border-t border-slate-700" :class="sidebarCollapsed ? 'px-1.5 py-3 flex flex-col items-center gap-3' : 'px-4 py-4 space-y-3'">
        <div class="flex items-center gap-2 text-xs" :title="sidebarCollapsed ? (traefikOnline ? 'Traefik connected' : 'Traefik offline') : undefined">
          <span class="w-2 h-2 rounded-full flex-shrink-0"
            :class="traefikOnline ? 'bg-emerald-400 shadow-[0_0_6px_1px_rgba(52,211,153,0.5)]' : 'bg-red-500'" />
          <span v-if="!sidebarCollapsed" class="text-slate-400">
            Traefik
            <span :class="traefikOnline ? 'text-emerald-400' : 'text-red-400'">
              {{ traefikOnline ? 'connected' : 'offline' }}
            </span>
          </span>
        </div>
        <div class="flex items-center" :class="sidebarCollapsed ? 'justify-center' : 'justify-between'">
          <span v-if="!sidebarCollapsed" class="text-xs text-slate-500 truncate">{{ auth.user }}</span>
          <button class="text-slate-500 hover:text-slate-300 transition-colors flex-shrink-0"
            :title="sidebarCollapsed ? 'Sign out' : undefined"
            @click="logout">
            <span v-if="!sidebarCollapsed" class="text-xs">Sign out</span>
            <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="w-4 h-4">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
            </svg>
          </button>
        </div>
      </div>
    </aside>

    <!-- Main content -->
    <main class="flex-1 overflow-y-auto">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { h, ref, computed, onMounted, onUnmounted } from 'vue'
import { RouterView } from 'vue-router'
import NavItem from '@/components/NavItem.vue'
import Login from '@/views/Login.vue'
import { useAuthStore } from '@/stores/auth'
import { useConfigStore } from '@/stores/config'
import { useCertStore } from '@/stores/certs'

// Render-function icon factory — no Vue runtime compiler needed
const mkIcon = (children: () => ReturnType<typeof h>[]) => ({
  render: () => h('svg', { viewBox: '0 0 24 24', fill: 'none', stroke: 'currentColor', 'stroke-width': '1.8', class: 'w-4 h-4 flex-shrink-0' }, children()),
})
const p = (d: string, extra = {}) => h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', d, ...extra })
const pl = (points: string) => h('polyline', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', points })
const r = (attrs: Record<string, string | number>) => h('rect', attrs)

// Dashboard: 2×2 grid of rounded squares
const IconDashboard = mkIcon(() => [
  r({ x: 3, y: 3, width: 8, height: 8, rx: 1.5 }),
  r({ x: 13, y: 3, width: 8, height: 8, rx: 1.5 }),
  r({ x: 3, y: 13, width: 8, height: 8, rx: 1.5 }),
  r({ x: 13, y: 13, width: 8, height: 8, rx: 1.5 }),
])
// Static Config: equalizer — 3 horizontal lines with tick marks at different positions
const IconFile = mkIcon(() => [
  p('M4 6h16M4 12h16M4 18h16'),
  p('M8 4v4M15 10v4M11 16v4'),
])
// Dynamic Config: shuffle/routing — two crossing arrows
const IconLayers = mkIcon(() => [
  p('M16 3h5v5M4 20L21 3'),
  p('M16 21h5v-5M4 4l5 5'),
])
// Certificates: shield with checkmark
const IconCert = mkIcon(() => [
  p('M12 3L4 6.5v5c0 4.8 3.6 9.3 8 10.5 4.4-1.2 8-5.7 8-10.5v-5L12 3z'),
  pl('9 12 11 14 15 10'),
])
// Docker Labels: price tag
const IconDocker = mkIcon(() => [
  p('M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z'),
  p('M7 7h.01', { 'stroke-width': '2.5' }),
])
// Audit Log: document with list lines
const IconAudit = mkIcon(() => [
  p('M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2m-4 0a2 2 0 002 2h0a2 2 0 002-2m-4 0a2 2 0 012-2h0a2 2 0 012 2'),
  p('M9 12h6M9 16h4'),
])
// mTLS: padlock
const IconMTLS = mkIcon(() => [
  r({ x: 5, y: 11, width: 14, height: 10, rx: 2 }),
  p('M8 11V7a4 4 0 018 0v4'),
  p('M12 16h.01', { 'stroke-width': '2.5' }),
])
// Activity: pulse waveform
const IconActivity = mkIcon(() => [
  pl('22 12 18 12 15 21 9 3 6 12 2 12'),
])

const sidebarCollapsed = ref(false)

const auth = useAuthStore()
const configStore = useConfigStore()
const certStore = useCertStore()

const traefikOnline = computed(() => configStore.status?.traefik ?? false)

const certBadge = computed(() => {
  const expired = certStore.expired.length
  const soon = certStore.expiringSoon.length
  if (expired) return { text: String(expired), color: 'red' }
  if (soon) return { text: String(soon), color: 'yellow' }
  return null
})

let pollTimer: ReturnType<typeof setInterval>

function onLogin() {
  configStore.fetchStatus()
  certStore.fetchCerts()
  pollTimer = setInterval(() => {
    configStore.fetchStatus()
    certStore.fetchCerts()
  }, 60_000)
}

async function logout() {
  clearInterval(pollTimer)
  await auth.logout()
}

onMounted(async () => {
  const ok = await auth.check()
  if (ok) onLogin()
})

onUnmounted(() => clearInterval(pollTimer))
</script>
