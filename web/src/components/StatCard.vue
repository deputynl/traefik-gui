<template>
  <div class="card flex flex-col gap-1.5">
    <div class="text-xs text-slate-500 uppercase tracking-wider font-medium">{{ label }}</div>
    <div class="flex items-end gap-2 mt-0.5">
      <template v-if="count !== undefined">
        <span class="text-2xl font-bold text-slate-100">{{ count }}</span>
        <div class="flex gap-1.5 mb-0.5">
          <span v-if="warn" class="badge badge-yellow text-xs">{{ warn }}w</span>
          <span v-if="err"  class="badge badge-red    text-xs">{{ err }}e</span>
        </div>
      </template>
      <template v-else>
        <span class="w-2 h-2 rounded-full flex-shrink-0"
          :class="ok ? 'bg-emerald-400' : 'bg-red-500'" />
        <span class="text-sm" :class="ok ? 'text-emerald-400' : 'text-red-400'">
          {{ text ?? (ok ? 'online' : 'offline') }}
        </span>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  ok?: boolean
  text?: string
  count?: number
  warn?: number
  err?: number
}>()
</script>
