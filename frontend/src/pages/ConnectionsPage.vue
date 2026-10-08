<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { Cable, Plus, RefreshCw } from 'lucide-vue-next'
import { api, errorMessage } from '../api/client'
import LoadError from '../components/LoadError.vue'
import EmptyState from '../components/EmptyState.vue'
import FormField from '../components/FormField.vue'
import FormModal from '../components/FormModal.vue'
import PageHeader from '../components/PageHeader.vue'
import SkeletonRows from '../components/SkeletonRows.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { relativeTime } from '../composables/format'
import { useToast } from '../composables/useToast'
import * as v from '../composables/validation'

type Connection = {
  id: string
  name: string
  status: string
  api_token_suffix: string
  last_checked_at?: string | null
  last_error?: string | null
}

const toast = useToast()
const list = ref<Connection[]>([])
const loading = ref(true)
const loadError = ref('')
let reloading = false
const showForm = ref(false)
const name = ref('')
const token = ref('')
const busy = ref(false)
const error = ref('')
const attempted = ref(false)
const errs = computed(() => ({ name: v.name('Nome', name.value), token: v.apiToken(token.value) }))
const testing = ref('')

async function reload() {
  if (reloading) return
  reloading = true
  loading.value = true
  try {
    list.value = (await api<Connection[]>('/api/connections')) ?? []
    loadError.value = ''
  } catch (e) {
    loadError.value = errorMessage(e)
  } finally {
    loading.value = false
    reloading = false
  }
}
onMounted(reload)

function openForm() {
  name.value = ''
  token.value = ''
  error.value = ''
  attempted.value = false
  showForm.value = true
}

async function create() {
  if (busy.value) return
  error.value = ''
  attempted.value = true
  if (errs.value.name || errs.value.token) return
  busy.value = true
  try {
    await api('/api/connections', {
      method: 'POST',
      body: JSON.stringify({ name: name.value.trim(), api_token: v.normalizeApiToken(token.value), test: true }),
    })
    showForm.value = false
    toast.success('Conexão validada e salva. O token não será exibido de novo.')
    await reload()
  } catch (e) {
    error.value = errorMessage(e, 'A Cloudflare recusou o token.')
  } finally {
    busy.value = false
  }
}

async function retest(c: Connection) {
  if (testing.value) return
  testing.value = c.id
  try {
    await api(`/api/connections/${c.id}/test`, { method: 'POST' })
    toast.success(`${c.name}: conexão válida.`)
  } catch (e) {
    toast.error(`${c.name}: ${errorMessage(e)}`)
  } finally {
    testing.value = ''
    await reload()
  }
}
</script>

<template>
  <div>
    <PageHeader
      title="Conexões"
      description="Uma conexão é um token da sua conta Cloudflare. Com ela o HomeAlias cria e atualiza os registros DNS."
    >
      <template #actions>
        <button type="button" class="btn btn-primary gap-2" @click="openForm"><Plus :size="18" aria-hidden="true" />Nova conexão</button>
      </template>
    </PageHeader>
    <LoadError v-if="loadError" :message="loadError" :busy="loading" @retry="reload" />
    <template v-if="!loadError">
      <div>
        <section class="surface" aria-label="Conexões cadastradas">
          <SkeletonRows v-if="loading" :rows="3" />
          <EmptyState
            v-else-if="!list.length"
            title="Nenhuma conexão ainda"
            description="Cadastre o primeiro token. Depois você poderá criar hosts nas zonas dele."
          >
            <template #icon><Cable :size="28" /></template>
            <button type="button" class="btn btn-primary" @click="openForm">Nova conexão</button>
          </EmptyState>
          <ul v-else class="divide-y divide-base-300">
            <li v-for="c in list" :key="c.id" class="flex flex-wrap items-center gap-x-4 gap-y-2 p-4">
              <div class="min-w-0 flex-1 basis-48">
                <p class="font-semibold truncate">{{ c.name }}</p>
                <p class="text-sm text-base-content/65">
                  Token <span class="font-data">{{ c.api_token_suffix }}</span>
                  · verificada {{ relativeTime(c.last_checked_at) }}
                </p>
                <p v-if="c.last_error" class="mt-1 text-sm text-error">{{ c.last_error }}</p>
              </div>
              <StatusBadge :status="c.status" />
              <button type="button" class="btn btn-sm btn-ghost gap-1.5" :disabled="testing === c.id" @click="retest(c)">
                <RefreshCw :size="14" :class="{ 'animate-spin': testing === c.id }" aria-hidden="true" />Testar
              </button>
            </li>
          </ul>
          <div v-if="list.length" class="border-t border-base-300 p-4 text-sm">
            Próximo passo: <RouterLink to="/hosts" class="link link-primary">criar um host</RouterLink>.
          </div>
        </section>
      </div>

    </template>

    <FormModal
      :open="showForm"
      title="Nova conexão"
      description="O token é validado na Cloudflare antes de ser salvo e nunca é exibido de novo."
      submit-label="Validar e salvar"
      busy-label="Validando…"
      :busy="busy"
      :error="error"
      @submit="create"
      @cancel="showForm = false"
    >
      <FormField v-slot="{ id, describedBy, invalid }" label="Nome" hint="Só para você identificar." :error="attempted ? errs.name : ''">
        <input
          :id="id"
          v-model="name"
          class="input input-bordered w-full"
          placeholder="Ex.: Conta pessoal"
          maxlength="80"
          autocomplete="off"
          required
          :aria-invalid="invalid"
          :aria-describedby="describedBy"
        />
      </FormField>
      <FormField
        v-slot="{ id, describedBy, invalid }"
        label="Token de API da Cloudflare"
        hint="Crie em Cloudflare → Meu perfil → Tokens de API, com a permissão Zone · DNS · Edit."
        :error="attempted ? errs.token : ''"
      >
        <input
          :id="id"
          v-model="token"
          type="password"
          class="input input-bordered w-full font-data"
          placeholder="Cole o token gerado na Cloudflare"
          maxlength="256"
          autocomplete="off"
          spellcheck="false"
          required
          :aria-invalid="invalid"
          :aria-describedby="describedBy"
        />
      </FormField>
    </FormModal>
  </div>
</template>
