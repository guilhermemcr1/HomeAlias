<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import {
  Bell,
  Cable,
  ChevronDown,
  Globe,
  History,
  KeyRound,
  LayoutDashboard,
  LifeBuoy,
  LogOut,
  Menu,
  Moon,
  ScrollText,
  ShieldCheck,
  Sun,
  UserRound,
  Users,
  X,
} from 'lucide-vue-next'
import { errorMessage } from '../api/client'
import { useToast } from '../composables/useToast'
import BrandMark from './BrandMark.vue'
import AppFooter from './AppFooter.vue'
import { useAuth } from '../composables/useAuth'
import { useTheme } from '../composables/useTheme'

const route = useRoute()
const router = useRouter()
const { user, isAdmin, logout } = useAuth()
const { theme, toggle } = useTheme()
const open = ref(false)

const main = [
  { to: '/', label: 'Painel', icon: LayoutDashboard },
  { to: '/connections', label: 'Conexões', icon: Cable },
  { to: '/hosts', label: 'Hosts', icon: Globe },
  { to: '/tokens', label: 'Tokens', icon: KeyRound },
  { to: '/history', label: 'Histórico', icon: History },
  { to: '/alerts', label: 'Alertas', icon: Bell },
]
const admin = [
  { to: '/admin/users', label: 'Usuários', icon: Users },
  { to: '/admin/audit', label: 'Auditoria', icon: ScrollText },
]

type MenuName = 'user' | 'admin'
const menu = ref<MenuName | null>(null)
const menuRoots = ref<Partial<Record<MenuName, HTMLElement | null>>>({})
const triggers = ref<Partial<Record<MenuName, HTMLButtonElement | null>>>({})

watch(
  () => route.fullPath,
  () => {
    open.value = false
    menu.value = null
  },
)

function toggleMenu(name: MenuName) {
  menu.value = menu.value === name ? null : name
}
function closeMenu(returnFocus = false) {
  const was = menu.value
  menu.value = null
  if (returnFocus && was) triggers.value[was]?.focus()
}
function onDocPointer(e: PointerEvent) {
  const root = menu.value ? menuRoots.value[menu.value] : null
  if (menu.value && !root?.contains(e.target as Node)) closeMenu()
}
function onDocKey(e: KeyboardEvent) {
  if (e.key !== 'Escape') return
  if (menu.value) closeMenu(true)
  else open.value = false
}
onMounted(() => {
  document.addEventListener('pointerdown', onDocPointer)
  document.addEventListener('keydown', onDocKey)
})
onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocPointer)
  document.removeEventListener('keydown', onDocKey)
})

function isActive(to: string) {
  return to === '/' ? route.path === '/' : route.path.startsWith(to)
}
const adminActive = () => admin.some((i) => isActive(i.to))

const signingOut = ref(false)
const toast = useToast()
async function signOut() {
  if (signingOut.value) return
  signingOut.value = true
  menu.value = null
  open.value = false
  try {
    await logout()
    await router.push('/login')
  } catch (e) {
    toast.error(errorMessage(e, 'Não foi possível sair. Tente novamente.'))
  } finally {
    signingOut.value = false
  }
}

const linkBase = 'flex items-center gap-2.5 rounded-btn px-3 h-10 text-sm font-medium transition-colors whitespace-nowrap'
// Barra horizontal compacta (ícones de 15px) para caber em uma linha com fonte monoespaçada.
const barLink = 'flex items-center gap-1.5 rounded-btn px-2.5 h-9 text-[13px] font-medium transition-colors whitespace-nowrap'
const linkOn = 'bg-primary text-primary-content'
const linkOff = 'hover:bg-base-200'
</script>

