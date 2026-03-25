<template>
  <div class="p-8">
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-slate-100">Certificates</h1>
        <p class="text-slate-400 mt-1 text-sm">ACME certificate status from acme.json</p>
      </div>
      <button class="btn-primary" @click="certStore.fetchCerts()">
        <svg class="w-4 h-4 mr-1.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h5M20 20v-5h-5M4 9a9 9 0 0115.45-3M20 15a9 9 0 01-15.45 3"/>
        </svg>
        Refresh
      </button>
    </div>

    <!-- ACME not available -->
    <div v-if="!certStore.available && !certStore.loading" class="card text-center py-12">
      <svg class="w-12 h-12 mx-auto text-slate-600 mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
      </svg>
      <p class="text-slate-400 font-medium">acme.json not found</p>
      <p class="text-slate-500 text-sm mt-1">No ACME certificates have been issued yet, or the path is not configured.</p>
    </div>

    <!-- Loading -->
    <div v-else-if="certStore.loading" class="card text-center py-12">
      <svg class="w-8 h-8 mx-auto animate-spin text-sky-500" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"/>
      </svg>
    </div>

    <!-- Cert list -->
    <template v-else-if="certStore.available">
      <!-- Warnings -->
      <div v-if="certStore.expired.length" class="mb-4 p-4 rounded-lg bg-red-900/30 border border-red-700/50 text-red-300 text-sm">
        <strong>{{ certStore.expired.length }} certificate(s) have expired</strong> and need to be renewed.
      </div>
      <div v-else-if="certStore.expiringSoon.length" class="mb-4 p-4 rounded-lg bg-yellow-900/30 border border-yellow-700/50 text-yellow-300 text-sm">
        <strong>{{ certStore.expiringSoon.length }} certificate(s) expire within 30 days.</strong>
      </div>

      <!-- Empty -->
      <div v-if="!certStore.certs.length" class="card text-center py-12">
        <p class="text-slate-400">No certificates found in acme.json.</p>
      </div>

      <!-- Table -->
      <div v-else class="card overflow-hidden p-0">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-slate-700 text-xs uppercase tracking-wider text-slate-500">
              <th class="text-left px-4 py-3">Domain</th>
              <th class="text-left px-4 py-3 hidden sm:table-cell">SANs</th>
              <th class="text-left px-4 py-3 hidden md:table-cell">Resolver</th>
              <th class="text-left px-4 py-3">Expiry</th>
              <th class="text-left px-4 py-3">Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="cert in certStore.certs" :key="cert.domain + cert.resolver"
              class="border-b border-slate-700/50 last:border-0 hover:bg-slate-700/20 transition-colors">
              <td class="px-4 py-3 font-mono text-slate-200 font-medium">{{ cert.domain }}</td>
              <td class="px-4 py-3 hidden sm:table-cell">
                <div class="flex flex-wrap gap-1">
                  <span v-for="san in cert.sans" :key="san" class="badge bg-slate-700 text-slate-400 border-slate-600 text-xs font-mono">
                    {{ san }}
                  </span>
                  <span v-if="!cert.sans?.length" class="text-slate-500 text-xs">—</span>
                </div>
              </td>
              <td class="px-4 py-3 hidden md:table-cell text-slate-400">{{ cert.resolver }}</td>
              <td class="px-4 py-3 text-slate-400 text-xs">
                {{ cert.expiry ? new Date(cert.expiry).toLocaleDateString() : '—' }}
              </td>
              <td class="px-4 py-3">
                <span v-if="!cert.expiry" class="badge bg-slate-700 text-slate-400 border-slate-600">unknown</span>
                <span v-else-if="cert.daysLeft < 0" class="badge badge-red">expired</span>
                <span v-else-if="cert.daysLeft <= 10" class="badge badge-red">{{ cert.daysLeft }}d left</span>
                <span v-else-if="cert.daysLeft <= 30" class="badge badge-yellow">{{ cert.daysLeft }}d left</span>
                <span v-else class="badge badge-green">{{ cert.daysLeft }}d left</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- mTLS section -->
      <div class="mt-8">
        <h2 class="text-sm font-semibold text-slate-300 uppercase tracking-wider mb-4">mTLS / TLS Options</h2>
        <div class="card">
          <p class="text-sm text-slate-400 mb-4">
            mTLS is configured via <span class="font-mono text-sky-400">tls.options</span> in a dynamic config file.
            Create or edit a dynamic file with the structure below, then reference the option name in your router.
          </p>
          <pre class="bg-slate-900 rounded-lg p-4 text-xs font-mono text-slate-300 overflow-x-auto">tls:
  options:
    myMTLS:
      minVersion: VersionTLS12
      clientAuth:
        caFiles:
          - /etc/traefik/certs/ca.crt
        clientAuthType: RequireAndVerifyClientCert</pre>
          <p class="text-xs text-slate-500 mt-3">
            Valid <code class="text-sky-400">clientAuthType</code> values:
            <code class="text-slate-300">NoClientCert</code>,
            <code class="text-slate-300">RequestClientCert</code>,
            <code class="text-slate-300">RequireAnyClientCert</code>,
            <code class="text-slate-300">VerifyClientCertIfGiven</code>,
            <code class="text-slate-300">RequireAndVerifyClientCert</code>
          </p>
          <p class="text-xs text-slate-500 mt-2">
            Then in your router: <code class="text-sky-400">tls.options: myMTLS@file</code>
          </p>
          <router-link to="/dynamic" class="inline-block mt-4 text-sm text-sky-400 hover:text-sky-300 transition-colors">
            → Manage dynamic config files
          </router-link>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useCertStore } from '@/stores/certs'

const certStore = useCertStore()

onMounted(() => certStore.fetchCerts())
</script>
