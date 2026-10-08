<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { ArrowRight, Check, Globe, RefreshCw } from 'lucide-vue-next'
import { api, errorMessage } from '../api/client'
import LoadError from '../components/LoadError.vue'
import EmptyState from '../components/EmptyState.vue'
import PageHeader from '../components/PageHeader.vue'
import SkeletonRows from '../components/SkeletonRows.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { relativeTime } from '../composables/format'

type Host = {
  id: string
  fqdn: string
  status: string
  last_ipv4?: string | null
  last_ipv6?: string | null
  last_seen_v4?: string | null
  last_seen_v6?: string | null
  proxied: boolean
}

const hosts = ref<Host[]>([])
const connections = ref<unknown[]>([])
const loading = ref(true)
const error = ref('')
const filter = ref<'all' | 'online' | 'warning' | 'offline'>('all')
const refreshing = ref(false)
const loaded = ref(false)
let timer: number | undefined

async function load() {
  if (refreshing.value) return
  refreshing.value = true
  try {
    const [h, c] = await Promise.all([api<Host[]>('/api/hosts'), api<unknown[]>('/api/connections')])
    hosts.value = h ?? []
    connections.value = c ?? []
    error.value = ''
    loaded.value = true
  } catch (e) {
    error.value = errorMessage(e, 'Não foi possível carregar os hosts.')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

onMounted(() => {
  load()
  timer = window.setInterval(load, 60_000)
})
onUnmounted(() => window.clearInterval(timer))

const order: Record<string, number> = { offline: 0, warning: 1, never_seen: 2, online: 3 }
const sorted = computed(() =>
  [...hosts.value].sort((a, b) => (order[a.status] ?? 9) - (order[b.status] ?? 9) || a.fqdn.localeCompare(b.fqdn)),
)
const visible = computed(() => (filter.value === 'all' ? sorted.value : sorted.value.filter((h) => h.status === filter.value)))
const counts = computed(() => ({
  all: hosts.value.length,
  online: hosts.value.filter((h) => h.status === 'online').length,
  warning: hosts.value.filter((h) => h.status === 'warning').length,
  offline: hosts.value.filter((h) => h.status === 'offline').length,
}))
const filters = [
  { key: 'all', label: 'Todos' },
  { key: 'online', label: 'Online' },
  { key: 'warning', label: 'Atenção' },
  { key: 'offline', label: 'Offline' },
] as const

function lastSeen(h: Host): string | null {
  const a = h.last_seen_v4 ? new Date(h.last_seen_v4).getTime() : 0
  const b = h.last_seen_v6 ? new Date(h.last_seen_v6).getTime() : 0
  const t = Math.max(a, b)
  return t ? new Date(t).toISOString() : null
}

const hasConnection = computed(() => connections.value.length > 0)
const hasHost = computed(() => hosts.value.length > 0)
const steps = computed(() => [
  { n: 1, title: 'Conecte sua Cloudflare', text: 'Cole um token de API com permissão Zone DNS Edit. Ele é validado e guardado criptografado.', to: '/connections', cta: 'Adicionar conexão', done: hasConnection.value },
  { n: 2, title: 'Crie um host', text: 'Escolha a zona e o subdomínio que vai acompanhar o seu IP.', to: '/hosts', cta: 'Criar host', done: hasHost.value },
  { n: 3, title: 'Gere um token e configure o cliente', text: 'Roteador, Docker ou Windows chamam o HomeAlias para manter o DNS atualizado.', to: '/tokens', cta: 'Gerar token', done: false },
])
const currentStep = computed(() => (hasHost.value ? 3 : hasConnection.value ? 2 : 1))
</script>

<template>
  <div>
    <PageHeader title="Painel" description="O estado dos seus hosts, os mais urgentes primeiro.">
      <template #actions>
        <button type="button" class="btn btn-ghost gap-2" :disabled="refreshing" @click="load">
          <RefreshCw :size="16" aria-hidden="true" />Atualizar
        </button>
      </template>
    </PageHeader>

    <LoadError v-if="error" :message="error" :busy="refreshing" @retry="load" />

    <div v-if="loading" class="surface"><SkeletonRows /></div>

    <section v-else-if="loaded && !hasHost" class="surface p-6 sm:p-8">
      <h2 class="font-display text-2xl font-bold tracking-tight">Vamos colocar o primeiro host no ar</h2>
      <p class="mt-1.5 text-base-content/70">Três passos. Você pode voltar aqui a qualquer momento.</p>
      <ol class="mt-6 grid gap-3">
        <li
          v-for="s in steps"
          :key="s.n"
          class="flex flex-wrap items-center gap-4 rounded-btn border p-4"
          :class="s.n === currentStep ? 'border-primary bg-primary/5' : 'border-base-300'"
        >
          <span
            class="grid size-9 shrink-0 place-items-center rounded-full font-bold"
            :class="s.done ? 'bg-success text-success-content' : s.n === currentStep ? 'bg-primary text-primary-content' : 'bg-base-300'"
          >
            <Check v-if="s.done" :size="18" aria-label="Concluído" />
            <span v-else>{{ s.n }}</span>
          </span>
          <div class="min-w-0 flex-1 basis-64">
            <h3 class="font-semibold">{{ s.title }}</h3>
            <p class="text-sm text-base-content/70">{{ s.text }}</p>
          </div>
          <RouterLink v-if="s.n === currentStep" :to="s.to" class="btn btn-primary gap-2">
            {{ s.cta }}<ArrowRight :size="16" aria-hidden="true" />
          </RouterLink>
        </li>
      </ol>
    </section>

    <section v-else-if="loaded" class="surface">
      <div class="flex flex-wrap gap-2 border-b border-base-300 p-3" role="group" aria-label="Filtrar por estado">
        <button
          v-for="f in filters"
          :key="f.key"
          type="button"
          class="btn btn-sm"
          :class="filter === f.key ? 'btn-primary' : 'btn-ghost'"
          :aria-pressed="filter === f.key"
          @click="filter = f.key"
        >
          {{ f.label }}<span class="font-data opacity-80">{{ counts[f.key] }}</span>
        </button>
      </div>
      <EmptyState v-if="!visible.length" title="Nenhum host neste estado" description="Troque o filtro para ver os demais.">
        <template #icon><Globe :size="28" /></template>
      </EmptyState>
      <div v-else class="overflow-x-auto">
        <table class="table">
          <thead>
            <tr>
              <th>Host</th>
              <th>Estado</th>
              <th class="hidden sm:table-cell">IPv4</th>
              <th class="hidden lg:table-cell">IPv6</th>
              <th>Último contato</th>
              <th class="hidden md:table-cell">Cloudflare</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="h in visible" :key="h.id" class="hover">
              <td class="font-data font-semibold">{{ h.fqdn }}</td>
              <td><StatusBadge :status="h.status" /></td>
              <td class="hidden sm:table-cell font-data">{{ h.last_ipv4 || '—' }}</td>
              <td class="hidden lg:table-cell font-data max-w-48 truncate" :title="h.last_ipv6 ?? ''">{{ h.last_ipv6 || '—' }}</td>
              <td class="whitespace-nowrap">{{ relativeTime(lastSeen(h)) }}</td>
              <td class="hidden md:table-cell">{{ h.proxied ? 'Com proxy' : 'Só DNS' }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>
