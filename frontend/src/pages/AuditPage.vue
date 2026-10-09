<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ScrollText } from 'lucide-vue-next'
import { api, errorMessage } from '../api/client'
import LoadError from '../components/LoadError.vue'
import EmptyState from '../components/EmptyState.vue'
import PageHeader from '../components/PageHeader.vue'
import SkeletonRows from '../components/SkeletonRows.vue'
import { dateTime } from '../composables/format'
import { actionLabel, resourceLabel, auditSummary } from '../composables/uiCopy'

type Row = { id: number; created_at: string; action: string; resource_type: string; resource_id: string; ip: string; summary: string }

const rows = ref<Row[]>([])
const loading = ref(true)
const error = ref('')

let reloading = false
async function reload() {
  if (reloading) return
  reloading = true
  loading.value = true
  error.value = ''
  try {
    rows.value = (await api<Row[]>('/api/audit')) ?? []
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
    reloading = false
  }
}
onMounted(reload)
</script>

<template>
  <div>
    <PageHeader title="Auditoria" description="Quem fez o quê no painel, com data e IP." />
    <LoadError v-if="error" :message="error" :busy="loading" @retry="reload" />
    <section v-if="!error" class="surface">
      <SkeletonRows v-if="loading" />
      <EmptyState v-else-if="!rows.length" title="Sem eventos registrados">
        <template #icon><ScrollText :size="28" /></template>
      </EmptyState>
      <div v-else class="overflow-x-auto">
        <table class="table">
          <thead><tr><th>Quando</th><th>Ação</th><th class="hidden md:table-cell">Recurso</th><th class="hidden sm:table-cell">IP</th><th>Resumo</th></tr></thead>
          <tbody>
            <tr v-for="r in rows" :key="r.id" class="hover">
              <td class="whitespace-nowrap">{{ dateTime(r.created_at) }}</td>
              <td :title="r.action">{{ actionLabel(r.action) }}</td>
              <td class="hidden md:table-cell" :title="r.resource_type">{{ resourceLabel(r.resource_type) }}</td>
              <td class="hidden sm:table-cell font-data">{{ r.ip }}</td>
              <td>{{ auditSummary(r.summary) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>
