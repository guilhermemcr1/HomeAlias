<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import AppShell from './components/AppShell.vue'
import AuthLayout from './components/AuthLayout.vue'
import ToastHost from './components/ToastHost.vue'

const route = useRoute()
const isPublic = computed(() => route.meta.public === true)
</script>

<template>
  <a href="#conteudo" class="sr-only focus:not-sr-only focus:fixed focus:left-3 focus:top-3 focus:z-50 btn btn-primary btn-sm">Ir para o conteúdo</a>
  <AuthLayout v-if="isPublic">
    <RouterView v-slot="{ Component }">
      <Transition name="page" mode="out-in"><component :is="Component" :key="route.path" /></Transition>
    </RouterView>
  </AuthLayout>
  <AppShell v-else />
  <ToastHost />
</template>
