<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { History } from 'lucide-vue-next'
import { api, errorMessage } from '../api/client'
import LoadError from '../components/LoadError.vue'
import EmptyState from '../components/EmptyState.vue'
import PageHeader from '../components/PageHeader.vue'
import SkeletonRows from '../components/SkeletonRows.vue'
import { dateTime } from '../composables/format'
import { resultLabel, updateError } from '../composables/uiCopy'

type Ev = {
  id: number
  created_at: string
  result: string
  detected_ip?: string | null
  previous_ip?: string | null
  client_type: string
  changed: boolean
  error_reason?: string | null
}

const clientLabel: Record<string, string> = {
  shell: 'Script Linux',
  docker: 'Docker',
  windows: 'Windows',
  dyndns: 'Roteador (DDNS)',
  duckdns: 'cURL / DuckDNS',
  agent: 'Agente de atualização',
}
const events = ref<Ev[]>([])
const loading = ref(true)
const error = ref('')
const result = ref('')
const onlyChanged = ref(false)

let reloading = false
async function reload() {
  if (reloading) return
  reloading = true
  loading.value = true
  error.value = ''
  try {
    events.value = (await api<Ev[]>('/api/history')) ?? []
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
    reloading = false
  }
}
onMounted(reload)

const results = computed(() => [...new Set(events.value.map((e) => e.result))])
const rows = computed(() =>
  events.value.filter((e) => (!result.value || e.result === result.value) && (!onlyChanged.value || e.changed)),
)
const isOk = (r: string) => ['good', 'nochg', 'ok', 'success', 'updated', 'unchanged'].includes(r.toLowerCase())
</script>

<template>
  <div>
    <PageHeader title="Histórico" description="Veja as atualizações enviadas pelos seus dispositivos e o resultado de cada tentativa." />
    <LoadError v-if="error" :message="error" :busy="loading" @retry="reload" />
    <section v-if="!error" class="surface">
      <SkeletonRows v-if="loading" />
      <EmptyState v-else-if="!events.length" title="Ainda não há atualizações" description="Quando seu dispositivo enviar a primeira atualização, o resultado aparecerá aqui.">
        <template #icon><History :size="28" /></template>
      </EmptyState>
      <template v-else>
        <div class="flex flex-wrap items-center gap-4 border-b border-base-300 p-3">
          <label class="flex w-full min-w-0 flex-col gap-2 text-sm font-semibold sm:w-auto sm:flex-row sm:items-center">
            Resultado
            <select v-model="result" class="select select-bordered select-sm w-full min-w-0 sm:w-auto">
              <option value="">Todos</option>
              <option v-for="r in results" :key="r" :value="r">{{ resultLabel(r) }}</option>
            </select>
          </label>
          <label class="flex items-center gap-2 text-sm cursor-pointer">
            <input v-model="onlyChanged" type="checkbox" class="checkbox checkbox-primary checkbox-sm" />Apenas mudanças de IP
          </label>
          <span class="ml-auto text-sm text-base-content/65">{{ rows.length }} de {{ events.length }}</span>
        </div>
        <div class="overflow-x-auto">
          <table class="table">
            <thead>
              <tr><th>Quando</th><th>Resultado</th><th>IP detectado</th><th class="hidden md:table-cell">Cliente</th><th>IP alterado</th><th class="hidden lg:table-cell">Erro</th></tr>
            </thead>
            <tbody>
              <tr v-for="e in rows" :key="e.id" class="hover">
                <td class="whitespace-nowrap">{{ dateTime(e.created_at) }}</td>
                <td><span class="badge badge-outline h-auto whitespace-normal py-1" :class="isOk(e.result) ? 'badge-success' : 'badge-error'" :title="e.result">{{ resultLabel(e.result) }}</span></td>
                <td class="font-data">
                  {{ e.detected_ip || '—' }}
                  <span v-if="e.changed && e.previous_ip" class="block text-xs text-base-content/60">antes {{ e.previous_ip }}</span>
                </td>
                <td class="hidden md:table-cell">{{ clientLabel[e.client_type] ?? e.client_type }}</td>
                <td>{{ e.changed ? 'Sim' : 'Não' }}</td>
                <td class="hidden lg:table-cell">
                  <template v-if="e.error_reason">
                    <p class="text-error">{{ updateError(e.error_reason) }}</p>
                    <details class="mt-1 text-sm"><summary class="cursor-pointer">Detalhes técnicos</summary><p class="mt-1 break-words">{{ e.error_reason }}</p></details>
                  </template>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <EmptyState v-if="!rows.length" title="Nada com esses filtros" />
      </template>
    </section>
  </div>
</template>
