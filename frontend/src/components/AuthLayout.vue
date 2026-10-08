<script setup lang="ts">
import BackgroundRipple from './ui/BackgroundRipple.vue'
</script>

<template>
  <div class="relative flex min-h-dvh items-center justify-center overflow-hidden bg-base-200 p-6">
    <BackgroundRipple />
    <main id="conteudo" class="pointer-events-none relative z-10 flex w-full justify-center">
      <div class="pointer-events-auto w-full max-w-sm">
        <div class="rise rounded-box border border-base-300 bg-base-100/85 p-7 shadow-2xl backdrop-blur-md sm:p-8" style="--d: 0ms">
          <slot />
        </div>
      </div>
    </main>
  </div>
</template>
