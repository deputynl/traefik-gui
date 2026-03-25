<template>
  <div class="p-8">
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-slate-100">Certificates</h1>
      <p class="text-slate-400 mt-1 text-sm">
        ACME certificate status, resolver configuration, and mTLS options
      </p>
    </div>
    <PlaceholderView
      title="Certificate Management"
      description="View certificate expiry from acme.json, configure Let's Encrypt / ZeroSSL resolvers (HTTP-01, DNS-01, TLS-ALPN-01), and manage mTLS client CA stores. Coming in Phase 5."
      icon="cert"
    />
  </div>
</template>

<script setup lang="ts">
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
