<script setup lang="ts">
import { Github } from 'lucide-vue-next'
import BrandMark from './BrandMark.vue'

const year = new Date().getFullYear()
const links = [
  { label: 'Projeto', href: 'https://github.com/guilhermemcr1/HomeAlias', description: 'Repositório HomeAlias no GitHub' },
  { label: 'Autor', href: 'https://github.com/guilhermemcr1', description: 'Perfil de guilhermemcr1 no GitHub' },
]
</script>

<template>
  <footer class="footer shrink-0 border-t border-base-300 bg-base-100 text-base-content">
    <div class="mx-auto flex w-full max-w-7xl flex-col gap-2 px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-8">
      <aside class="flex min-w-0 items-center gap-3">
        <BrandMark :size="24" class="shrink-0" />
        <p class="text-sm">© {{ year }} HomeAlias</p>
      </aside>
      <nav class="flex flex-wrap items-center gap-2" aria-label="HomeAlias no GitHub">
        <a
          v-for="link in links"
          :key="link.href"
          :href="link.href"
          target="_blank"
          rel="noopener noreferrer"
          class="btn btn-ghost btn-sm !min-h-[44px] gap-2 text-sm"
          :aria-label="`${link.description} (abre em nova aba)`"
          :title="`${link.description} (abre em nova aba)`"
        >
          <Github :size="20" class="shrink-0" aria-hidden="true" />
          {{ link.label }}
        </a>
      </nav>
    </div>
  </footer>
</template>
