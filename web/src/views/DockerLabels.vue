<template>
  <div class="p-8">
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-slate-100">Docker Labels</h1>
      <p class="text-slate-400 mt-1 text-sm">
        Inspect Traefik labels on running containers and generate label snippets
      </p>
    </div>
    <PlaceholderView
      title="Docker Label Inspector"
      description="Browse running containers and their Traefik labels. Generate and copy label snippets for common patterns (expose a service, add middleware, configure TLS). Coming in Phase 6."
      icon="docker"
    />
  </div>
</template>

<script setup lang="ts">
import PlaceholderView from '@/components/PlaceholderView.vue'
</script>
