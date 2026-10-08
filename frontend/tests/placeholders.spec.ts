import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

// Regra Impeccable: todo campo de texto traz um placeholder de exemplo
// (o rótulo continua sendo o <label>; o placeholder só mostra o formato esperado).
const NO_PLACEHOLDER_TYPES = ['checkbox', 'radio', 'hidden', 'file', 'range', 'color', 'submit', 'button']

function vueFiles(dir: string): string[] {
  return readdirSync(dir).flatMap((f) => {
    const p = join(dir, f)
    return statSync(p).isDirectory() ? vueFiles(p) : p.endsWith('.vue') ? [p] : []
  })
}

describe('placeholders nos campos', () => {
  const offenders: string[] = []
  for (const file of vueFiles('src')) {
    const src = readFileSync(file, 'utf8')
    for (const m of src.matchAll(/<(input|textarea)\b[^>]*?\/?>/gs)) {
      const tag = m[0]
      const type = /\btype="([^"]+)"/.exec(tag)?.[1]
      if (type && NO_PLACEHOLDER_TYPES.includes(type)) continue
      if (!/\b:?placeholder=/.test(tag)) offenders.push(`${file}: ${tag.replace(/\s+/g, ' ').slice(0, 80)}`)
    }
  }
  it('nenhum input/textarea de texto fica sem placeholder', () => {
    expect(offenders).toEqual([])
  })
})
