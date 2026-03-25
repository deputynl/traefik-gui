<template>
  <div class="p-8">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-slate-100">Audit Log</h1>
        <p class="text-slate-400 mt-1 text-sm">Recent configuration changes</p>
      </div>
      <button class="btn-primary" @click="auditStore.fetchAudit()">
        <svg class="w-4 h-4 mr-1.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h5M20 20v-5h-5M4 9a9 9 0 0115.45-3M20 15a9 9 0 01-15.45 3"/>
        </svg>
        Refresh
      </button>
    </div>

    <div v-if="auditStore.loading" class="card text-center py-12">
      <svg class="w-8 h-8 mx-auto animate-spin text-sky-500" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
      </svg>
    </div>

    <div v-else-if="!auditStore.entries.length" class="card text-center py-12">
      <p class="text-slate-400">No audit log entries yet.</p>
      <p class="text-slate-500 text-sm mt-1">Changes made via the GUI will appear here.</p>
    </div>

    <div v-else class="card overflow-hidden p-0">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-slate-700 text-xs uppercase tracking-wider text-slate-500">
            <th class="text-left px-4 py-3">Time</th>
            <th class="text-left px-4 py-3 hidden sm:table-cell">User</th>
            <th class="text-left px-4 py-3">Action</th>
            <th class="text-left px-4 py-3 hidden md:table-cell">Detail</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(entry, i) in auditStore.entries" :key="i"
            class="border-b border-slate-700/50 last:border-0 hover:bg-slate-700/20 transition-colors">
            <td class="px-4 py-3 text-xs text-slate-400 whitespace-nowrap">
              {{ new Date(entry.time).toLocaleString() }}
            </td>
            <td class="px-4 py-3 text-slate-300 hidden sm:table-cell">{{ entry.user }}</td>
            <td class="px-4 py-3">
              <span class="badge" :class="actionClass(entry.action)">{{ actionLabel(entry.action) }}</span>
            </td>
            <td class="px-4 py-3 text-slate-400 font-mono text-xs hidden md:table-cell">{{ entry.detail }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuditStore } from '@/stores/audit'

const auditStore = useAuditStore()

function actionLabel(action: string): string {
  switch (action) {
    case 'save_static_config':   return 'static config saved'
    case 'save_dynamic_config':  return 'dynamic file saved'
    case 'create_dynamic_config': return 'dynamic file created'
    case 'delete_dynamic_config': return 'dynamic file deleted'
    default: return action
  }
}

function actionClass(action: string): string {
  if (action.includes('delete')) return 'badge-red'
  if (action.includes('create')) return 'badge-green'
  return 'badge-sky'
}

onMounted(() => auditStore.fetchAudit())
</script>
