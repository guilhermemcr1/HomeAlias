<script setup lang="ts">
withDefaults(defineProps<{ size?: number; inverse?: boolean }>(), { size: 28, inverse: false })
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 32 32" fill="none" aria-hidden="true">
    <rect width="32" height="32" rx="9" :class="inverse ? 'fill-primary-content' : 'fill-primary'" />
    <g stroke="currentColor" :class="inverse ? 'text-primary' : 'text-primary-content'" stroke-width="2.4" stroke-linejoin="round" stroke-linecap="round">
      <path d="M8 22V11l8-5 8 5v11" />
      <path d="M12.5 22v-6h7v6" />
    </g>
  </svg>
</template>
