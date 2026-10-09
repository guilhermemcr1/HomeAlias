import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { actionLabel, auditSummary, resultLabel, resourceLabel, updateError } from '../src/composables/uiCopy'
import { password } from '../src/composables/validation'

describe('user-facing language', () => {
  it('translates backend update outcomes without changing their codes', () => {
    expect(resultLabel('updated')).toBe('IP atualizado')
    expect(resultLabel('unchanged')).toBe('IP mantido')
    expect(resultLabel('rejected')).toBe('Atualização recusada')
    expect(resultLabel('unknown_code')).toBe('Resultado não identificado')
    expect(updateError('type not enabled')).toContain('IPv4 ou IPv6')
  })

  it('covers the actions actually emitted by the backend audit log', () => {
    const paths = ['account', 'auth', 'connections', 'hosts', 'tokens', 'users']
    for (const path of paths) {
      const source = readFileSync(new URL(`../../backend/internal/http/panel/${path}.go`, import.meta.url), 'utf8')
      for (const match of source.matchAll(/Action: "([^"]+)"/g)) expect(actionLabel(match[1])).not.toBe('Outra ação')
    }
    expect(resourceLabel('ddns_token')).toBe('Token DDNS')
    expect(auditSummary('host casa.example.com')).toBe('Endereço: casa.example.com')
    expect(auditSummary('password reset by admin')).toContain('administrador')
  })

  it('keeps the common-password list aligned with server validation', () => {
    const source = readFileSync(new URL('../../backend/internal/validate/validate.go', import.meta.url), 'utf8')
    const list = source.match(/for _, p := range \[\]string\{([\s\S]+?)\}/)?.[1]
    expect(list).toBeTruthy()
    for (const match of list!.matchAll(/"([^"]+)"/g)) expect(password(match[1])).not.toBe('')
  })
})
