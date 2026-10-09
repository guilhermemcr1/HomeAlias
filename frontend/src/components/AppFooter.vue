<script setup lang="ts">
import { Github, UserRound } from 'lucide-vue-next'

const year = new Date().getFullYear()
const links = [
  { icon: Github, href: 'https://github.com/guilhermemcr1/HomeAlias', description: 'Código no GitHub' },
  { icon: UserRound, href: 'https://github.com/guilhermemcr1', description: 'GitHub de guilhermemcr1' },
]
</script>

<template>
  <footer class="footer shrink-0 border-t border-base-300 bg-base-100 text-base-content">
    <div class="mx-auto flex w-full max-w-7xl flex-wrap items-center justify-between gap-x-3 px-4 py-1 sm:px-8">
      <aside class="min-w-0 py-1">
        <p class="flex flex-wrap items-center gap-x-2 text-[12px] leading-5">
          <span class="font-semibold">HomeAlias</span>
          <span class="whitespace-nowrap text-base-content/70">© {{ year }}</span>
        </p>
      </aside>
      <nav class="ml-auto flex items-center" aria-label="Links do projeto">
        <div
          v-for="link in links"
          :key="link.href"
          class="tooltip tooltip-top footer-tooltip"
          :data-tip="link.description"
        >
          <a
            :href="link.href"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-ghost btn-sm btn-square !min-h-[44px] !min-w-[44px] text-base-content/70 hover:text-base-content"
            :aria-label="`${link.description} (abre em nova aba)`"
          >
            <component :is="link.icon" :size="18" aria-hidden="true" />
          </a>
        </div>
      </nav>
    </div>
  </footer>
</template>

<style scoped>
/* Keep right-edge tooltips inside small viewports. */
.footer-tooltip::before {
  left: auto;
  right: 0;
  transform: none;
  max-width: min(18rem, calc(100vw - 112px));
}
</style>
