import { computed, ref } from 'vue'
import { api, setUnauthorizedHandler } from '../api/client'

export type User = { id: string; email: string; name?: string; role: 'admin' | 'user' }

const user = ref<User | null>(null)
const ready = ref(false)
let loading: Promise<void> | null = null

async function load() {
  try {
    user.value = await api<User>('/api/auth/me', { skipAuthRedirect: true })
  } catch {
    user.value = null
  } finally {
    ready.value = true
  }
}

export function useAuth() {
  return {
    user,
    ready,
    isAdmin: computed(() => user.value?.role === 'admin'),
    /** Carrega a sessão uma vez; chamadas seguintes reaproveitam o resultado. */
    ensure() {
      if (!loading) loading = load()
      return loading
    },
    async login(email: string, password: string) {
      await api('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
        skipAuthRedirect: true,
      })
      loading = load()
      await loading
    },
    async logout() {
      await api('/api/auth/logout', { method: 'POST', skipAuthRedirect: true })
      user.value = null
      loading = null
    },
    /** Recarrega o usuário (após editar o perfil). */
    async refresh() {
      loading = load()
      await loading
    },
    clear() {
      user.value = null
    },
  }
}

export function installAuthRedirect(redirect: () => void) {
  setUnauthorizedHandler(() => {
    user.value = null
    redirect()
  })
}
