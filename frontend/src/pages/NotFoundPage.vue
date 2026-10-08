<script setup lang="ts">
import { RouterLink } from 'vue-router'
import EmptyState from '../components/EmptyState.vue'
</script>

<template>
  <EmptyState title="Página não encontrada" description="O endereço não existe ou foi movido.">
    <RouterLink to="/" class="btn btn-primary">Voltar ao painel</RouterLink>
  </EmptyState>
</template>
