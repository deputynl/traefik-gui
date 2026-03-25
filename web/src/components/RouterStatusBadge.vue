<template>
  <span class="badge" :class="cls">{{ status }}</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
const props = defineProps<{ status: string }>()
const cls = computed(() => {
  switch (props.status) {
    case 'enabled':  return 'badge-green'
    case 'warning':  return 'badge-yellow'
    case 'error':    return 'badge-red'
    default:         return 'bg-slate-700 text-slate-400 border border-slate-600'
  }
})
</script>