<template>
  <div class="flex min-h-dvh flex-col">
    <header class="app-header sticky top-0 z-40 border-b border-base-300 bg-base-100/95 backdrop-blur">
      <div class="mx-auto flex min-h-14 w-full max-w-7xl items-center gap-3 px-4 py-1 sm:px-8">
        <RouterLink to="/" class="flex shrink-0 items-center gap-2.5 xl:mr-2 xl:border-r xl:border-base-300 xl:pr-4" aria-label="HomeAlias, ir para o painel">
          <BrandMark :size="24" />
          <span class="font-display text-base font-bold tracking-tight">HomeAlias</span>
        </RouterLink>

        <nav class="desktop-navigation hidden min-w-0 flex-1 flex-wrap items-center gap-0.5 xl:flex" aria-label="Principal">
          <RouterLink
            v-for="i in main"
            :key="i.to"
            :to="i.to"
            :class="[barLink, isActive(i.to) ? linkOn : linkOff]"
            :aria-current="isActive(i.to) ? 'page' : undefined"
          >
            <component :is="i.icon" :size="15" class="shrink-0" aria-hidden="true" />{{ i.label }}
          </RouterLink>
          <div v-if="isAdmin" :ref="(el) => (menuRoots.admin = el as HTMLElement)" class="relative">
            <button
              :ref="(el) => (triggers.admin = el as HTMLButtonElement)"
              type="button"
              :class="[barLink, adminActive() ? linkOn : linkOff]"
              aria-haspopup="menu"
              :aria-expanded="menu === 'admin'"
              aria-controls="admin-menu"
              @click="toggleMenu('admin')"
            >
              <ShieldCheck :size="15" class="shrink-0" aria-hidden="true" />Administração<ChevronDown :size="12" class="opacity-70" aria-hidden="true" />
            </button>
            <Transition name="pop">
              <div v-if="menu === 'admin'" id="admin-menu" role="menu" class="absolute left-0 top-full mt-2 w-52 rounded-box border border-base-300 bg-base-100 p-1.5 shadow-lg">
                <RouterLink v-for="i in admin" :key="i.to" :to="i.to" role="menuitem" :class="[linkBase, 'h-11', isActive(i.to) ? linkOn : linkOff]">
                  <component :is="i.icon" :size="18" aria-hidden="true" />{{ i.label }}
                </RouterLink>
              </div>
            </Transition>
          </div>
        </nav>

        <div class="ml-auto flex shrink-0 items-center gap-1.5 xl:ml-0">
                    <div :ref="(el) => (menuRoots.user = el as HTMLElement)" class="relative">
            <button
              :ref="(el) => (triggers.user = el as HTMLButtonElement)"
              type="button"
              class="flex items-center gap-2 rounded-btn py-1.5 pl-1.5 pr-2 transition-colors hover:bg-base-200"
              :class="{ 'bg-base-200': menu === 'user' }"
              aria-haspopup="menu"
              :aria-expanded="menu === 'user'"
              aria-controls="user-menu"
              :aria-label="`Menu de ${user?.name || user?.email || 'usuário'}`"
              @click="toggleMenu('user')"
            >
              <span class="grid size-8 shrink-0 place-items-center rounded-full bg-primary text-sm font-bold text-primary-content" aria-hidden="true">
                {{ (user?.name || user?.email || '?').trim().charAt(0).toUpperCase() }}
              </span>
              <ChevronDown :size="14" class="hidden text-base-content/60 sm:block" aria-hidden="true" />
            </button>
            <Transition name="pop">
              <div v-if="menu === 'user'" id="user-menu" role="menu" aria-label="Menu do usuário" class="absolute right-0 top-full mt-2 w-64 max-w-[calc(100vw-2rem)] rounded-box border border-base-300 bg-base-100 p-1.5 shadow-lg">
                <div class="border-b border-base-300 px-3 pb-2.5 pt-2">
                  <p class="truncate text-sm font-semibold" :title="user?.name || user?.email">{{ user?.name || user?.email }}</p>
                  <p class="truncate text-xs text-base-content/65" :title="user?.email">{{ isAdmin ? 'Administrador' : 'Usuário' }} · {{ user?.email }}</p>
                </div>
                <div class="pt-1.5">
                  <RouterLink to="/conta" role="menuitem" :class="[linkBase, 'h-11', linkOff]"><UserRound :size="18" aria-hidden="true" />Minha conta</RouterLink>
                  <RouterLink :to="{ path: '/conta', query: { senha: '1' } }" role="menuitem" :class="[linkBase, 'h-11', linkOff]"><KeyRound :size="18" aria-hidden="true" />Alterar senha</RouterLink>
                  <RouterLink to="/ajuda" role="menuitem" :class="[linkBase, 'h-11', linkOff]"><LifeBuoy :size="18" aria-hidden="true" />Ajuda</RouterLink>
                  <button type="button" role="menuitem" :class="[linkBase, 'h-11 w-full', linkOff]" @click="toggle">
                    <component :is="theme === 'homealias-dark' ? Sun : Moon" :size="18" aria-hidden="true" />{{ theme === 'homealias-dark' ? 'Tema claro' : 'Tema escuro' }}
                  </button>
                  <button type="button" role="menuitem" :class="[linkBase, 'h-11 w-full text-error', linkOff]" :disabled="signingOut" @click="signOut"><LogOut :size="18" aria-hidden="true" />Sair</button>
                </div>
              </div>
            </Transition>
          </div>

          <button
            type="button"
            class="btn btn-ghost btn-square mobile-menu-toggle xl:hidden"
            :aria-label="open ? 'Fechar menu' : 'Abrir menu'"
            :aria-expanded="open"
            aria-controls="mobile-nav"
            @click="open = !open"
          >
            <component :is="open ? X : Menu" :size="22" aria-hidden="true" />
          </button>
        </div>
      </div>

      <Transition name="pop">
        <nav v-if="open" id="mobile-nav" class="mobile-navigation max-h-[calc(100dvh-3.5rem)] overflow-y-auto border-t border-base-300 bg-base-100 px-4 pb-4 pt-2 xl:hidden" aria-label="Principal">
          <ul class="flex flex-col gap-1">
            <li v-for="i in main" :key="i.to">
              <RouterLink :to="i.to" :class="[linkBase, 'h-11', isActive(i.to) ? linkOn : linkOff]" :aria-current="isActive(i.to) ? 'page' : undefined">
                <component :is="i.icon" :size="20" aria-hidden="true" />{{ i.label }}
              </RouterLink>
            </li>
          </ul>
          <template v-if="isAdmin">
            <p class="mb-1 mt-4 px-3 text-xs font-semibold uppercase tracking-wider text-base-content/60">Administração</p>
            <ul class="flex flex-col gap-1">
              <li v-for="i in admin" :key="i.to">
                <RouterLink :to="i.to" :class="[linkBase, 'h-11', isActive(i.to) ? linkOn : linkOff]">
                  <component :is="i.icon" :size="20" aria-hidden="true" />{{ i.label }}
                </RouterLink>
              </li>
            </ul>
          </template>
          <div class="mt-3 border-t border-base-300 pt-3">
            <RouterLink to="/ajuda" :class="[linkBase, 'h-11', isActive('/ajuda') ? linkOn : linkOff]"><LifeBuoy :size="20" aria-hidden="true" />Ajuda</RouterLink>
          </div>
        </nav>
      </Transition>
    </header>

    <main id="conteudo" class="mx-auto w-full max-w-6xl flex-1 px-4 py-6 sm:px-8 sm:py-10">
      <RouterView v-slot="{ Component }">
        <Transition name="page" mode="out-in">
          <component :is="Component" :key="route.path" />
        </Transition>
      </RouterView>
    </main>
    <AppFooter />
  </div>
</template>
