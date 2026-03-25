<template>
  <RouterLink
    :to="to"
    class="flex items-center gap-2.5 px-3 py-2 rounded-lg text-sm transition-colors"
    :class="isActive
      ? 'bg-sky-500/15 text-sky-400 font-medium'
      : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700/60'"
  >
    <component :is="icon" />
    {{ label }}
  </RouterLink>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const props = defineProps<{
  to: string
  icon: object
  label: string
}>()

const route = useRoute()
const isActive = computed(() =>
  props.to === '/' ? route.path === '/' : route.path.startsWith(props.to)
)
</script>
