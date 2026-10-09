<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Bell, Mail, Plus, Send } from 'lucide-vue-next'
import { api, errorMessage } from '../api/client'
import LoadError from '../components/LoadError.vue'
import EmptyState from '../components/EmptyState.vue'
import FormField from '../components/FormField.vue'
import FormModal from '../components/FormModal.vue'
import PageHeader from '../components/PageHeader.vue'
import SkeletonRows from '../components/SkeletonRows.vue'
import { useToast } from '../composables/useToast'
import * as v from '../composables/validation'

type Channel = { id: string; name: string; type: string; created_by_admin?: boolean }
type Rule = { id: string; trigger: string; scope_type: string; created_by_admin?: boolean }

const toast = useToast()
const channels = ref<Channel[]>([])
const rules = ref<Rule[]>([])
const loading = ref(true)
const loadError = ref('')
let reloading = false
const busy = ref(false)
const error = ref('')
const testing = ref('')
const showForm = ref(false)
const attempted = ref(false)
const form = ref({ type: 'email', name: '', destination: '' })
const errs = computed(() => ({
  name: v.name('Nome', form.value.name),
  destination: v.channelDestination(form.value.type, form.value.destination),
}))
const destinationPlaceholder = computed(() => (form.value.type === 'email' ? 'Ex.: voce@exemplo.com' : 'Ex.: 123456789 ou @meucanal'))

const triggerLabel: Record<string, string> = {
  no_contact: 'Host sem atualizações',
  update_failure: 'Falha na atualização',
  ip_changed: 'IP alterado',
  token_expiring: 'Token próximo do vencimento',
  connection_invalid: 'Problema na conexão com a Cloudflare',
}
const scopeLabel: Record<string, string> = { all: 'todos os hosts', host: 'um host', user: 'um usuário' }

