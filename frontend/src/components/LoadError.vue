<script setup lang="ts">
defineProps<{ message: string; busy?: boolean }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <div class="mb-4 flex flex-wrap items-center gap-3 rounded-btn border border-error/50 bg-error/10 p-4" role="alert">
    <p class="min-w-0 flex-1 basis-48 text-sm">{{ message }}</p>
    <button type="button" class="btn btn-sm btn-ghost" :disabled="busy" @click="$emit('retry')">{{ busy ? 'Carregando…' : 'Tentar novamente' }}</button>
  </div>
</template>
