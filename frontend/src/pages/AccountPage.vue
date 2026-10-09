<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { KeyRound, Pencil } from 'lucide-vue-next'
import { api, errorMessage } from '../api/client'
import FormField from '../components/FormField.vue'
import PasswordInput from '../components/PasswordInput.vue'
import FormModal from '../components/FormModal.vue'
import PageHeader from '../components/PageHeader.vue'
import { useAuth } from '../composables/useAuth'
import { useToast } from '../composables/useToast'
import * as v from '../composables/validation'

const toast = useToast()
const route = useRoute()
const router = useRouter()
const { user, isAdmin, refresh } = useAuth()

// --- perfil ---
const showProfile = ref(false)
const profile = ref({ name: '', email: '', current_password: '' })
const profileBusy = ref(false)
const profileError = ref('')
const profileTried = ref(false)
watch(showProfile, (open) => { if (!open) profile.value.current_password = '' })
const emailChanged = computed(() => profile.value.email.trim().toLowerCase() !== (user.value?.email ?? ''))
const profileErrs = computed(() => ({
  name: v.name('Nome', profile.value.name),
  email: v.email(profile.value.email),
  current: emailChanged.value ? v.required('Senha atual', profile.value.current_password) : '',
}))

function openProfile() {
  profile.value = { name: user.value?.name ?? '', email: user.value?.email ?? '', current_password: '' }
  profileError.value = ''
  profileTried.value = false
  showProfile.value = true
}

async function saveProfile() {
  if (profileBusy.value) return
  profileError.value = ''
  profileTried.value = true
  if (Object.values(profileErrs.value).some(Boolean)) return
  profileBusy.value = true
  try {
    await api('/api/auth/me', {
      method: 'PATCH',
      body: JSON.stringify({
        name: profile.value.name.trim(),
        email: profile.value.email.trim().toLowerCase(),
        current_password: emailChanged.value ? profile.value.current_password : '',
      }),
    })
    await refresh()
    showProfile.value = false
    toast.success('Perfil atualizado.')
  } catch (e) {
    profileError.value = errorMessage(e, 'Não foi possível salvar o perfil.')
  } finally {
    profileBusy.value = false
  }
}

// --- senha ---
const showPassword = ref(false)
const pw = ref({ current: '', next: '', confirm: '' })
const pwBusy = ref(false)
const pwError = ref('')
const pwTried = ref(false)
watch(showPassword, (open) => { if (!open) pw.value = { current: '', next: '', confirm: '' } })
const pwErrs = computed(() => ({
  current: v.required('Senha atual', pw.value.current),
  next: v.password(pw.value.next) || (pw.value.next === pw.value.current ? 'A nova senha deve ser diferente da atual.' : ''),
  confirm: v.passwordMatch(pw.value.next, pw.value.confirm),
}))

function openPassword() {
  pw.value = { current: '', next: '', confirm: '' }
  pwError.value = ''
  pwTried.value = false
  showPassword.value = true
}

// O menu do usuário leva para /conta?senha=1; abre o modal e limpa o parâmetro.
function openFromQuery() {
  if (route.query.senha) {
    openPassword()
    router.replace({ path: '/conta' })
  }
}
onMounted(openFromQuery)
watch(() => route.query.senha, openFromQuery)

async function savePassword() {
  if (pwBusy.value) return
  pwError.value = ''
  pwTried.value = true
  if (Object.values(pwErrs.value).some(Boolean)) return
  pwBusy.value = true
  try {
    await api('/api/auth/password', {
      method: 'POST',
      body: JSON.stringify({ current_password: pw.value.current, new_password: pw.value.next }),
    })
    showPassword.value = false
    toast.success('Senha alterada. Seus outros acessos foram encerrados.')
  } catch (e) {
    pwError.value = errorMessage(e, 'Não foi possível alterar a senha.')
  } finally {
    pwBusy.value = false
  }
}
</script>

