<template>
  <div class="card">
    <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wider mb-4">{{ title }}</h2>
    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string }>()
</script>
