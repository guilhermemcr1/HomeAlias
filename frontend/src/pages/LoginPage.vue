<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Eye, EyeOff } from 'lucide-vue-next'
import * as v from '../composables/validation'
import { focusFirstInvalid } from '../composables/formFocus'
import { ApiError } from '../api/client'
import FormField from '../components/FormField.vue'
import CanvasText from '../components/ui/CanvasText.vue'
import { useAuth } from '../composables/useAuth'

const email = ref('')
const password = ref('')
const show = ref(false)
const phase = ref<'idle' | 'checking' | 'success'>('idle')
const busy = computed(() => phase.value !== 'idle')
const error = ref('')
const attempted = ref(false)
const formElement = ref<HTMLFormElement | null>(null)
const errs = computed(() => ({ email: v.email(email.value), password: v.required('Senha', password.value) }))
const route = useRoute()
const router = useRouter()
const { login } = useAuth()

function destination(): string {
  const next = route.query.next
  return typeof next === 'string' && next.startsWith('/') && !next.startsWith('//') ? next : '/'
}

async function submit() {
  if (busy.value) return
  error.value = ''
  attempted.value = true
  if (Object.values(errs.value).some(Boolean)) {
    await focusFirstInvalid(formElement.value)
    return
  }
  phase.value = 'checking'
  try {
    await login(email.value.trim(), password.value)
    phase.value = 'success'
    // Dá um instante para ver a confirmação antes de trocar de página.
    await new Promise((r) => setTimeout(r, 800))
    await router.push(destination())
  } catch (e) {
    phase.value = 'idle'
    if (e instanceof ApiError && e.status === 401) error.value = 'E-mail ou senha incorretos.'
    else if (e instanceof ApiError && e.status === 429) error.value = 'Muitas tentativas. Aguarde um minuto e tente de novo.'
    else error.value = e instanceof Error ? e.message : 'Não foi possível entrar.'
    // Preserva os campos para corrigir credenciais ou repetir após falha de rede.
  }
}
</script>

<template>
  <div>
    <h1 class="mb-8 text-center text-3xl font-extrabold leading-tight tracking-tight">
      <span class="block text-base-content/70 text-lg font-semibold">Bem-vindo ao</span>
      <CanvasText
        text="HomeAlias"
        class="mt-1"
        :colors="['rgba(141,199,255,1)', 'rgba(141,199,255,0.85)', 'rgba(141,199,255,0.7)', 'rgba(141,199,255,0.55)', 'rgba(141,199,255,0.4)', 'rgba(141,199,255,0.25)', 'rgba(141,199,255,0.12)']"
        :line-gap="4"
        :animation-duration="12"
      />
    </h1>
    <form ref="formElement" :aria-busy="busy" class="flex flex-col gap-5" novalidate @submit.prevent="submit">
      <FormField v-slot="{ id, describedBy, invalid }" label="E-mail" :error="attempted ? errs.email : ''">
        <input :id="id" v-model="email" type="email" inputmode="email" autocapitalize="none" spellcheck="false" class="input input-bordered w-full" placeholder="voce@exemplo.com" maxlength="254" autocomplete="username" :aria-invalid="invalid" :aria-describedby="describedBy" :disabled="busy" autofocus required />
      </FormField>
      <FormField v-slot="{ id, describedBy, invalid }" label="Senha" :error="attempted ? errs.password : ''">
        <div class="relative">
          <input
            :id="id"
            v-model="password"
            :type="show ? 'text' : 'password'"
            class="input input-bordered w-full pr-12"
            placeholder="Sua senha"
            maxlength="128"
            autocomplete="current-password"
            :aria-invalid="invalid"
            :aria-describedby="describedBy"
            :disabled="busy"
            required
          />
          <button
            type="button"
            class="btn btn-ghost btn-sm btn-square absolute right-1.5 top-1/2 -translate-y-1/2"
            :aria-label="show ? 'Ocultar senha' : 'Mostrar senha'"
            :aria-pressed="show"
            @click="show = !show"
          >
            <component :is="show ? EyeOff : Eye" :size="18" />
          </button>
        </div>
      </FormField>
      <p v-if="error" class="rounded-btn border border-error/50 bg-error/10 px-3 py-2.5 text-sm" role="alert">{{ error }}</p>
      <button
        class="btn btn-block h-12"
        :class="phase === 'success' ? 'btn-success' : 'btn-primary'"
        type="submit"
        :disabled="busy"
        :aria-busy="phase === 'checking'"
      >
        <span v-if="phase === 'checking'" class="loading loading-spinner loading-sm" />
        <svg v-else-if="phase === 'success'" class="check-pop size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path class="check-draw" d="M5 12.5l4.5 4.5L19 7.5" />
        </svg>
        {{ phase === 'checking' ? 'Verificando…' : phase === 'success' ? 'Acesso confirmado' : 'Entrar' }}
      </button>
      <span class="sr-only" role="status">{{ phase === 'checking' ? 'Verificando credenciais' : phase === 'success' ? 'Login confirmado, redirecionando' : '' }}</span>
    </form>
  </div>
</template>
