import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createSSRApp, reactive, ref, type Ref } from 'vue'
import { renderToString } from '@vue/server-renderer'

const state = vi.hoisted(() => ({} as {
  auth: { ready: Ref<boolean>; user: Ref<{ id: string; role: string } | null> }
  route: { path: string; meta: { public?: boolean }; matched: object[] }
}))
vi.mock('../src/composables/useAuth', () => ({ useAuth: () => state.auth }))
vi.mock('vue-router', () => ({
  useRoute: () => state.route,
  RouterView: { template: '<div>Public page</div>' },
}))
vi.mock('../src/components/AppShell.vue', () => ({
  default: { template: '<nav aria-label="Principal">Private panel</nav>' },
}))
vi.mock('../src/components/AuthLayout.vue', () => ({
  default: { template: '<main><slot /></main>' },
}))
vi.mock('../src/components/ToastHost.vue', () => ({ default: { template: '<div />' } }))

import App from '../src/App.vue'

const render = () => renderToString(createSSRApp(App))

beforeEach(() => {
  state.auth = { ready: ref(false), user: ref(null) }
  state.route = reactive({ path: '/', meta: {}, matched: [] })
})

describe('authentication layout gate', () => {
  it('hides the navbar while the initial session request is pending', async () => {
    const html = await render()
    expect(html).toContain('Carregando…')
    expect(html).not.toContain('Private panel')
  })

  it('waits for the initial route even after authentication resolves', async () => {
    state.auth.ready.value = true
    state.auth.user.value = { id: 'user', role: 'user' }
    expect(await render()).not.toContain('Private panel')
    state.route.matched = [{}]
    expect(await render()).toContain('Private panel')
  })

  it('never shows the panel between an anonymous response and the login redirect', async () => {
    state.auth.ready.value = true
    state.route.matched = [{}]
    expect(await render()).not.toContain('Private panel')
    state.route.path = '/login'
    state.route.meta = { public: true }
    const html = await render()
    expect(html).toContain('Public page')
    expect(html).not.toContain('Private panel')
    expect(html).not.toContain('Carregando…')
  })

  it('removes the panel immediately when a session expires before redirect finishes', async () => {
    state.auth.ready.value = true
    state.auth.user.value = { id: 'user', role: 'user' }
    state.route.matched = [{}]
    expect(await render()).toContain('Private panel')
    state.auth.user.value = null
    expect(await render()).not.toContain('Private panel')
  })
})
