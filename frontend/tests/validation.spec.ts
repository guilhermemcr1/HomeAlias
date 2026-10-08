import { describe, expect, it } from 'vitest'
import * as v from '../src/composables/validation'

describe('validation (espelho do backend)', () => {
  it('name: obrigatório, limite de 80 e sem controle', () => {
    expect(v.name('Nome', '  ')).toMatch(/obrigatório/)
    expect(v.name('Nome', 'é'.repeat(81))).toMatch(/80/)
    expect(v.name('Nome', 'a\u0000b')).toMatch(/inválidos/)
    expect(v.name('Nome', ' Conta 🏠 ')).toBe('')
  })

  it('email', () => {
    expect(v.email('Ana@Exemplo.COM')).toBe('')
    for (const bad of ['', 'abc', 'a@b', 'Ana <a@b.com>', 'a b@c.com']) expect(v.email(bad)).not.toBe('')
  })

  it('password: 12 a 128', () => {
    expect(v.password('12345678901')).not.toBe('')
    expect(v.password('123456789012')).toBe('')
    expect(v.password('a'.repeat(129))).not.toBe('')
  })

  it('apiToken: sem espaços e tamanho plausível', () => {
    expect(v.apiToken('a'.repeat(40))).toBe('')
    expect(v.apiToken('curto')).not.toBe('')
    expect(v.apiToken(`abc def ${'a'.repeat(40)}`)).toMatch(/espaços/)
  })

  it('hostName: rótulos DNS ou @', () => {
    for (const ok of ['casa', 'Casa', 'a.b', '@', 'nas-01']) expect(v.hostName(ok)).toBe('')
    for (const bad of ['', '-a', 'a-', 'a..b', 'a b', 'a_b', 'café', 'a'.repeat(64), 'a/b']) expect(v.hostName(bad)).not.toBe('')
  })

  it('channelDestination: e-mail ou Chat ID', () => {
    expect(v.channelDestination('email', 'a@b.com')).toBe('')
    expect(v.channelDestination('email', 'x')).not.toBe('')
    for (const ok of ['123456789', '-1001234567890', '@meucanal']) expect(v.channelDestination('telegram', ok)).toBe('')
    expect(v.channelDestination('telegram', 'abc')).not.toBe('')
  })
})

it('handles unicode limits, full DNS lengths, blank selections and invalid channel types', () => {
  expect(v.name('Nome', '🏠'.repeat(80))).toBe('')
  expect(v.name('Nome', '🏠'.repeat(81))).not.toBe('')
  expect(v.required('Host', '')).not.toBe('')
  expect(v.channelDestination('invalid', '123456789')).not.toBe('')
  expect(v.fqdn('@', 'exemplo.com')).toBe('')
  expect(v.fqdn('a'.repeat(63), `${'b'.repeat(63)}.${'c'.repeat(63)}.${'d'.repeat(63)}`)).not.toBe('')
})
