import { afterEach, describe, expect, it, vi } from 'vitest'
import { api, setUnauthorizedHandler } from '../src/api/client'

afterEach(() => { vi.unstubAllGlobals(); vi.useRealTimers(); setUnauthorizedHandler(() => {}) })
function mockFetch(fn: typeof fetch) {
  vi.stubGlobal('document', { cookie: 'homealias_csrf=abc%20123' })
  vi.stubGlobal('fetch', fn)
}
describe('API resilience', () => {
  it('sends CSRF and credentials on writes', async () => {
    const fetcher = vi.fn(async () => new Response(null, { status: 204 }))
    mockFetch(fetcher)
    await api('/api/test', { method: 'POST', body: '{}' })
    const init = fetcher.mock.calls[0] as unknown as [string, RequestInit]
    expect(new Headers(init[1].headers).get('X-CSRF-Token')).toBe('abc 123')
    expect(init[1].credentials).toBe('include')
  })
  it('does not leak HTML errors or internal server details', async () => {
    for (const status of [400, 409, 500, 502, 503, 504]) {
      mockFetch(vi.fn(async () => new Response('<html>secret stack</html>', { status })))
      await expect(api('/api/test')).rejects.not.toThrow('secret')
    }
  })
  it('retains plain validation messages', async () => {
    mockFetch(vi.fn(async () => new Response('Nome é obrigatório.', { status: 400 })))
    await expect(api('/api/test')).rejects.toThrow('Nome é obrigatório.')
  })
  it('explains conflicts without exposing structured proxy responses', async () => {
    mockFetch(vi.fn(async () => new Response('O novo hostname já possui registros na Cloudflare.', { status: 409 })))
    await expect(api('/api/test')).rejects.toThrow('O novo hostname já possui registros na Cloudflare.')
    mockFetch(vi.fn(async () => new Response('{"internal":"secret"}', { status: 409 })))
    await expect(api('/api/test')).rejects.not.toThrow('secret')
  })
  it('redirects expired sessions but respects public login requests', async () => {
    const redirect = vi.fn(); setUnauthorizedHandler(redirect)
    mockFetch(vi.fn(async () => new Response('', { status: 401 })))
    await expect(api('/api/test', { skipAuthRedirect: true })).rejects.toThrow('sessão expirou')
    expect(redirect).not.toHaveBeenCalled()
    await expect(api('/api/test')).rejects.toThrow('sessão expirou')
    expect(redirect).toHaveBeenCalledOnce()
  })
  it('recovers from invalid successful JSON', async () => {
    mockFetch(vi.fn(async () => new Response('<html>proxy</html>')))
    await expect(api('/api/test')).rejects.toThrow('resposta inválida')
  })
  it('bounds stalled requests and gives safe retry guidance', async () => {
    vi.useFakeTimers()
    mockFetch(vi.fn((_url, init) => new Promise((_resolve, reject) => {
      init?.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')))
    })))
    const result = expect(api('/api/test', { timeoutMs: 100 })).rejects.toThrow('Verifique se a operação foi concluída')
    await vi.advanceTimersByTimeAsync(100)
    await result
  })
  it('supports caller cancellation', async () => {
    const controller = new AbortController()
    mockFetch(vi.fn((_url, init) => new Promise((_resolve, reject) => {
      init?.signal?.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')))
    })))
    const result = expect(api('/api/test', { signal: controller.signal })).rejects.toThrow('cancelada')
    controller.abort(); await result
  })
})