async function reload() {
  if (reloading) return
  reloading = true
  loading.value = true
  try {
    ;[channels.value, rules.value] = await Promise.all([
      api<Channel[]>('/api/alert-channels').then((r) => r ?? []),
      api<Rule[]>('/api/alert-rules').then((r) => r ?? []),
    ])
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
  form.value = { type: 'email', name: '', destination: '' }
  error.value = ''
  attempted.value = false
  showForm.value = true
}

async function createChannel() {
  if (busy.value) return
  error.value = ''
  attempted.value = true
  if (errs.value.name || errs.value.destination) return
  busy.value = true
  try {
    await api('/api/alert-channels', { method: 'POST', body: JSON.stringify({ ...form.value, name: form.value.name.trim(), destination: form.value.destination.trim() }) })
    toast.success('Canal cadastrado. Envie um teste para confirmar.')
    showForm.value = false
    await reload()
  } catch (e) {
    error.value = errorMessage(e, 'Não foi possível cadastrar o canal.')
  } finally {
    busy.value = false
  }
}

async function test(c: Channel) {
  if (testing.value) return
  testing.value = c.id
  try {
    const res = await api<{ status: string; error?: string }>(`/api/alert-channels/${c.id}/test`, { method: 'POST' })
    if (res.status === 'ok') toast.success(`Teste enviado para ${c.name}.`)
    else toast.error(`Não foi possível enviar o teste para ${c.name}. Confira o destino e peça ao administrador para verificar a configuração de envio.`)
  } catch (e) {
    toast.error(errorMessage(e))
  } finally {
    testing.value = ''
  }
}
</script>

<template>
  <div>
    <PageHeader title="Alertas" description="Configure e teste o envio de mensagens por e-mail ou Telegram.">
      <template #actions>
        <button type="button" class="btn btn-primary gap-2" @click="openForm"><Plus :size="18" aria-hidden="true" />Novo canal</button>
      </template>
    </PageHeader>
    <LoadError v-if="loadError" :message="loadError" :busy="loading" @retry="reload" />
    <template v-if="!loadError">
      <p class="mb-5 text-sm text-base-content/80">Por enquanto, apenas mensagens de teste são enviadas. O envio automático por regras ainda não está disponível.</p>
      <div>
        <div class="space-y-6">
          <section class="surface" aria-label="Canais">
            <h2 class="font-display text-lg font-bold p-4 pb-0">Canais</h2>
            <SkeletonRows v-if="loading" :rows="2" />
            <EmptyState v-else-if="!channels.length" title="Nenhum canal" description="Adicione um e-mail ou uma conversa do Telegram e envie uma mensagem de teste.">
              <template #icon><Bell :size="28" /></template>
              <button type="button" class="btn btn-primary" @click="openForm">Novo canal</button>
            </EmptyState>
            <ul v-else class="divide-y divide-base-300 mt-2">
              <li v-for="c in channels" :key="c.id" class="flex flex-wrap items-center gap-3 p-4">
                <component :is="c.type === 'email' ? Mail : Send" :size="18" class="text-base-content/70" aria-hidden="true" />
                <div class="min-w-0 flex-1 basis-40">
                  <p class="font-semibold truncate">{{ c.name }}</p>
                  <p class="text-sm text-base-content/65">{{ c.type === 'email' ? 'E-mail' : 'Telegram' }}<span v-if="c.created_by_admin"> · definido pelo administrador</span></p>
                </div>
                <button type="button" class="btn btn-sm btn-ghost" :disabled="testing === c.id" @click="test(c)">
                  <span v-if="testing === c.id" class="loading loading-spinner loading-xs" />Enviar teste
                </button>
              </li>
            </ul>
          </section>
          <section class="surface" aria-label="Regras">
            <h2 class="font-display text-lg font-bold p-4 pb-0">Regras cadastradas</h2>
            <p v-if="!loading && !rules.length" class="p-4 text-base-content/70">Nenhuma regra definida ainda.</p>
            <ul v-else class="divide-y divide-base-300 mt-2">
              <li v-for="r in rules" :key="r.id" class="flex flex-wrap items-center gap-2 p-4">
                <span class="font-semibold">{{ triggerLabel[r.trigger] ?? r.trigger }}</span>
                <span class="text-sm text-base-content/65">em {{ scopeLabel[r.scope_type] ?? r.scope_type }}</span>
                <span v-if="r.created_by_admin" class="badge badge-info badge-outline ml-auto">Definida pelo administrador</span>
              </li>
            </ul>
          </section>
        </div>

      </div>
    </template>

    <FormModal
      :open="showForm"
      title="Novo canal"
      size="wide"
      description="Depois de cadastrar, envie um teste para confirmar que o aviso chega."
      submit-label="Cadastrar canal"
      busy-label="Cadastrando…"
      :busy="busy"
      :error="error"
      @submit="createChannel"
      @cancel="showForm = false"
    >
      <div class="form-grid">
        <FormField v-slot="{ id }" label="Tipo">
          <select :id="id" v-model="form.type" class="select select-bordered w-full">
            <option value="email">E-mail</option>
            <option value="telegram">Telegram</option>
          </select>
        </FormField>
        <FormField v-slot="{ id, describedBy, invalid }" label="Nome" hint="Para reconhecer o canal na lista." :error="errs.name" :submitted="attempted">
          <input :id="id" v-model="form.name" class="input input-bordered w-full" :placeholder="form.type === 'email' ? 'Ex.: Meu e-mail' : 'Ex.: Avisos no Telegram'" maxlength="80" autocomplete="off" required :aria-invalid="invalid" :aria-describedby="describedBy" />
        </FormField>
      </div>
      <FormField v-slot="{ id, describedBy, invalid }" :label="form.type === 'email' ? 'Endereço de e-mail' : 'Conversa ou canal do Telegram'" :hint="form.type === 'email' ? 'Para qual e-mail devemos enviar a mensagem?' : 'Use o número da conversa (Chat ID) ou o nome do canal com @.'" :error="errs.destination" :submitted="attempted">
        <input
          :id="id"
          v-model="form.destination"
          class="input input-bordered w-full"
          :type="form.type === 'email' ? 'email' : 'text'"
          :inputmode="form.type === 'email' ? 'email' : 'text'"
          autocapitalize="none"
          :placeholder="destinationPlaceholder"
          maxlength="254"
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
