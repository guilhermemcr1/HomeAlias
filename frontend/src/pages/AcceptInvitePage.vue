<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, errorMessage } from '../api/client'
import FormField from '../components/FormField.vue'
import { useToast } from '../composables/useToast'
import { focusFirstInvalid } from '../composables/formFocus'
import * as v from '../composables/validation'

const route = useRoute()
const router = useRouter()
const toast = useToast()
const token = ref(typeof route.query.token === 'string' ? route.query.token : '')
const name = ref('')
const password = ref('')
const formElement = ref<HTMLFormElement | null>(null)
const busy = ref(false)
const error = ref('')
const attempted = ref(false)
const errs = computed(() => ({
  token: v.required('Código do convite', token.value.trim()),
  name: v.name('Nome', name.value),
  password: v.password(password.value),
}))

async function accept() {
  if (busy.value) return
  error.value = ''
  attempted.value = true
  if (Object.values(errs.value).some(Boolean)) {
    await focusFirstInvalid(formElement.value)
    return
  }
  busy.value = true
  try {
    await api('/api/auth/invite/accept', {
      method: 'POST',
      body: JSON.stringify({ token: token.value.trim(), password: password.value, name: name.value.trim() }),
      skipAuthRedirect: true,
    })
    toast.success('Conta ativada. Entre com a senha que você definiu.')
    router.push('/login')
  } catch (e) {
    error.value = errorMessage(e, 'Convite inválido ou expirado.')
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="page-title">Ativar conta</h1>
    <p class="mt-1.5 mb-8 text-base-content/70">Você foi convidado. Defina seu nome e uma senha para entrar.</p>
    <form ref="formElement" :aria-busy="busy" class="flex flex-col gap-5" novalidate @submit.prevent="accept">
      <FormField v-slot="{ id, describedBy, invalid }" label="Código do convite" :error="attempted ? errs.token : ''">
        <input :id="id" v-model="token" class="input input-bordered w-full font-data" placeholder="Cole o código do link de convite" maxlength="256" required autocomplete="off" spellcheck="false" :disabled="busy" :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
      <FormField v-slot="{ id, describedBy, invalid }" label="Nome" :error="attempted ? errs.name : ''">
        <input :id="id" v-model="name" class="input input-bordered w-full" placeholder="Ex.: Ana Souza" maxlength="80" autocomplete="name" required :disabled="busy" :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
      <FormField v-slot="{ id, describedBy, invalid }" label="Senha" hint="Mínimo de 12 caracteres. Uma frase longa é melhor que uma senha curta e complicada." :error="attempted ? errs.password : ''">
        <input :id="id" v-model="password" type="password" class="input input-bordered w-full" placeholder="Pelo menos 12 caracteres" maxlength="128" autocomplete="new-password" required :disabled="busy" :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
      <p v-if="error" class="rounded-btn border border-error/50 bg-error/10 px-3 py-2.5 text-sm" role="alert">{{ error }}</p>
      <button class="btn btn-primary btn-block h-12" type="submit" :disabled="busy">
        <span v-if="busy" class="loading loading-spinner loading-sm" />
        Ativar conta
      </button>
    </form>
    <p class="mt-8 text-sm"><RouterLink to="/login" class="link">Já tenho conta</RouterLink></p>
  </div>
</template>
