export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

type ApiInit = RequestInit & { skipAuthRedirect?: boolean; timeoutMs?: number }
let onUnauthorized: () => void = () => {}
export function setUnauthorizedHandler(fn: () => void) { onUnauthorized = fn }

function csrfToken(): string {
  const match = document.cookie.match(/(?:^|; )homealias_csrf=([^;]*)/)
  try { return match ? decodeURIComponent(match[1]) : '' } catch { return '' }
}

const FRIENDLY: Record<number, string> = {
  400: 'Dados inválidos. Revise os campos e tente de novo.',
  401: 'Sua sessão expirou. Entre novamente.',
  403: 'Você não tem permissão para fazer isso.',
  404: 'Item não encontrado. Atualize a lista e tente novamente.',
  409: 'Os dados foram alterados ou já existem. Atualize a lista e revise os campos.',
  429: 'Muitas tentativas. Aguarde um minuto e tente de novo.',
  500: 'Erro no servidor. Tente novamente em instantes.',
  502: 'O servidor está indisponível. Tente novamente em instantes.',
  503: 'O servidor está indisponível. Tente novamente em instantes.',
  504: 'O servidor demorou para responder. Tente novamente em instantes.',
}

export async function api<T>(path: string, init: ApiInit = {}): Promise<T> {
  const { skipAuthRedirect, timeoutMs = 30_000, signal, ...rest } = init
  const headers = new Headers(rest.headers)
  if (rest.method && !['GET', 'HEAD'].includes(rest.method.toUpperCase())) {
    headers.set('X-CSRF-Token', csrfToken())
    if (!headers.has('Content-Type') && rest.body) headers.set('Content-Type', 'application/json')
  }
  const controller = new AbortController()
  const abort = () => controller.abort(signal?.reason)
  if (signal?.aborted) abort()
  else signal?.addEventListener('abort', abort, { once: true })
  let timedOut = false
  const timer = setTimeout(() => { timedOut = true; controller.abort() }, timeoutMs)
  try {
    const res = await fetch(path, { ...rest, headers, signal: controller.signal, credentials: 'include' })
    if (!res.ok) {
      if (res.status === 401 && !skipAuthRedirect) onUnauthorized()
      const body = (await res.text()).trim()
      // Only expose plain validation messages; proxy HTML and internals aren't useful UI errors.
      const validation = res.status === 400 && body && body !== 'bad request' && !body.startsWith('<')
      throw new ApiError(res.status, validation ? body : (FRIENDLY[res.status] ?? 'Não foi possível concluir a solicitação. Tente novamente.'))
    }
    if (res.status === 204) return undefined as T
    try {
      return await res.json() as T
    } catch (e) {
      if (controller.signal.aborted) throw e
      throw new ApiError(res.status, 'O servidor enviou uma resposta inválida. Tente novamente.')
    }
  } catch (e) {
    if (e instanceof ApiError) throw e
    if (timedOut) throw new ApiError(0, 'O servidor demorou para responder. Verifique se a operação foi concluída antes de tentar novamente.')
    if (signal?.aborted) throw new ApiError(0, 'Solicitação cancelada.')
    throw new ApiError(0, 'Não foi possível falar com o servidor. Verifique sua conexão e tente novamente.')
  } finally {
    clearTimeout(timer)
    signal?.removeEventListener('abort', abort)
  }
}

export function errorMessage(e: unknown, fallback = 'Não foi possível concluir a operação. Tente novamente.'): string {
  return e instanceof Error && e.message ? e.message : fallback
}
