<script setup lang="ts">
import BrandMark from './BrandMark.vue'
</script>

<template>
  <div class="min-h-screen grid lg:grid-cols-[minmax(0,1.1fr)_minmax(0,1fr)] bg-base-200">
    <aside class="hidden lg:flex flex-col justify-between bg-primary text-primary-content p-12 relative overflow-hidden">
      <div class="flex items-center gap-3">
        <BrandMark :size="36" inverse />
        <span class="font-display text-xl font-bold tracking-tight">HomeAlias</span>
      </div>
      <div class="relative z-10">
        <p class="font-display font-extrabold text-6xl leading-[0.95] tracking-tight max-w-[12ch]">
          Seu IP muda. Seu endereço não.
        </p>
        <p class="mt-6 text-lg max-w-[40ch] opacity-90">
          DDNS self-hosted sobre a sua Cloudflare. Cada atualização registrada, cada queda avisada.
        </p>
      </div>
      <div class="font-data text-sm opacity-90 space-y-1" aria-hidden="true">
        <p>casa.seudominio.com <span class="opacity-60">→</span> 203.0.113.42</p>
        <p class="opacity-70">exemplo ilustrativo</p>
      </div>
      <svg class="absolute -right-24 -bottom-24 w-[34rem] opacity-[0.12]" viewBox="0 0 200 200" fill="none" aria-hidden="true">
        <circle cx="100" cy="100" r="95" stroke="currentColor" stroke-width="2" />
        <circle cx="100" cy="100" r="65" stroke="currentColor" stroke-width="2" />
        <circle cx="100" cy="100" r="35" stroke="currentColor" stroke-width="2" />
        <circle cx="100" cy="100" r="8" fill="currentColor" />
      </svg>
    </aside>
    <main id="conteudo" class="flex items-center justify-center p-6 sm:p-12">
      <div class="w-full max-w-sm enter">
        <div class="flex items-center gap-2.5 mb-10 lg:hidden">
          <BrandMark :size="32" />
          <span class="font-display text-xl font-bold tracking-tight">HomeAlias</span>
        </div>
        <slot />
      </div>
    </main>
  </div>
</template>
