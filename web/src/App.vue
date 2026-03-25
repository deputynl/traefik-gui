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
    <aside class="w-56 flex-shrink-0 bg-slate-800 border-r border-slate-700 flex flex-col">
      <!-- Logo -->
      <div class="px-5 py-5 border-b border-slate-700">
        <div class="flex items-center gap-2.5">
          <div class="w-7 h-7 rounded-lg bg-sky-500 flex items-center justify-center flex-shrink-0">
            <svg class="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"/>
            </svg>
          </div>
          <div>
            <div class="text-sm font-semibold text-slate-100 leading-tight">Traefik GUI</div>
            <div class="text-xs text-slate-500 leading-tight">v0.1.0</div>
          </div>
        </div>
      </div>

      <!-- Navigation -->
      <nav class="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
        <NavItem to="/" :icon="IconDashboard" label="Dashboard" />
        <NavItem to="/static" :icon="IconFile" label="Static Config" />
        <NavItem to="/dynamic" :icon="IconLayers" label="Dynamic Config" />
        <NavItem to="/certificates" :icon="IconCert" label="Certificates" :badge="certBadge" />
        <NavItem to="/docker" :icon="IconDocker" label="Docker Labels" />
        <NavItem to="/activity" :icon="IconActivity" label="Activity" />
        <NavItem to="/audit" :icon="IconAudit" label="Audit Log" />
      </nav>

      <!-- Footer: status + user -->
      <div class="px-4 py-4 border-t border-slate-700 space-y-3">
        <div class="flex items-center gap-2 text-xs">
          <span class="w-2 h-2 rounded-full flex-shrink-0"
            :class="traefikOnline ? 'bg-emerald-400 shadow-[0_0_6px_1px_rgba(52,211,153,0.5)]' : 'bg-red-500'" />
          <span class="text-slate-400">
            Traefik
            <span :class="traefikOnline ? 'text-emerald-400' : 'text-red-400'">
              {{ traefikOnline ? 'connected' : 'offline' }}
            </span>
          </span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-xs text-slate-500 truncate">{{ auth.user }}</span>
          <button class="text-xs text-slate-500 hover:text-slate-300 transition-colors" @click="logout">
            Sign out
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
import { computed, onMounted, onUnmounted } from 'vue'
import { RouterView } from 'vue-router'
import NavItem from '@/components/NavItem.vue'
import Login from '@/views/Login.vue'
import { useAuthStore } from '@/stores/auth'
import { useConfigStore } from '@/stores/config'
import { useCertStore } from '@/stores/certs'

const IconDashboard = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="w-4 h-4"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>` }
const IconFile = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="w-4 h-4"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414A1 1 0 0121 9.414V19a2 2 0 01-2 2z"/></svg>` }
const IconLayers = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="w-4 h-4"><path stroke-linecap="round" stroke-linejoin="round" d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>` }
const IconCert = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="w-4 h-4"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/></svg>` }
const IconDocker = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="w-4 h-4"><path stroke-linecap="round" stroke-linejoin="round" d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>` }
const IconAudit    = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="w-4 h-4"><path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4"/></svg>` }
const IconActivity = { template: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" class="w-4 h-4"><polyline stroke-linecap="round" stroke-linejoin="round" points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>` }

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
