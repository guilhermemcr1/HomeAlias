<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { KeyRound, Pencil, Plus, Users } from 'lucide-vue-next'
import { RouterLink } from 'vue-router'
import { api, errorMessage } from '../api/client'
import CodeBlock from '../components/CodeBlock.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import LoadError from '../components/LoadError.vue'
import EmptyState from '../components/EmptyState.vue'
import FormField from '../components/FormField.vue'
import PasswordInput from '../components/PasswordInput.vue'
import FormModal from '../components/FormModal.vue'
import PageHeader from '../components/PageHeader.vue'
import SkeletonRows from '../components/SkeletonRows.vue'
import { useAuth } from '../composables/useAuth'
import { useToast } from '../composables/useToast'
import * as v from '../composables/validation'

type U = { id: string; email: string; name?: string; role: string; status: string }

const toast = useToast()
const { user: me } = useAuth()
const users = ref<U[]>([])
const loading = ref(true)
const loadError = ref('')
let reloading = false
const showForm = ref(false)
const busy = ref(false)
const error = ref('')
const form = ref({ email: '', name: '', initial_password: '', send_invite: true })
const inviteLink = ref('')
const attempted = ref(false)
watch(showForm, (open) => { if (!open) form.value.initial_password = '' })
watch(() => form.value.send_invite, (invite) => { if (invite) form.value.initial_password = '' })
const errs = computed(() => ({
  email: v.email(form.value.email),
  name: v.name('Nome', form.value.name),
  password: form.value.send_invite ? '' : v.password(form.value.initial_password),
}))

const statusLabel: Record<string, string> = { active: 'Ativo', invite_pending: 'Convite pendente', disabled: 'Desativado' }
const statusCls: Record<string, string> = { active: 'badge-success', invite_pending: 'badge-warning', disabled: 'badge-ghost' }

