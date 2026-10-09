// Espelha backend/internal/validate: o frontend avisa cedo, o backend decide.
// Cada validador devolve '' quando válido ou a mensagem (pt-BR) para o campo.

const DNS_LABEL = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/
const TELEGRAM_ID = /^(-?\d{5,20}|@[A-Za-z][A-Za-z0-9_]{4,31})$/
const EMAIL = /^[^\s@<>()[\]\\,;:"]+@[^\s@<>()[\]\\,;:"]+\.[^\s@<>()[\]\\,;:"]+$/

export const limits = { name: 80, email: 254, passwordMin: 12, passwordMax: 128, tokenMin: 20, tokenMax: 256 } as const

const hasControl = (s: string) => /[\u0000-\u001f\u007f-\u009f]/.test(s)
const runes = (s: string) => [...s].length

export function name(label: string, v: string): string {
  const s = v.trim()
  if (!s) return `${label} é obrigatório.`
  if (runes(s) > limits.name) return `${label} deve ter no máximo ${limits.name} caracteres.`
  if (hasControl(s)) return `${label} contém caracteres inválidos.`
  return ''
}

export function email(v: string): string {
  const s = v.trim().toLowerCase()
  if (!s) return 'E-mail é obrigatório.'
  if (s.length > limits.email || !EMAIL.test(s)) return 'Informe um e-mail válido, como nome@exemplo.com.'
  return ''
}

export function password(v: string): string {
  const n = runes(v)
  if (n < limits.passwordMin) return `A senha precisa ter pelo menos ${limits.passwordMin} caracteres.`
  if (n > limits.passwordMax) return `A senha deve ter no máximo ${limits.passwordMax} caracteres.`
  if (!v.trim()) return 'A senha não pode ser só espaços.'
  if (new Set([...v.toLowerCase()]).size < 5) return 'A senha é repetitiva demais. Use caracteres variados ou uma frase longa.'
  if (commonPasswords.has(v.toLowerCase().replace(/[^\p{L}\p{Nd}]/gu, ''))) return 'Essa senha é muito comum. Escolha outra, de preferência uma frase longa.'
  return ''
}

export function passwordMatch(a: string, b: string): string {
  if (!b) return 'Repita a senha para confirmar.'
  return a === b ? '' : 'As senhas são diferentes. Digite a mesma senha nos dois campos.'
}

// Espelha a lista de backend/internal/validate para avisar antes de enviar.
const commonPasswords = new Set([
  'changemenow', 'changemenow123', 'homealias', 'homealias123', 'homealiasadmin',
  'password1234', 'password12345', 'passwordpassword', 'senha1234567', 'senhasenha1234',
  '123456789012', '1234567890123', '123456123456', 'qwertyuiop12', 'qwertyuiopas',
  'qwerty123456', 'abcdefghijkl', 'abc123abc123', 'administrator', 'administrador',
  'admin1234567', 'adminadmin123', 'letmein12345', 'welcome12345', 'iloveyou1234',
  'mudar123456', 'trocar123456', 'senhaforte123', 'minhasenha123', 'brasil123456',
])

export function inviteToken(value: string): string {
  if (!value.trim()) return 'O código do convite é obrigatório.'
  return /^[A-Za-z0-9_-]{20,256}$/.test(value.trim()) ? '' : 'O código do convite está incompleto ou inválido. Abra o link enviado pelo administrador.'
}

/** Mesma tolerância do backend: remove "Bearer " e aspas coladas junto do token. */
export function normalizeApiToken(v: string): string {
  return v.trim().replace(/^bearer\s+/i, '').replace(/^["'`]+|["'`]+$/g, '').trim()
}

export function apiToken(v: string): string {
  const s = normalizeApiToken(v)
  if (!s) return 'O token de API é obrigatório.'
  if (/\s/.test(s)) return 'O token de API não pode conter espaços.'
  if (s.length < limits.tokenMin || s.length > limits.tokenMax) return 'O token de API tem tamanho inválido. Copie-o inteiro da Cloudflare.'
  if (!/^[A-Za-z0-9_-]+$/.test(s)) return 'O token de API tem caracteres inválidos. Cole só o token (letras, números, - e _), sem aspas.'
  return ''
}

export function hostName(v: string): string {
  const s = v.trim().toLowerCase()
  if (!s) return 'O subdomínio é obrigatório.'
  if (s === '@') return ''
  if (!s.split('.').every((l) => DNS_LABEL.test(l))) {
    return 'Subdomínio inválido: use letras, números e hífen, sem começar ou terminar com hífen.'
  }
  return ''
}

export function channelDestination(type: string, v: string): string {
  const s = v.trim()
  if (type === 'email') return email(s)
  if (type !== 'telegram') return 'Tipo de canal inválido.'
  if (!TELEGRAM_ID.test(s)) return 'Informe o número da conversa (Chat ID), como 123456789, ou o nome do canal, como @meucanal.'
  return ''
}

export function required(label: string, v: string): string {
  return v ? '' : `${label} é obrigatório.`
}

/** Primeiro erro não vazio de uma lista de resultados. */
export function firstError(...errs: string[]): string {
  return errs.find(Boolean) ?? ''
}

export function fqdn(host: string, zone: string): string {
  const full = host.trim() === '@' ? zone.trim() : `${host.trim()}.${zone.trim()}`
  return full.length > 253 ? 'O endereço completo pode ter no máximo 253 caracteres.' : ''
}
