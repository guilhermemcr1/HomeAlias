<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { Cable, Pencil, Plus, RefreshCw, Trash2 } from 'lucide-vue-next'
import { api, errorMessage } from '../api/client'
import LoadError from '../components/LoadError.vue'
import EmptyState from '../components/EmptyState.vue'
import FormField from '../components/FormField.vue'
import PasswordInput from '../components/PasswordInput.vue'
import FormModal from '../components/FormModal.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
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
const editingId = ref('')
const errs = computed(() => ({ name: v.name('Nome', name.value), token: editingId.value && !token.value.trim() ? '' : v.apiToken(token.value) }))
const testing = ref('')
const removing = ref<Connection | null>(null)
const deleting = ref(false)
const deleteError = ref('')

function confirmRemoval(c: Connection) {
  deleteError.value = ''
  removing.value = c
}

async function remove() {
  if (deleting.value || !removing.value) return
  deleting.value = true
  deleteError.value = ''
  const connection = removing.value
  try {
    await api(`/api/connections/${connection.id}`, { method: 'DELETE' })
    list.value = list.value.filter(c => c.id !== connection.id)
    removing.value = null
    toast.success('Conexão excluída.')
  } catch (e) {
    deleteError.value = errorMessage(e)
  } finally {
    deleting.value = false
  }
}

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
  editingId.value = ''
  name.value = ''
  token.value = ''
  error.value = ''
  attempted.value = false
  showForm.value = true
}

function openEdit(c: Connection) {
  editingId.value = c.id
  name.value = c.name
  token.value = ''
  error.value = ''
  attempted.value = false
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  token.value = ''
}

async function create() {
  if (busy.value) return
  error.value = ''
  attempted.value = true
  if (errs.value.name || errs.value.token) return
  busy.value = true
  try {
    await api(editingId.value ? `/api/connections/${editingId.value}` : '/api/connections', {
      method: editingId.value ? 'PATCH' : 'POST',
      body: JSON.stringify({ name: name.value.trim(), ...(!editingId.value || token.value.trim() ? { api_token: v.normalizeApiToken(token.value) } : {}), test: true }),
    })
    closeForm()
    toast.success(editingId.value ? 'Conexão atualizada. Seus hosts continuam usando esta conta.' : 'Conexão salva. Você já pode adicionar um host.')
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
      description="Conecte sua conta da Cloudflare para manter seus endereços atualizados."
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
            description="Adicione uma conexão para usar os domínios da sua conta Cloudflare."
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
                  · verificação: {{ c.last_checked_at ? relativeTime(c.last_checked_at) : 'ainda não realizada' }}
                </p>
                <div v-if="c.last_error" class="mt-1 text-sm">
                  <p class="text-error">Não foi possível conferir esta conexão. Verifique o token e tente testar novamente.</p>
                  <details class="mt-1"><summary class="cursor-pointer">Detalhes técnicos</summary><p class="mt-1 break-words">{{ c.last_error }}</p></details>
                </div>
              </div>
              <StatusBadge :status="c.status" />
              <button type="button" class="btn btn-sm btn-ghost gap-1.5" @click="openEdit(c)"><Pencil :size="14" aria-hidden="true" />Editar conexão</button>
              <button type="button" class="btn btn-sm btn-ghost gap-1.5" :disabled="testing === c.id" @click="retest(c)">
                <RefreshCw :size="14" :class="{ 'animate-spin': testing === c.id }" aria-hidden="true" />Testar
              </button>
              <button type="button" class="btn btn-sm btn-ghost text-error gap-1.5" :disabled="!!testing" @click="confirmRemoval(c)"><Trash2 :size="14" aria-hidden="true" />Excluir conexão</button>
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
      :title="editingId ? 'Editar conexão' : 'Nova conexão'"
      size="wide"
      :description="editingId ? 'Para trocar apenas o nome, deixe o token em branco. Se informar outro token, vamos conferir se ele funciona antes de salvar.' : 'Vamos conferir o token na Cloudflare antes de salvar. Depois disso, ele ficará oculto.'"
      submit-label="Validar e salvar"
      busy-label="Validando…"
      :busy="busy"
      :error="error"
      @submit="create"
      @cancel="closeForm"
    >
      <FormField v-slot="{ id, describedBy, invalid }" label="Nome" hint="Escolha um nome fácil de reconhecer." :error="errs.name" :submitted="attempted">
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
        :hint="editingId ? 'Opcional. O novo token precisa acessar os domínios dos seus hosts. Permissões na Cloudflare: Zone · Zone · Read e Zone · DNS · Edit.' : 'Na Cloudflare, abra Meu perfil → Tokens de API e crie um token. Permissões necessárias: Zone · Zone · Read e Zone · DNS · Edit.'"
        :error="errs.token" :submitted="attempted"
      >
        <PasswordInput secret-label="token"
          :id="id"
          v-model="token"
          class="input input-bordered w-full font-data"
          placeholder="Cole o token gerado na Cloudflare"
          maxlength="256"
          autocomplete="off"
          spellcheck="false"
          :required="!editingId"
          :aria-invalid="invalid"
          :aria-describedby="describedBy"
        />
      </FormField>
    </FormModal>
    <ConfirmDialog
      :open="!!removing"
      title="Excluir conexão?"
      :confirm-label="deleting ? 'Excluindo…' : 'Excluir conexão'"
      danger
      :busy="deleting"
      @confirm="remove"
      @cancel="removing = null"
    >
      <p>Você vai excluir <strong class="break-words">{{ removing?.name }}</strong> do HomeAlias. A conta e os registros DNS na Cloudflare continuam intactos.</p>
      <p class="mt-3 text-sm">Se houver hosts vinculados, remova-os na aba Hosts antes de excluir esta conexão.</p>
      <p v-if="deleteError" class="mt-3 text-sm text-error break-words" role="alert">{{ deleteError }}</p>
    </ConfirmDialog>
  </div>
</template>
