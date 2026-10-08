<script setup lang="ts">
defineProps<{ title: string; description?: string }>()
</script>

<template>
  <div class="flex flex-col items-start gap-3 px-6 py-10">
    <div v-if="$slots.icon" class="text-primary"><slot name="icon" /></div>
    <h2 class="font-display text-xl font-bold tracking-tight">{{ title }}</h2>
    <p v-if="description" class="text-base-content/70 max-w-[60ch]">{{ description }}</p>
    <div class="flex flex-wrap gap-2 mt-1"><slot /></div>
  </div>
</template>
