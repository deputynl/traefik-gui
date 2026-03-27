<template>
  <RouterLink
    :to="to"
    :title="collapsed ? label : undefined"
    class="flex items-center rounded-lg text-sm transition-colors"
    :class="[
      collapsed ? 'justify-center px-2 py-2' : 'gap-2.5 px-3 py-2',
      isActive
        ? 'bg-sky-500/15 text-sky-400 font-medium'
        : 'text-slate-400 hover:text-slate-200 hover:bg-slate-700/60',
    ]"
  >
    <component :is="icon" />
    <template v-if="!collapsed">
      <span class="flex-1">{{ label }}</span>
      <span v-if="badge" class="text-xs font-bold px-1.5 py-0.5 rounded-full leading-none"
        :class="badge.color === 'red' ? 'bg-red-500/20 text-red-400' : 'bg-yellow-500/20 text-yellow-400'">
        {{ badge.text }}
      </span>
    </template>
    <span v-else-if="badge" class="absolute top-0.5 right-0.5 w-1.5 h-1.5 rounded-full"
      :class="badge.color === 'red' ? 'bg-red-400' : 'bg-yellow-400'" />
  </RouterLink>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const props = defineProps<{
  to: string
  icon: object
  label: string
  badge?: { text: string; color: 'red' | 'yellow' } | null
  collapsed?: boolean
}>()

const route = useRoute()
const isActive = computed(() =>
  props.to === '/' ? route.path === '/' : route.path.startsWith(props.to)
)
</script>
