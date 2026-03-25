<template>
  <div class="p-8">
    <div class="mb-8 flex items-start justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-100">Dashboard</h1>
        <p class="text-slate-400 mt-1 text-sm">
          <template v-if="traefik.lastFetched">
            Updated {{ timeAgo(traefik.lastFetched) }}
          </template>
          <template v-else>Overview of your Traefik setup</template>
        </p>
      </div>
      <button class="btn btn-secondary" :disabled="traefik.loading" @click="refresh">
        <svg class="w-4 h-4" :class="{ 'animate-spin': traefik.loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
        </svg>
        Refresh
      </button>
    </div>

    <!-- ── Live status row (when Traefik API connected) ── -->
    <template v-if="traefik.online">
      <div class="grid grid-cols-2 gap-4 mb-6 sm:grid-cols-4">
        <StatCard label="Routers" :count="traefik.routerCounts.total"
          :warn="traefik.routerCounts.warnings" :err="traefik.routerCounts.errors" />
        <StatCard label="Services" :count="traefik.serviceCounts.total"
          :err="traefik.serviceCounts.errors" />
        <StatCard label="Middlewares" :count="traefik.middlewareCounts.total" />
        <div class="card flex flex-col gap-1.5">
          <div class="text-xs text-slate-500 uppercase tracking-wider font-medium">Traefik API</div>
          <div class="flex items-center gap-2 mt-0.5">
            <span class="w-2 h-2 rounded-full bg-emerald-400 shadow-[0_0_6px_1px_rgba(52,211,153,0.4)]" />
            <span class="text-sm text-emerald-400">connected</span>
          </div>
        </div>
      </div>

      <!-- Router table -->
      <div class="card mb-6 overflow-hidden p-0">
        <div class="px-5 py-4 border-b border-slate-700 flex items-center justify-between">
          <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wider">HTTP Routers</h2>
          <span class="text-xs text-slate-500">{{ traefik.routers.length }} routers</span>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-slate-700/60">
                <th class="text-left px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Name</th>
                <th class="text-left px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Rule</th>
                <th class="text-left px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Service</th>
                <th class="text-left px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">TLS</th>
                <th class="text-left px-5 py-3 text-xs font-medium text-slate-500 uppercase tracking-wider">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in sortedRouters" :key="r.name"
                class="border-b border-slate-700/30 last:border-0 hover:bg-slate-700/20 transition-colors">
                <td class="px-5 py-3">
                  <div class="font-medium text-slate-200 truncate max-w-[12rem]">{{ shortName(r.name) }}</div>
                  <div class="text-xs text-slate-500">{{ r.provider }}</div>
                </td>
                <td class="px-5 py-3 font-mono text-xs text-slate-300 max-w-[16rem]">
                  <span class="truncate block" :title="r.rule">{{ r.rule }}</span>
                </td>
                <td class="px-5 py-3 text-slate-400 text-xs font-mono truncate max-w-[10rem]">
                  {{ r.service }}
                </td>
                <td class="px-5 py-3">
                  <span v-if="r.tls?.certResolver" class="badge badge-sky">{{ r.tls.certResolver }}</span>
                  <span v-else-if="r.tls" class="badge badge-sky">tls</span>
                  <span v-else class="text-slate-600 text-xs">—</span>
                </td>
                <td class="px-5 py-3">
                  <RouterStatusBadge :status="r.status" />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <!-- ── Traefik API offline banner ── -->
    <div v-else class="card border-slate-700 mb-6 flex items-start gap-3">
      <span class="w-2 h-2 rounded-full bg-red-500 flex-shrink-0 mt-1.5" />
      <div>
        <p class="text-sm font-medium text-slate-300">Traefik API not reachable</p>
        <p class="text-xs text-slate-500 mt-0.5">
          Expected at <span class="font-mono text-slate-400">{{ configStore.appConfig?.traefikApiUrl }}</span>.
          Live router data unavailable.
        </p>
      </div>
    </div>

    <!-- ── Config paths ── -->
    <div class="card mb-6" v-if="configStore.appConfig">
      <h2 class="text-sm font-semibold text-slate-300 mb-4 uppercase tracking-wider">Configuration Paths</h2>
      <div class="space-y-1">
        <PathRow label="Static Config" :path="configStore.appConfig.paths.staticConfig"
          :found="configStore.appConfig.paths.staticConfigFound" />
        <PathRow label="Dynamic Dir"  :path="configStore.appConfig.paths.dynamicDir"  :found="configStore.appConfig.paths.dynamicDirFound" />
        <PathRow label="ACME Storage" :path="configStore.appConfig.paths.acmePath"   :found="configStore.appConfig.paths.acmePathFound" />
        <PathRow label="Traefik API"  :path="configStore.appConfig.traefikApiUrl" :found="traefik.online" />
      </div>
    </div>

    <!-- ── Static config summary ── -->
    <div v-if="configStore.appConfig?.staticConfig" class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <!-- Entry points -->
      <div v-if="epCount" class="card">
        <h2 class="text-sm font-semibold text-slate-300 mb-3 uppercase tracking-wider">Entry Points</h2>
        <div class="space-y-1.5">
          <div v-for="(ep, name) in configStore.appConfig.staticConfig.entryPoints" :key="name"
            class="flex items-center justify-between py-1.5 border-b border-slate-700/50 last:border-0">
            <span class="text-slate-300 text-sm font-medium">{{ name }}</span>
            <span class="path-chip text-xs">{{ ep.address }}</span>
          </div>
        </div>
      </div>

      <!-- Providers + resolvers -->
      <div class="card">
        <h2 class="text-sm font-semibold text-slate-300 mb-3 uppercase tracking-wider">Providers</h2>
        <div class="flex flex-wrap gap-2 mb-4">
          <span v-if="configStore.appConfig.staticConfig.providers?.docker" class="badge badge-sky">Docker</span>
          <span v-if="configStore.appConfig.staticConfig.providers?.file" class="badge badge-sky">File</span>
        </div>
        <h2 class="text-sm font-semibold text-slate-300 mb-3 uppercase tracking-wider">Cert Resolvers</h2>
        <div class="flex flex-wrap gap-2">
          <span v-for="(_, name) in configStore.appConfig.staticConfig.certificatesResolvers" :key="name"
            class="badge badge-green">{{ name }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useConfigStore } from '@/stores/config'
import { useTraefikStore } from '@/stores/traefik'
import StatCard from '@/components/StatCard.vue'
import PathRow from '@/components/PathRow.vue'
import RouterStatusBadge from '@/components/RouterStatusBadge.vue'

const configStore = useConfigStore()
const traefik = useTraefikStore()

const epCount = computed(() =>
  Object.keys(configStore.appConfig?.staticConfig?.entryPoints ?? {}).length)

const sortedRouters = computed(() =>
  [...traefik.routers].sort((a, b) => a.name.localeCompare(b.name)))

function shortName(name: string) {
  // "cockpit@file" → "cockpit"
  return name.split('@')[0]
}

function timeAgo(date: Date) {
  const s = Math.floor((Date.now() - date.getTime()) / 1000)
  if (s < 10) return 'just now'
  if (s < 60) return `${s}s ago`
  return `${Math.floor(s / 60)}m ago`
}

let refreshTimer: ReturnType<typeof setInterval>

async function refresh() {
  await traefik.fetchAll()
}

onMounted(async () => {
  if (!configStore.appConfig) configStore.fetchConfig()
  await traefik.fetchAll()
  refreshTimer = setInterval(() => traefik.fetchAll(), 30_000)
})

onUnmounted(() => clearInterval(refreshTimer))
</script>
