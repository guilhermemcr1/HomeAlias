<script setup lang="ts">
import BackgroundRipple from './ui/BackgroundRipple.vue'
import AppFooter from './AppFooter.vue'
</script>

<template>
  <div class="relative flex min-h-dvh flex-col overflow-hidden bg-base-200">
    <BackgroundRipple />
    <main id="conteudo" class="pointer-events-none relative z-10 flex w-full flex-1 items-center justify-center p-6">
      <div class="pointer-events-auto w-full max-w-sm">
        <div class="rise rounded-box border border-base-300 bg-base-100/85 p-7 shadow-2xl backdrop-blur-md sm:p-8" style="--d: 0ms">
          <slot />
        </div>
      </div>
    </main>
    <AppFooter class="relative z-10" />
  </div>
</template>