async function reload() {
  if (reloading) return
  reloading = true
  loading.value = true
  try {
    users.value = (await api<U[]>('/api/users')) ?? []
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
  form.value = { email: '', name: '', initial_password: '', send_invite: true }
  error.value = ''
  attempted.value = false
  showForm.value = true
}

async function create() {
  if (busy.value) return
  error.value = ''
  attempted.value = true
  if (Object.values(errs.value).some(Boolean)) return
  inviteLink.value = ''
  busy.value = true
  try {
    const res = await api<{ invite_token?: string }>('/api/users', {
      method: 'POST',
      body: JSON.stringify({ ...form.value, email: form.value.email.trim().toLowerCase(), name: form.value.name.trim(), initial_password: form.value.send_invite ? '' : form.value.initial_password }),
    })
    if (res?.invite_token) inviteLink.value = `${window.location.origin}/convite?token=${res.invite_token}`
    toast.success('Usuário criado.')
    showForm.value = false
    await reload()
  } catch (e) {
    error.value = errorMessage(e, 'Não foi possível criar o usuário.')
  } finally {
    busy.value = false
  }
}

// --- editar ---
const editing = ref<U | null>(null)
const edit = ref({ name: '', email: '', role: 'user' })
const editBusy = ref(false)
const editError = ref('')
const editTried = ref(false)
const editErrs = computed(() => ({ name: v.name('Nome', edit.value.name), email: v.email(edit.value.email) }))

function askEdit(u: U) {
  edit.value = { name: u.name ?? '', email: u.email, role: u.role }
  editError.value = ''
  editTried.value = false
  editing.value = u
}

async function saveEdit() {
  if (editBusy.value) return
  if (!editing.value) return
  editError.value = ''
  editTried.value = true
  if (Object.values(editErrs.value).some(Boolean)) return
  editBusy.value = true
  try {
    await api(`/api/users/${editing.value.id}`, {
      method: 'PATCH',
      body: JSON.stringify({ name: edit.value.name.trim(), email: edit.value.email.trim().toLowerCase(), role: edit.value.role }),
    })
    toast.success('Usuário atualizado.')
    editing.value = null
    await reload()
  } catch (e) {
    editError.value = errorMessage(e, 'Não foi possível salvar.')
  } finally {
    editBusy.value = false
  }
}

// --- redefinir senha ---
const resetting = ref<U | null>(null)
const reset = ref({ next: '', confirm: '' })
const resetBusy = ref(false)
const resetError = ref('')
const resetTried = ref(false)
watch(resetting, (user) => { if (!user) reset.value = { next: '', confirm: '' } })
const resetErrs = computed(() => ({ next: v.password(reset.value.next), confirm: v.passwordMatch(reset.value.next, reset.value.confirm) }))

function askReset(u: U) {
  reset.value = { next: '', confirm: '' }
  resetError.value = ''
  resetTried.value = false
  resetting.value = u
}

async function saveReset() {
  if (resetBusy.value) return
  if (!resetting.value) return
  resetError.value = ''
  resetTried.value = true
  if (Object.values(resetErrs.value).some(Boolean)) return
  resetBusy.value = true
  try {
    await api(`/api/users/${resetting.value.id}/password`, { method: 'POST', body: JSON.stringify({ new_password: reset.value.next }) })
    toast.success(`Senha de ${resetting.value.email} redefinida. Os acessos dessa conta foram encerrados.`)
    resetting.value = null
    await reload()
  } catch (e) {
    resetError.value = errorMessage(e, 'Não foi possível redefinir a senha.')
  } finally {
    resetBusy.value = false
  }
}

const statusBusy = ref(false)
async function enable(u: U) {
  if (statusBusy.value) return
  statusBusy.value = true
  try {
    await api(`/api/users/${u.id}/enable`, { method: 'POST' })
    toast.success(`${u.email} reativado.`)
    await reload()
  } catch (e) {
    toast.error(errorMessage(e))
  } finally {
    statusBusy.value = false
  }
}

const disabling = ref<U | null>(null)
async function confirmDisable() {
  if (!disabling.value || statusBusy.value) return
  statusBusy.value = true
  try {
    await api(`/api/users/${disabling.value.id}/disable`, { method: 'POST' })
    toast.success(`${disabling.value.email} desativado.`)
    disabling.value = null
    await reload()
  } catch (e) {
    toast.error(errorMessage(e))
  } finally {
    statusBusy.value = false
  }
}
</script>

<template>
  <div>
    <PageHeader title="Usuários" description="Adicione as pessoas que podem acessar o HomeAlias. Você pode gerar um convite ou definir uma senha inicial.">
      <template #actions>
        <button type="button" class="btn btn-primary gap-2" @click="openForm"><Plus :size="18" aria-hidden="true" />Novo usuário</button>
      </template>
    </PageHeader>
    <LoadError v-if="loadError" :message="loadError" :busy="loading" @retry="reload" />
    <template v-if="!loadError">

      <div v-if="inviteLink" class="surface mb-6 space-y-3 p-5">
        <p class="font-semibold">Copie o link e envie para a pessoa. Por segurança, ele só aparece desta vez.</p>
        <CodeBlock :code="inviteLink" caption="Link de convite" />
      </div>

      <section class="surface">
        <SkeletonRows v-if="loading" />
        <EmptyState v-else-if="!users.length" title="Nenhum usuário"><template #icon><Users :size="28" /></template></EmptyState>
        <ul v-else class="divide-y divide-base-300">
          <li v-for="u in users" :key="u.id" class="flex flex-wrap items-center gap-x-4 gap-y-2 p-4">
            <div class="min-w-0 flex-1 basis-56">
              <p class="font-semibold truncate">{{ u.name || u.email }}</p>
              <p class="text-sm text-base-content/65 truncate">{{ u.email }} · {{ u.role === 'admin' ? 'Administrador' : 'Usuário' }}</p>
            </div>
            <span class="badge badge-outline h-7" :class="statusCls[u.status]">{{ statusLabel[u.status] ?? u.status }}</span>
            <div class="flex flex-wrap gap-1">
              <RouterLink v-if="u.id === me?.id" to="/conta" class="btn btn-sm btn-ghost">Minha conta</RouterLink>
              <template v-else>
                <button type="button" class="btn btn-sm btn-ghost gap-1.5" @click="askEdit(u)"><Pencil :size="14" aria-hidden="true" />Editar</button>
                <button type="button" class="btn btn-sm btn-ghost gap-1.5" @click="askReset(u)"><KeyRound :size="14" aria-hidden="true" />Senha</button>
                <button v-if="u.status === 'disabled'" type="button" class="btn btn-sm btn-ghost" :disabled="statusBusy" @click="enable(u)">Reativar</button>
                <button v-else-if="u.role === 'user'" type="button" class="btn btn-sm btn-ghost text-error" :disabled="statusBusy" @click="disabling = u">Desativar</button>
              </template>
            </div>
          </li>
        </ul>
      </section>

    </template>

    <FormModal
      :open="showForm"
      title="Novo usuário"
      size="wide"
      description="Gere um convite para a pessoa escolher a senha ou defina uma senha inicial."
      submit-label="Criar usuário"
      busy-label="Criando…"
      :busy="busy"
      :error="error"
      @submit="create"
      @cancel="showForm = false"
    >
      <div class="form-grid">
        <FormField v-slot="{ id, describedBy, invalid }" label="E-mail" :error="errs.email" :submitted="attempted">
          <input :id="id" v-model="form.email" type="email" inputmode="email" autocapitalize="none" spellcheck="false" class="input input-bordered w-full" placeholder="Ex.: ana@exemplo.com" maxlength="254" autocomplete="off" required :aria-invalid="invalid" :aria-describedby="describedBy" />
        </FormField>
        <FormField v-slot="{ id, describedBy, invalid }" label="Nome" :error="errs.name" :submitted="attempted">
          <input :id="id" v-model="form.name" class="input input-bordered w-full" placeholder="Ex.: Ana Souza" maxlength="80" autocomplete="off" required :aria-invalid="invalid" :aria-describedby="describedBy" />
        </FormField>
      </div>
      <label class="flex items-start gap-3 cursor-pointer">
        <input v-model="form.send_invite" type="checkbox" class="checkbox checkbox-primary checkbox-sm mt-0.5" />Gerar link de convite (a pessoa define a própria senha)
      </label>
      <FormField v-if="!form.send_invite" v-slot="{ id, describedBy, invalid }" label="Senha inicial" hint="Mínimo de 12 caracteres. Uma frase longa é melhor que uma senha curta e complicada. Compartilhe a senha por um meio seguro." :error="errs.password" :submitted="attempted">
        <PasswordInput secret-label="senha inicial" :id="id" v-model="form.initial_password" class="input input-bordered w-full" placeholder="Pelo menos 12 caracteres" maxlength="128" autocomplete="new-password" required :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
    </FormModal>

    <FormModal
      :open="!!editing"
      :title="editing ? `Editar ${editing.name || editing.email}` : 'Editar usuário'"
      size="wide"
      description="Mudar e-mail ou perfil de acesso desconecta a pessoa."
      submit-label="Salvar"
      busy-label="Salvando…"
      :busy="editBusy"
      :error="editError"
      @submit="saveEdit"
      @cancel="editing = null"
    >
      <div class="form-grid">
        <FormField v-slot="{ id, describedBy, invalid }" label="Nome" :error="editErrs.name" :submitted="editTried">
          <input :id="id" v-model="edit.name" class="input input-bordered w-full" placeholder="Ex.: Ana Souza" maxlength="80" autocomplete="off" required :aria-invalid="invalid" :aria-describedby="describedBy" />
        </FormField>
        <FormField v-slot="{ id, describedBy, invalid }" label="E-mail" :error="editErrs.email" :submitted="editTried">
          <input :id="id" v-model="edit.email" type="email" inputmode="email" autocapitalize="none" spellcheck="false" class="input input-bordered w-full" placeholder="Ex.: ana@exemplo.com" maxlength="254" autocomplete="off" required :aria-invalid="invalid" :aria-describedby="describedBy" />
        </FormField>
      </div>
      <FormField v-slot="{ id }" label="Perfil de acesso">
        <select :id="id" v-model="edit.role" class="select select-bordered w-full">
          <option value="user">Usuário</option>
          <option value="admin">Administrador</option>
        </select>
      </FormField>
    </FormModal>

    <FormModal
      :open="!!resetting"
      :title="resetting ? `Redefinir senha de ${resetting.name || resetting.email}` : 'Redefinir senha'"
      description="A pessoa será desconectada e entra com a nova senha. Compartilhe a nova senha por um meio seguro."
      submit-label="Redefinir senha"
      busy-label="Redefinindo…"
      :busy="resetBusy"
      :error="resetError"
      @submit="saveReset"
      @cancel="resetting = null"
    >
      <FormField v-slot="{ id, describedBy, invalid }" label="Nova senha" hint="Mínimo de 12 caracteres. Uma frase longa é melhor que uma senha curta e complicada." :error="resetErrs.next" :submitted="resetTried">
        <PasswordInput secret-label="nova senha" :id="id" v-model="reset.next" class="input input-bordered w-full" placeholder="Pelo menos 12 caracteres" maxlength="128" autocomplete="new-password" required :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
      <FormField v-slot="{ id, describedBy, invalid }" label="Confirmar nova senha" :error="resetErrs.confirm" :submitted="resetTried">
        <PasswordInput secret-label="confirmação de senha" :id="id" v-model="reset.confirm" class="input input-bordered w-full" placeholder="Repita a nova senha" maxlength="128" autocomplete="new-password" required :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
    </FormModal>

    <ConfirmDialog :open="!!disabling" :title="disabling ? `Desativar ${disabling.email}?` : 'Desativar usuário'" confirm-label="Desativar" danger :busy="statusBusy" @confirm="confirmDisable" @cancel="disabling = null">
      A pessoa perde o acesso ao painel. Os dados dela são mantidos.
    </ConfirmDialog>
  </div>
</template>
