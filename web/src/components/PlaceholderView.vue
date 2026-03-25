<template>
  <div class="card flex flex-col items-center justify-center text-center py-16 max-w-lg">
    <div class="w-14 h-14 rounded-2xl bg-sky-500/10 border border-sky-500/20 flex items-center justify-center mb-5">
      <svg v-if="icon === 'file'" class="w-7 h-7 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.6">
        <path stroke-linecap="round" stroke-linejoin="round"
          d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414A1 1 0 0121 9.414V19a2 2 0 01-2 2z"/>
      </svg>
      <svg v-else-if="icon === 'layers'" class="w-7 h-7 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.6">
        <path stroke-linecap="round" stroke-linejoin="round"
          d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
      </svg>
      <svg v-else-if="icon === 'cert'" class="w-7 h-7 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.6">
        <path stroke-linecap="round" stroke-linejoin="round"
          d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
      </svg>
      <svg v-else-if="icon === 'docker'" class="w-7 h-7 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.6">
        <path stroke-linecap="round" stroke-linejoin="round"
          d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z"/>
      </svg>
    </div>
    <h3 class="text-base font-semibold text-slate-200 mb-2">{{ title }}</h3>
    <p class="text-sm text-slate-400 leading-relaxed">{{ description }}</p>
    <span class="badge badge-sky mt-5">Coming soon</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description: string
  icon: 'file' | 'layers' | 'cert' | 'docker'
}>()
</script>
