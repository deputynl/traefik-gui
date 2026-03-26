<template>
  <div class="p-8 max-w-3xl">
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-slate-100">mTLS</h1>
      <p class="text-slate-400 mt-1 text-sm">Mutual TLS — manage your CA and client certificates.</p>
    </div>

    <!-- Step 1: CA -->
    <div class="card mb-5">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wider mb-1">
            Certificate Authority
          </h2>
          <p v-if="status?.caExists" class="text-xs text-slate-500">
            Expires <span class="text-slate-400">{{ status.caExpires }}</span>
          </p>
          <p v-else class="text-xs text-slate-500">No CA generated yet.</p>
        </div>
        <div class="flex gap-2 flex-shrink-0">
          <a v-if="status?.caExists" href="/api/mtls/ca/download"
            class="btn btn-secondary text-xs">
            Download CA cert
          </a>
          <button class="btn text-xs"
            :class="status?.caExists ? 'btn-secondary text-orange-400 border-orange-800 hover:border-orange-600' : 'btn-primary'"
            :disabled="caLoading" @click="generateCA">
            {{ caLoading ? 'Generating…' : status?.caExists ? 'Regenerate CA' : 'Generate CA' }}
          </button>
        </div>
      </div>
      <div v-if="status?.caExists && status.clients.length === 0 && !status.applied" class="mt-3 text-xs text-yellow-400/80">
        CA generated. Next: apply the TLS option to Traefik, then issue client certificates.
      </div>
    </div>

    <!-- Step 2: Apply TLS option -->
    <div class="card mb-5">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wider mb-1">
            Traefik TLS Option
          </h2>
          <p class="text-xs text-slate-500">
            Writes <span class="font-mono">mtls.yml</span> to your dynamic config directory,
            defining the <span class="font-mono">mtls</span> TLS option that entry points reference.
          </p>
        </div>
        <div class="flex items-center gap-3 flex-shrink-0">
          <span class="flex items-center gap-1.5 text-xs"
            :class="status?.applied ? 'text-emerald-400' : 'text-slate-500'">
            <span class="w-2 h-2 rounded-full flex-shrink-0"
              :class="status?.applied ? 'bg-emerald-400' : 'bg-slate-600'" />
            {{ status?.applied ? 'Applied' : 'Not applied' }}
          </span>
          <button class="btn btn-primary text-xs" :disabled="!status?.caExists || applyLoading" @click="applyTLS">
            {{ applyLoading ? 'Applying…' : status?.applied ? 'Re-apply' : 'Apply to Traefik' }}
          </button>
        </div>
      </div>
      <p v-if="!status?.caExists" class="mt-2 text-xs text-slate-600">
        Generate a CA first.
      </p>
    </div>

    <!-- Step 3: Client certificates -->
    <div class="card">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wider">
          Client Certificates
        </h2>
        <button class="btn btn-primary text-xs" :disabled="!status?.caExists" @click="showIssue = true">
          + Issue certificate
        </button>
      </div>

      <p v-if="!status?.caExists" class="text-xs text-slate-600">Generate a CA first.</p>
      <p v-else-if="!status.clients.length" class="text-xs text-slate-500">
        No client certificates issued yet.
      </p>
      <table v-else class="w-full text-xs">
        <thead>
          <tr class="text-left text-slate-500 border-b border-slate-700">
            <th class="pb-2 font-medium">Name</th>
            <th class="pb-2 font-medium">Issued</th>
            <th class="pb-2 font-medium">Expires</th>
            <th class="pb-2"></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="c in status.clients" :key="c.id"
            class="border-b border-slate-800 last:border-0">
            <td class="py-2.5 pr-4 text-slate-300 font-medium">{{ c.name }}</td>
            <td class="py-2.5 pr-4 text-slate-500">{{ formatDate(c.issued) }}</td>
            <td class="py-2.5 pr-4">
              <span :class="expiryClass(c.expires)">{{ formatDate(c.expires) }}</span>
            </td>
            <td class="py-2.5 text-right">
              <div class="flex items-center justify-end gap-2">
                <a :href="`/api/mtls/clients/${c.id}/download`"
                  class="text-sky-400 hover:text-sky-300 transition-colors">Download</a>
                <button class="text-slate-500 hover:text-red-400 transition-colors"
                  @click="revoke(c.id, c.name)">Revoke</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Issue modal -->
    <div v-if="showIssue" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4"
      @click.self="showIssue = false">
      <div class="card w-full max-w-sm">
        <h3 class="text-sm font-semibold text-slate-200 mb-4">Issue client certificate</h3>
        <label class="field-label">Name</label>
        <input v-model="newCertName" type="text" class="input w-full mb-4"
          placeholder="e.g. Michel's laptop" @keydown.enter="issueClient" />
        <div class="flex gap-2 justify-end">
          <button class="btn btn-secondary text-xs" @click="showIssue = false">Cancel</button>
          <button class="btn btn-primary text-xs" :disabled="!newCertName.trim() || issueLoading"
            @click="issueClient">
            {{ issueLoading ? 'Generating…' : 'Generate & Download' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Feedback -->
    <div v-if="msg" class="fixed bottom-6 right-6 px-4 py-2.5 rounded-lg border text-sm shadow-xl z-50"
      :class="msg.ok
        ? 'bg-emerald-900/90 border-emerald-700 text-emerald-300'
        : 'bg-red-900/90 border-red-700 text-red-300'">
      {{ msg.text }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

interface ClientEntry {
  id: string
  name: string
  issued: string
  expires: string
}

interface MTLSStatus {
  caExists: boolean
  caExpires: string | null
  clients: ClientEntry[]
  applied: boolean
}

const status = ref<MTLSStatus | null>(null)
const caLoading = ref(false)
const applyLoading = ref(false)
const issueLoading = ref(false)
const showIssue = ref(false)
const newCertName = ref('')
const msg = ref<{ ok: boolean; text: string } | null>(null)

function showMsg(ok: boolean, text: string) {
  msg.value = { ok, text }
  setTimeout(() => { msg.value = null }, 4000)
}

async function fetchStatus() {
  const res = await fetch('/api/mtls')
  if (res.ok) status.value = await res.json()
}

async function generateCA() {
  if (status.value?.caExists &&
    !confirm('Regenerating the CA will invalidate all existing client certificates. Continue?')) return
  caLoading.value = true
  try {
    const res = await fetch('/api/mtls/ca', { method: 'POST' })
    if (!res.ok) throw new Error((await res.json()).error)
    await fetchStatus()
    showMsg(true, 'CA generated successfully.')
  } catch (e) {
    showMsg(false, String(e))
  } finally {
    caLoading.value = false
  }
}

async function applyTLS() {
  applyLoading.value = true
  try {
    const res = await fetch('/api/mtls/apply', { method: 'POST' })
    if (!res.ok) throw new Error((await res.json()).error)
    await fetchStatus()
    showMsg(true, 'mtls.yml written to dynamic config directory.')
  } catch (e) {
    showMsg(false, String(e))
  } finally {
    applyLoading.value = false
  }
}

async function issueClient() {
  if (!newCertName.value.trim()) return
  issueLoading.value = true
  try {
    const res = await fetch('/api/mtls/clients', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newCertName.value.trim() }),
    })
    if (!res.ok) throw new Error((await res.json()).error)
    const entry: ClientEntry = await res.json()
    // Trigger download immediately.
    const a = document.createElement('a')
    a.href = `/api/mtls/clients/${entry.id}/download`
    a.click()
    newCertName.value = ''
    showIssue.value = false
    await fetchStatus()
    showMsg(true, 'Certificate issued — download started.')
  } catch (e) {
    showMsg(false, String(e))
  } finally {
    issueLoading.value = false
  }
}

async function revoke(id: string, name: string) {
  if (!confirm(`Revoke certificate for "${name}"? The client will lose access immediately.`)) return
  try {
    const res = await fetch(`/api/mtls/clients/${id}`, { method: 'DELETE' })
    if (!res.ok) throw new Error((await res.json()).error)
    await fetchStatus()
    showMsg(true, `Certificate for "${name}" revoked.`)
  } catch (e) {
    showMsg(false, String(e))
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString([], { year: 'numeric', month: 'short', day: 'numeric' })
}

function expiryClass(iso: string) {
  const days = (new Date(iso).getTime() - Date.now()) / 86400000
  if (days < 0) return 'text-red-400'
  if (days < 30) return 'text-yellow-400'
  return 'text-slate-400'
}

onMounted(fetchStatus)
</script>
