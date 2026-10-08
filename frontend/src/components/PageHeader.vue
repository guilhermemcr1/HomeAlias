<script setup lang="ts">
defineProps<{ title: string; description?: string }>()
</script>

<template>
  <header class="flex flex-wrap items-end justify-between gap-4 mb-6">
    <div class="min-w-0">
      <h1 class="page-title">{{ title }}</h1>
      <p v-if="description" class="mt-1.5 text-base-content/70 max-w-[65ch]">{{ description }}</p>
    </div>
    <div class="flex min-w-0 max-w-full flex-wrap gap-2"><slot name="actions" /></div>
  </header>
</template>
