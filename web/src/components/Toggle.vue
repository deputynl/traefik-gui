<template>
  <label class="flex items-center gap-3 cursor-pointer select-none w-fit">
    <div class="relative flex-shrink-0">
      <input :checked="modelValue" type="checkbox" class="sr-only peer"
        @change="$emit('update:modelValue', ($event.target as HTMLInputElement).checked)" />
      <div
        class="rounded-full transition-colors bg-slate-700 peer-checked:bg-sky-500"
        :class="small ? 'w-8 h-4' : 'w-9 h-5'"
      ></div>
      <div
        class="absolute top-0.5 left-0.5 bg-white rounded-full transition-transform peer-checked:translate-x-4"
        :class="small ? 'w-3 h-3' : 'w-4 h-4'"
      ></div>
    </div>
    <span v-if="label" class="text-slate-300" :class="small ? 'text-xs' : 'text-sm'">{{ label }}</span>
  </label>
</template>

<script setup lang="ts">
defineProps<{ modelValue: boolean; label?: string; small?: boolean }>()
defineEmits<{ 'update:modelValue': [value: boolean] }>()
</script>