<template>
  <div>
    <PageHeader title="Minha conta" description="Seus dados de acesso ao painel." />

    <section class="surface max-w-2xl divide-y divide-base-300" aria-label="Dados da conta">
      <div class="flex flex-wrap items-center gap-4 p-5">
        <dl class="min-w-0 flex-1 basis-60 grid gap-3">
          <div>
            <dt class="text-sm text-base-content/65">Nome</dt>
            <dd class="font-semibold break-words">{{ user?.name || '—' }}</dd>
          </div>
          <div>
            <dt class="text-sm text-base-content/65">E-mail</dt>
            <dd class="font-semibold break-all">{{ user?.email }}</dd>
          </div>
          <div>
            <dt class="text-sm text-base-content/65">Perfil de acesso</dt>
            <dd class="font-semibold">{{ isAdmin ? 'Administrador' : 'Usuário' }}</dd>
          </div>
        </dl>
        <button type="button" class="btn btn-ghost gap-2" @click="openProfile"><Pencil :size="16" aria-hidden="true" />Editar perfil</button>
      </div>
      <div class="flex flex-wrap items-center gap-4 p-5">
        <div class="min-w-0 flex-1 basis-60">
          <h2 class="font-semibold">Senha</h2>
          <p class="text-sm text-base-content/70">Ao trocar a senha, os outros dispositivos conectados são desconectados.</p>
        </div>
        <button type="button" class="btn btn-ghost gap-2" @click="openPassword"><KeyRound :size="16" aria-hidden="true" />Alterar senha</button>
      </div>
    </section>

    <FormModal
      :open="showProfile"
      title="Editar perfil"
      size="wide"
      submit-label="Salvar"
      busy-label="Salvando…"
      :busy="profileBusy"
      :error="profileError"
      @submit="saveProfile"
      @cancel="showProfile = false"
    >
      <div class="form-grid">
        <FormField v-slot="{ id, describedBy, invalid }" label="Nome" :error="profileErrs.name" :submitted="profileTried">
          <input :id="id" v-model="profile.name" class="input input-bordered w-full" placeholder="Ex.: Ana Souza" maxlength="80" autocomplete="name" required :aria-invalid="invalid" :aria-describedby="describedBy" />
        </FormField>
        <FormField v-slot="{ id, describedBy, invalid }" label="E-mail" hint="É o e-mail que você usa para entrar." :error="profileErrs.email" :submitted="profileTried">
          <input :id="id" v-model="profile.email" type="email" inputmode="email" autocapitalize="none" spellcheck="false" class="input input-bordered w-full" placeholder="Ex.: ana@exemplo.com" maxlength="254" autocomplete="email" required :aria-invalid="invalid" :aria-describedby="describedBy" />
        </FormField>
      </div>
      <FormField v-if="emailChanged" v-slot="{ id, describedBy, invalid }" label="Senha atual" hint="Necessária para trocar o e-mail." :error="profileErrs.current" :submitted="profileTried">
        <PasswordInput secret-label="senha atual" :id="id" v-model="profile.current_password" class="input input-bordered w-full" placeholder="Sua senha atual" maxlength="128" autocomplete="current-password" required :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
    </FormModal>

    <FormModal
      :open="showPassword"
      title="Alterar senha"
      submit-label="Alterar senha"
      busy-label="Alterando…"
      :busy="pwBusy"
      :error="pwError"
      @submit="savePassword"
      @cancel="showPassword = false"
    >
      <FormField v-slot="{ id, describedBy, invalid }" label="Senha atual" :error="pwErrs.current" :submitted="pwTried">
        <PasswordInput secret-label="senha atual" :id="id" v-model="pw.current" class="input input-bordered w-full" placeholder="Sua senha atual" maxlength="128" autocomplete="current-password" required :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
      <FormField v-slot="{ id, describedBy, invalid }" label="Nova senha" hint="Mínimo de 12 caracteres. Uma frase longa é melhor que uma senha curta e complicada." :error="pwErrs.next" :submitted="pwTried">
        <PasswordInput secret-label="nova senha" :id="id" v-model="pw.next" class="input input-bordered w-full" placeholder="Pelo menos 12 caracteres" maxlength="128" autocomplete="new-password" required :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
      <FormField v-slot="{ id, describedBy, invalid }" label="Confirmar nova senha" :error="pwErrs.confirm" :submitted="pwTried">
        <PasswordInput secret-label="confirmação de senha" :id="id" v-model="pw.confirm" class="input input-bordered w-full" placeholder="Repita a nova senha" maxlength="128" autocomplete="new-password" required :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
    </FormModal>
  </div>
</template>
