<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import AppShell from './components/AppShell.vue'
import AuthLayout from './components/AuthLayout.vue'
import ToastHost from './components/ToastHost.vue'
import { useAuth } from './composables/useAuth'

const route = useRoute()
const { ready, user } = useAuth()
const isPublic = computed(() => route.meta.public === true)
const isPending = computed(() =>
  !ready.value || route.matched.length === 0 || (!isPublic.value && !user.value),
)
</script>

<template>
  <a href="#conteudo" class="sr-only focus:not-sr-only focus:fixed focus:left-3 focus:top-3 focus:z-50 btn btn-primary btn-sm">Ir para o conteúdo</a>
  <main v-if="isPending" id="conteudo" class="grid min-h-dvh place-items-center bg-base-100 px-6">
    <p role="status" class="text-sm text-base-content">Carregando…</p>
  </main>
  <AuthLayout v-else-if="isPublic">
    <RouterView v-slot="{ Component }">
      <Transition name="page" mode="out-in"><component :is="Component" :key="route.path" /></Transition>
    </RouterView>
  </AuthLayout>
  <AppShell v-else />
  <ToastHost />
</template>
