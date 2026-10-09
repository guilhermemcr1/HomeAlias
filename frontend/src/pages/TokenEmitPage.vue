<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { KeyRound, Plus, TriangleAlert } from 'lucide-vue-next'
import { api, errorMessage } from '../api/client'
import ClientInstructions from '../components/ClientInstructions.vue'
import CodeBlock from '../components/CodeBlock.vue'
import LoadError from '../components/LoadError.vue'
import EmptyState from '../components/EmptyState.vue'
import FormField from '../components/FormField.vue'
import FormModal from '../components/FormModal.vue'
import PageHeader from '../components/PageHeader.vue'
import * as v from '../composables/validation'

type Host = { id: string; fqdn: string; enable_a: boolean; enable_aaaa: boolean }
type Issued = { token: string; token_prefix: string; instructions: Record<string, string> }

const route = useRoute()
const hosts = ref<Host[]>([])
const loaded = ref(false)
const loadError = ref('')
const name = ref('')
const hostId = ref(typeof route.query.host === 'string' ? route.query.host : '')
const issued = ref<Issued | null>(null)
const busy = ref(false)
const error = ref('')
const showForm = ref(false)
const attempted = ref(false)
const errs = computed(() => ({ host: hosts.value.some((h) => h.id === hostId.value) ? '' : 'Escolha um host disponível.', name: v.name('Nome do token', name.value) }))
const selectedHost = computed(() => hosts.value.find((h) => h.id === hostId.value))
const hostName = computed(() => selectedHost.value?.fqdn ?? '')

let reloading = false
async function reload() {
  if (reloading) return
  reloading = true
  loaded.value = false
  loadError.value = ''
  try {
    hosts.value = (await api<Host[]>('/api/hosts')) ?? []
    if (!hostId.value && hosts.value.length === 1) hostId.value = hosts.value[0].id
    if (hosts.value.length) showForm.value = true
  } catch (e) {
    loadError.value = errorMessage(e)
  } finally {
    loaded.value = true
    reloading = false
  }
}
onMounted(reload)

async function emit() {
  if (busy.value) return
  error.value = ''
  attempted.value = true
  if (errs.value.host || errs.value.name) return
  busy.value = true
  try {
    issued.value = await api<Issued>('/api/tokens', {
      method: 'POST',
      body: JSON.stringify({ name: name.value.trim(), host_ids: [hostId.value] }),
    })
    showForm.value = false
  } catch (e) {
    error.value = errorMessage(e, 'Não foi possível gerar o token.')
  } finally {
    busy.value = false
  }
}

function again() {
  issued.value = null
  attempted.value = false
  error.value = ''
  name.value = ''
  showForm.value = true
}
</script>

<template>
  <div>
    <PageHeader
      title="Tokens DDNS"
      description="Um token é um código de acesso que permite ao seu dispositivo atualizar um host. Gere um para cada dispositivo."
    >
      <template v-if="hosts.length && !issued" #actions>
        <button type="button" class="btn btn-primary gap-2" @click="showForm = true"><Plus :size="18" aria-hidden="true" />Novo token</button>
      </template>
    </PageHeader>
    <LoadError v-if="loadError" :message="loadError" :busy="!loaded" @retry="reload" />
    <p v-if="!loaded" role="status" class="py-6 text-base-content/70">Carregando hosts…</p>

    <EmptyState
      v-if="!loadError && loaded && !hosts.length"
      class="surface"
      title="Crie um host antes de gerar o token"
      description="Primeiro, adicione o endereço que seu dispositivo vai atualizar."
    >
      <RouterLink to="/hosts" class="btn btn-primary">Ir para Hosts</RouterLink>
    </EmptyState>

    <div v-else-if="issued" class="surface space-y-5 p-5 sm:p-6">
      <div class="flex items-start gap-3 rounded-btn border border-warning/60 bg-warning/10 p-3" role="status">
        <TriangleAlert :size="20" class="mt-0.5 shrink-0 text-warning" aria-hidden="true" />
        <p class="text-sm">
          <strong>Copie o token agora.</strong> Por segurança ele não aparece de novo; se perder, gere outro.
        </p>
      </div>
      <CodeBlock :code="issued.token" :mask="issued.token" :caption="`Token para ${hostName}`" />
      <div>
        <h2 class="font-display text-lg font-bold mb-3">Configure seu dispositivo</h2>
        <ClientInstructions :token="issued.token" :hostname="hostName" :ipv4-enabled="selectedHost?.enable_a ?? true" :ipv6-enabled="selectedHost?.enable_aaaa ?? false" />
      </div>
      <div class="flex flex-wrap gap-2 pt-1">
        <RouterLink to="/" class="btn btn-primary">Ir para o painel</RouterLink>
        <button type="button" class="btn btn-ghost" @click="again">Gerar outro token</button>
      </div>
    </div>

    <EmptyState v-else-if="!loadError && loaded" class="surface" title="Gere um token para começar" description="Por segurança o token só aparece uma vez, logo após ser gerado.">
      <template #icon><KeyRound :size="28" /></template>
      <button type="button" class="btn btn-primary" @click="showForm = true">Gerar token</button>
    </EmptyState>

    <FormModal
      :open="showForm"
      title="Novo token DDNS"
      size="wide"
      description="O token aparece uma única vez, logo após ser gerado."
      submit-label="Gerar token"
      busy-label="Gerando…"
      :busy="busy"
      :error="error"
      @submit="emit"
      @cancel="showForm = false"
    >
      <div class="form-grid">
        <FormField v-slot="{ id, describedBy, invalid }" label="Host" :error="errs.host" :submitted="attempted">
          <select :id="id" v-model="hostId" class="select select-bordered w-full" required :aria-invalid="invalid" :aria-describedby="describedBy">
            <option disabled value="">Escolha um host…</option>
            <option v-for="h in hosts" :key="h.id" :value="h.id">{{ h.fqdn }}</option>
          </select>
        </FormField>
        <FormField v-slot="{ id, describedBy, invalid }" label="Nome do token" hint="Use o nome do dispositivo que fará as atualizações." :error="errs.name" :submitted="attempted">
          <input :id="id" v-model="name" class="input input-bordered w-full" placeholder="Ex.: roteador, NAS" maxlength="80" autocomplete="off" required :aria-invalid="invalid" :aria-describedby="describedBy" />
        </FormField>
      </div>
    </FormModal>
  </div>
</template>
