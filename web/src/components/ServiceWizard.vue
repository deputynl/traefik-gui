<template>
  <!-- Backdrop -->
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" @click.self="$emit('close')">
    <div class="w-full max-w-md bg-slate-800 border border-slate-700 rounded-2xl shadow-2xl">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-5 border-b border-slate-700">
        <h2 class="text-base font-semibold text-slate-100">New Service</h2>
        <button class="text-slate-500 hover:text-slate-300 transition-colors" @click="$emit('close')">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <!-- Form -->
      <form class="px-6 py-5 space-y-4" @submit.prevent="submit">
        <!-- Name -->
        <div>
          <label class="block text-xs font-medium text-slate-400 mb-1.5">Service name <span class="text-red-400">*</span></label>
          <input
            v-model="form.name"
            type="text"
            placeholder="myapp"
            class="input w-full"
            required
            pattern="[a-zA-Z0-9\-_]+"
            title="Letters, digits, hyphens and underscores only"
          />
          <p class="text-xs text-slate-500 mt-1">Will create <span class="font-mono text-slate-400">{{ form.name || 'name' }}.yml</span></p>
        </div>

        <!-- Hostname -->
        <div>
          <label class="block text-xs font-medium text-slate-400 mb-1.5">Hostname <span class="text-red-400">*</span></label>
          <input
            v-model="form.hostname"
            type="text"
            placeholder="myapp.example.com"
            class="input w-full"
            required
          />
        </div>

        <!-- Backend URL -->
        <div>
          <label class="block text-xs font-medium text-slate-400 mb-1.5">Backend URL <span class="text-red-400">*</span></label>
          <input
            v-model="form.backendUrl"
            type="url"
            placeholder="http://10.0.0.1:8080"
            class="input w-full"
            required
          />
        </div>

        <!-- Cert resolver -->
        <div>
          <label class="block text-xs font-medium text-slate-400 mb-1.5">Certificate resolver</label>
          <input
            v-model="form.certResolver"
            type="text"
            :placeholder="defaultCertResolver || 'cf-dns'"
            class="input w-full"
          />
        </div>

        <!-- Insecure backend -->
        <label class="flex items-center gap-3 cursor-pointer select-none">
          <div class="relative">
            <input v-model="form.insecureBackend" type="checkbox" class="sr-only peer" />
            <div class="w-9 h-5 bg-slate-700 peer-checked:bg-sky-500 rounded-full transition-colors"></div>
            <div class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform peer-checked:translate-x-4"></div>
          </div>
          <div>
            <span class="text-sm text-slate-300">Skip TLS verification</span>
            <p class="text-xs text-slate-500">For backends with self-signed certificates (Proxmox, etc.)</p>
          </div>
        </label>

        <!-- Error -->
        <p v-if="store.error" class="text-sm text-red-400 bg-red-900/20 border border-red-800 rounded-lg px-3 py-2">
          {{ store.error }}
        </p>

        <!-- Actions -->
        <div class="flex items-center justify-end gap-3 pt-2">
          <button type="button" class="btn btn-secondary" @click="$emit('close')">Cancel</button>
          <button type="submit" class="btn btn-primary" :disabled="store.saving">
            <svg v-if="store.saving" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
            </svg>
            {{ store.saving ? 'Creating…' : 'Create service' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { useDynamicStore } from '@/stores/dynamic'
import { useConfigStore } from '@/stores/config'

const emit = defineEmits<{ close: []; created: [name: string] }>()

const store = useDynamicStore()
const configStore = useConfigStore()

const defaultCertResolver = Object.keys(
  configStore.appConfig?.staticConfig?.certificatesResolvers ?? {}
)[0] ?? ''

const form = reactive({
  name: '',
  hostname: '',
  backendUrl: '',
  certResolver: defaultCertResolver,
  insecureBackend: false,
})

async function submit() {
  const created = await store.createService({
    name: form.name,
    hostname: form.hostname,
    backendUrl: form.backendUrl,
    certResolver: form.certResolver,
    insecureBackend: form.insecureBackend,
    entryPoints: ['websecure'],
  })
  if (created) {
    emit('created', created)
  }
}
</script>
