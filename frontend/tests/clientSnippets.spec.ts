import { describe, expect, it } from 'vitest'
import * as s from '../src/composables/clientSnippets'

const c = { origin: 'https://ddns.exemplo.com', hostname: 'casa.exemplo.com', token: 'tok_ABC-123' }

describe('clientSnippets', () => {
  it('preenche os placeholders do script', () => {
    const out = s.configureScript('U=__HOMEALIAS_URL__ T=__HOMEALIAS_TOKEN__ H=__HOMEALIAS_HOSTNAME__ U2=__HOMEALIAS_URL__', c)
    expect(out).toBe('U=https://ddns.exemplo.com T=tok_ABC-123 H=casa.exemplo.com U2=https://ddns.exemplo.com')
    expect(out).not.toContain('__HOMEALIAS')
  })

  it('docker usa imagem pública e não grava o token no comando do loop', () => {
    const run = s.dockerRun(c)
    expect(run).toContain('curlimages/curl')
    expect(run).toContain('-e HA_TOKEN=tok_ABC-123')
    expect(run).toContain('$HA_TOKEN')
    expect(run).not.toContain('homealias-agent')
    expect(s.dockerCompose(c)).toContain('$$HA_TOKEN')
  })

  it('campos do roteador: servidor sem esquema e senha = token', () => {
    const f = Object.fromEntries(s.routerFields(c).map((x) => [x.label, x.value]))
    expect(f['Servidor / Host de atualização']).toBe('ddns.exemplo.com')
    expect(f['Senha']).toBe(c.token)
    expect(f['Nome do host / Domínio']).toBe(c.hostname)
    expect(f['URL completa (se o roteador pedir)']).toBe('https://ddns.exemplo.com/nic/update?hostname=casa.exemplo.com')
  })

  it('cron e curl', () => {
    expect(s.cronLine()).toMatch(/^\*\/5 \* \* \* \* /)
    expect(s.curlCommand(c)).toContain('Authorization: Bearer tok_ABC-123')
  })
})
