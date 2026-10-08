<script setup lang="ts">
withDefaults(defineProps<{ rows?: number }>(), { rows: 4 })
</script>

<template>
  <div class="flex flex-col gap-3 p-4" aria-busy="true" aria-label="Carregando">
    <div v-for="n in rows" :key="n" class="skeleton h-11 w-full" />
  </div>
</template>
