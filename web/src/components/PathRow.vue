<template>
  <div class="flex items-center justify-between gap-4 py-2 border-b border-slate-700/60 last:border-0">
    <span class="text-sm text-slate-400 w-32 flex-shrink-0">{{ label }}</span>
    <div class="flex items-center gap-2 min-w-0">
      <span class="path-chip truncate max-w-xs" :title="path">{{ path }}</span>
      <span v-if="found === false" class="badge badge-yellow">not found</span>
      <span v-else-if="found === true" class="badge badge-green">found</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  path: string
  found?: boolean
}>()
</script>
