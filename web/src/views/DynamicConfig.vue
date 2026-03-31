<template>
  <div class="p-8 flex-1 overflow-y-auto min-h-0">
    <!-- Header -->
    <div class="flex items-start justify-between mb-8 gap-4">
      <div>
        <h1 class="text-2xl font-bold text-slate-100">Dynamic Config</h1>
        <p class="text-slate-400 mt-1 text-sm">
          <span class="path-chip">{{ configStore.appConfig?.paths.dynamicDir ?? '…' }}</span>
        </p>
      </div>
      <button class="btn btn-primary flex-shrink-0" @click="showWizard = true">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4"/>
        </svg>
        New service
      </button>
    </div>

    <!-- Loading -->
    <div v-if="store.loading" class="flex items-center gap-3 text-slate-400">
      <svg class="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
      </svg>
      Loading files…
    </div>

    <!-- Error -->
    <div v-else-if="store.error" class="card border-red-800 bg-red-900/20 text-red-400">
      <p class="font-medium">Failed to load dynamic config</p>
      <p class="text-sm mt-1 opacity-80">{{ store.error }}</p>
    </div>

    <!-- Empty state -->
    <div v-else-if="!store.files.length" class="card flex flex-col items-center py-16 text-center max-w-sm">
      <div class="w-12 h-12 rounded-xl bg-sky-500/10 border border-sky-500/20 flex items-center justify-center mb-4">
        <svg class="w-6 h-6 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.6">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
        </svg>
      </div>
      <p class="text-slate-300 font-medium">No dynamic files yet</p>
      <p class="text-slate-500 text-sm mt-1">Create your first service to get started.</p>
    </div>

    <!-- File grid -->
    <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
      <ServiceCard
        v-for="file in store.files"
        :key="file.name"
        :file="file"
        @edit="router.push({ name: 'dynamic-editor', params: { file: file.name } })"
      />
    </div>

    <!-- Stats footer -->
    <div v-if="store.files.length" class="mt-6 flex gap-4 text-xs text-slate-500">
      <span>{{ activeCount }} active</span>
      <span v-if="inactiveCount">· {{ inactiveCount }} inactive (.bak)</span>
    </div>

    <!-- New service wizard -->
    <ServiceWizard
      v-if="showWizard"
      @close="showWizard = false"
      @created="onCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDynamicStore } from '@/stores/dynamic'
import { useConfigStore } from '@/stores/config'
import ServiceCard from '@/components/ServiceCard.vue'
import ServiceWizard from '@/components/ServiceWizard.vue'

const router = useRouter()
const store = useDynamicStore()
const configStore = useConfigStore()
const showWizard = ref(false)

const activeCount = computed(() => store.files.filter(f => f.active).length)
const inactiveCount = computed(() => store.files.filter(f => !f.active).length)

async function onCreated(name: string) {
  showWizard.value = false
  await store.fetchFiles()
  router.push({ name: 'dynamic-editor', params: { file: name } })
}

onMounted(async () => {
  await store.fetchFiles()
  // Ensure config is loaded for the wizard's cert resolver default.
  if (!configStore.appConfig) configStore.fetchConfig()
})
</script>
