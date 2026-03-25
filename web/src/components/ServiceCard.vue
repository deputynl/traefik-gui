<template>
  <div
    class="card group cursor-pointer hover:border-slate-600 transition-colors flex flex-col gap-3"
    :class="{ 'opacity-60': !file.active }"
    @click="$emit('edit')"
  >
    <!-- Header row -->
    <div class="flex items-start justify-between gap-2">
      <div class="flex items-center gap-2 min-w-0">
        <span
          class="w-2 h-2 rounded-full flex-shrink-0 mt-0.5"
          :class="file.active ? 'bg-emerald-400' : 'bg-slate-500'"
        />
        <span class="font-medium text-slate-200 text-sm truncate">{{ file.name }}</span>
      </div>
      <span v-if="!file.active" class="badge badge-yellow flex-shrink-0">inactive</span>
    </div>

    <!-- Hostnames -->
    <div v-if="file.hostnames?.length" class="flex flex-wrap gap-1.5">
      <span
        v-for="host in file.hostnames"
        :key="host"
        class="text-xs font-mono text-sky-300 bg-sky-900/30 border border-sky-800/50 px-2 py-0.5 rounded"
      >{{ host }}</span>
    </div>
    <div v-else-if="file.active" class="text-xs text-slate-500 italic">no Host() rule</div>

    <!-- Backend URLs -->
    <div v-if="file.backends?.length" class="flex flex-col gap-1">
      <div
        v-for="url in file.backends"
        :key="url"
        class="flex items-center gap-1.5 text-xs text-slate-400"
      >
        <svg class="w-3 h-3 flex-shrink-0 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6"/>
        </svg>
        <span class="font-mono truncate">{{ url }}</span>
      </div>
    </div>

    <!-- Badges row -->
    <div class="flex flex-wrap items-center gap-1.5 mt-auto pt-1 border-t border-slate-700/60">
      <span v-if="file.certResolver" class="badge badge-sky text-xs">{{ file.certResolver }}</span>
      <span v-if="file.insecureSkipVerify" class="badge badge-yellow text-xs">insecure</span>
      <span v-if="file.middlewareCount" class="badge badge-sky text-xs">{{ file.middlewareCount }} middleware{{ file.middlewareCount > 1 ? 's' : '' }}</span>
      <span class="ml-auto text-xs text-slate-600 group-hover:text-slate-500 transition-colors">Edit →</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FileSummary } from '@/stores/dynamic'

defineProps<{ file: FileSummary }>()
defineEmits<{ edit: [] }>()
</script>
