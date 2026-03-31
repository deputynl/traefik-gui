<template>
  <div class="p-8 flex-1 overflow-y-auto min-h-0">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-slate-100">Docker Labels</h1>
        <p class="text-slate-400 mt-1 text-sm">Traefik labels on running containers</p>
      </div>
      <button class="btn-primary" @click="dockerStore.fetchContainers()">
        <svg class="w-4 h-4 mr-1.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h5M20 20v-5h-5M4 9a9 9 0 0115.45-3M20 15a9 9 0 01-15.45 3"/>
        </svg>
        Refresh
      </button>
    </div>

    <!-- Socket not available -->
    <div v-if="!dockerStore.available && !dockerStore.loading" class="card text-center py-12">
      <svg class="w-12 h-12 mx-auto text-slate-600 mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 003 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z"/>
      </svg>
      <p class="text-slate-400 font-medium">Docker socket not available</p>
      <p class="text-slate-500 text-sm mt-1 max-w-md mx-auto">
        Mount the Docker socket into the container to enable this feature.
      </p>
      <pre class="mt-4 text-left inline-block bg-slate-900 rounded-lg px-4 py-3 text-xs font-mono text-slate-300">volumes:
  - /var/run/docker.sock:/var/run/docker.sock:ro</pre>
    </div>

    <!-- Loading -->
    <div v-else-if="dockerStore.loading" class="card text-center py-12">
      <svg class="w-8 h-8 mx-auto animate-spin text-sky-500" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
      </svg>
    </div>

    <!-- Container list -->
    <template v-else-if="dockerStore.available">
      <div class="flex gap-3 mb-4 text-sm text-slate-400">
        <label class="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" v-model="showAll" class="rounded border-slate-600 bg-slate-700 text-sky-500 focus:ring-sky-500" />
          Show all containers (including non-Traefik)
        </label>
      </div>

      <div v-if="!visibleContainers.length" class="card text-center py-12">
        <p class="text-slate-400">No containers with Traefik labels found.</p>
      </div>

      <div v-else class="space-y-4">
        <div v-for="c in visibleContainers" :key="c.id"
          class="card">
          <!-- Header -->
          <div class="flex items-start justify-between gap-4 mb-3">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <span class="w-2 h-2 rounded-full flex-shrink-0"
                  :class="c.state === 'running' ? 'bg-emerald-400' : 'bg-slate-500'" />
                <span class="font-semibold text-slate-100">{{ c.name }}</span>
                <span v-if="c.enabled" class="badge badge-green text-xs">traefik enabled</span>
              </div>
              <div class="text-xs text-slate-500 mt-0.5 font-mono">{{ c.image }} · {{ c.id }}</div>
            </div>
            <button v-if="Object.keys(c.traefikLabels).length"
              class="btn-secondary text-xs flex-shrink-0"
              @click="copySnippet(c)"
              :title="copied === c.id ? 'Copied!' : 'Copy as docker-compose labels'">
              {{ copied === c.id ? '✓ Copied' : 'Copy labels' }}
            </button>
          </div>

          <!-- Labels -->
          <div v-if="Object.keys(c.traefikLabels).length" class="space-y-1">
            <div v-for="(val, key) in c.traefikLabels" :key="key"
              class="flex gap-2 text-xs font-mono py-1 border-b border-slate-700/40 last:border-0">
              <span class="text-sky-400 flex-shrink-0">{{ key }}</span>
              <span class="text-slate-300 truncate">{{ val }}</span>
            </div>
          </div>
          <div v-else class="text-xs text-slate-500 italic">No Traefik labels</div>
        </div>
      </div>

      <!-- Snippet generator -->
      <div class="mt-8">
        <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wider mb-4">Label Snippet Generator</h2>
        <div class="card space-y-4">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs text-slate-400 mb-1">Service name</label>
              <input v-model="gen.name" class="input w-full" placeholder="myapp" />
            </div>
            <div>
              <label class="block text-xs text-slate-400 mb-1">Hostname</label>
              <input v-model="gen.host" class="input w-full" placeholder="app.example.com" />
            </div>
            <div>
              <label class="block text-xs text-slate-400 mb-1">Internal port</label>
              <input v-model="gen.port" class="input w-full" placeholder="8080" />
            </div>
            <div>
              <label class="block text-xs text-slate-400 mb-1">Cert resolver (optional)</label>
              <input v-model="gen.certResolver" class="input w-full" placeholder="cf-dns" />
            </div>
          </div>
          <div v-if="gen.name && gen.host">
            <label class="block text-xs text-slate-400 mb-1">Generated labels</label>
            <pre class="bg-slate-900 rounded-lg p-4 text-xs font-mono text-slate-300 overflow-x-auto">{{ generatedSnippet }}</pre>
            <button class="btn-secondary text-xs mt-2" @click="copyGenerated">
              {{ copiedGen ? '✓ Copied' : 'Copy' }}
            </button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useDockerStore, type DockerContainer } from '@/stores/docker'

const dockerStore = useDockerStore()
const showAll = ref(false)
const copied = ref('')
const copiedGen = ref(false)

const gen = ref({ name: '', host: '', port: '', certResolver: '' })

const visibleContainers = computed(() =>
  showAll.value
    ? dockerStore.containers
    : dockerStore.containers.filter(c => Object.keys(c.traefikLabels).length > 0)
)

const generatedSnippet = computed(() => {
  const { name, host, port, certResolver } = gen.value
  if (!name || !host) return ''
  const lines = [
    `labels:`,
    `  - "traefik.enable=true"`,
    `  - "traefik.http.routers.${name}.rule=Host(\`${host}\`)"`,
    `  - "traefik.http.routers.${name}.entrypoints=websecure"`,
  ]
  if (certResolver) {
    lines.push(`  - "traefik.http.routers.${name}.tls.certresolver=${certResolver}"`)
  } else {
    lines.push(`  - "traefik.http.routers.${name}.tls=true"`)
  }
  if (port) {
    lines.push(`  - "traefik.http.services.${name}.loadbalancer.server.port=${port}"`)
  }
  return lines.join('\n')
})

function copySnippet(c: DockerContainer) {
  const lines = ['labels:']
  for (const [k, v] of Object.entries(c.traefikLabels)) {
    lines.push(`  - "${k}=${v}"`)
  }
  navigator.clipboard.writeText(lines.join('\n'))
  copied.value = c.id
  setTimeout(() => { copied.value = '' }, 2000)
}

function copyGenerated() {
  navigator.clipboard.writeText(generatedSnippet.value)
  copiedGen.value = true
  setTimeout(() => { copiedGen.value = false }, 2000)
}

onMounted(() => dockerStore.fetchContainers())
</script>
