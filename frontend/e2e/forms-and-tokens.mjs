import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
const scriptTemplates = Object.fromEntries(['sh', 'ps1'].map(kind => [kind, readFileSync(new URL(`../../backend/internal/clientfiles/update.${kind}`, import.meta.url), 'utf8')]))
const { chromium } = await import(process.env.HOMEALIAS_PLAYWRIGHT_MODULE || 'playwright')
const browser = await chromium.launch({ executablePath: process.env.E2E_CHROME_PATH || '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] })
let checks = 0
for (const width of (process.env.E2E_WIDTHS ? process.env.E2E_WIDTHS.split(',').map(Number) : [375, 1280])) {
 const context = await browser.newContext({ viewport: { width, height: 844 }, reducedMotion: 'reduce' })
 const page = await context.newPage()
 let scriptFailure = true; let calls = 0; let empty = false
 await page.route('**/api/**', route => {
  const path = new URL(route.request().url()).pathname
  if (!path.startsWith('/api/')) return route.continue()
  if (path === '/api/auth/me' && route.request().method() === 'GET') return route.fulfill({ json: { id: 'me', name: 'Ana', email: 'ana@exemplo.com', role: 'admin' } })
  if (route.request().method() !== 'GET') {
   calls++
   if (path === '/api/tokens') return route.fulfill({ json: { token: 'homealias_' + 'a'.repeat(90), token_prefix: 'homealias', instructions: {} } })
   return route.fulfill({ status: 500, body: 'internal failure' })
  }
  const data = empty ? [] : ({
   '/api/hosts': [{ id: 'h1', fqdn: 'casa.exemplo.com', status: 'online', enable_a: true, enable_aaaa: false }],
   '/api/connections': [{ id: 'c1', name: 'Cloudflare', zones: [{ id: 'z1', name: 'exemplo.com' }] }],
   '/api/users': [{ id: 'u1', name: 'Usuário', email: 'u@exemplo.com', role: 'user', status: 'active' }],
  }[path] ?? [])
  return route.fulfill({ json: data })
 })
 await page.route('**/client/update.*', route => {
  const kind = new URL(route.request().url()).pathname.endsWith('.ps1') ? 'ps1' : 'sh'
  return route.fulfill(scriptFailure ? { status: 503, body: 'offline' } : { contentType: 'text/plain; charset=utf-8', body: scriptTemplates[kind] })
 })
 const base = process.env.E2E_BASE_URL || 'http://127.0.0.1:5173'
 // Account profile and administrator edit / password reset.
 for (const [path, opener, submit, fill, invalid] of [
  ['/conta', 'Editar perfil', 'Salvar', 'E-mail', 'E-mail'],
  ['/admin/users', 'Editar', 'Salvar', 'Nome', 'Nome'],
  ['/admin/users', 'Senha', 'Redefinir senha', 'Nova senha', 'Nova senha'],
 ]) {
  await page.goto(base + path); await page.waitForLoadState('networkidle')
  await page.getByRole('button', { name: opener, exact: true }).first().click()
  const dialog = page.locator('dialog[open]')
  await dialog.getByLabel(fill, { exact: true }).fill('')
  const before = calls
  await dialog.getByRole('button', { name: submit, exact: true }).click()
  await page.waitForTimeout(50)
  assert.equal(calls, before)
  assert.equal(await dialog.getByLabel(invalid, { exact: true }).getAttribute('aria-invalid'), 'true')
  checks += 2
  await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
 }
 // Token selection, generation, scripts retry and all instruction tabs.
 await page.goto(base + '/tokens'); await page.waitForLoadState('networkidle')
 const dialog = page.locator('dialog[open]')
 await dialog.getByRole('button', { name: 'Gerar token', exact: true }).click()
 await page.waitForTimeout(50)
 assert.equal(await dialog.getByLabel('Nome do token', { exact: true }).getAttribute('aria-invalid'), 'true'); checks++
 await dialog.getByLabel('Nome do token', { exact: true }).fill('NAS')
 await dialog.getByRole('button', { name: 'Gerar token', exact: true }).click()
 await page.getByRole('button', { name: 'Tentar novamente', exact: true }).waitFor()
 scriptFailure = false
 await page.getByRole('button', { name: 'Tentar novamente', exact: true }).click()
 await page.getByRole('button', { name: 'Baixar', exact: true }).first().waitFor()
 assert.equal(await page.getByRole('button', { name: 'Tentar novamente', exact: true }).count(), 0); checks++
 for (const tab of ['Script Linux', 'Docker', 'Roteador (DDNS)', 'Windows', 'cURL']) {
  await page.getByRole('tab', { name: tab, exact: true }).click()
  await page.waitForLoadState('networkidle')
  if (tab === 'Windows') {
   await page.locator('#panel-windows').getByRole('button', { name: 'Baixar', exact: true }).waitFor()
   const script = await page.locator('#panel-windows pre').first().textContent()
   assert.ok(script.includes('param(') && script.includes('casa.exemplo.com'))
   assert.ok(!script.includes('__HOMEALIAS_'))
   assert.equal(await page.locator('#panel-windows').getByRole('button', { name: 'Tentar novamente', exact: true }).count(), 0)
   checks += 3
  }
  assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1), `Instructions overflow ${tab} @ ${width}`); checks++
 }
 await page.screenshot({ path: `/tmp/homealias-token-${width}.png`, fullPage: true })
 // Theme parity and 200% text scaling in the form layout.
 await page.getByRole('button', { name: 'Menu de Ana', exact: true }).click()
 await page.getByRole('menuitem', { name: 'Tema escuro', exact: true }).click()
 await page.goto(base + '/conta'); await page.waitForLoadState('networkidle')
 await page.getByRole('button', { name: 'Alterar senha', exact: true }).first().click()
 await page.evaluate(() => { document.documentElement.style.fontSize = '200%' })
 assert.ok(await page.locator('dialog[open]').evaluate(el => el.scrollWidth <= el.clientWidth + 1), '200% form scaling'); checks++
 const zoomFits = await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)
 if (!zoomFits) { console.log(await page.locator('body').evaluate(el => [...el.querySelectorAll('*')].map(e => ({ tag: e.tagName, cls: e.className?.toString(), right: e.getBoundingClientRect().right, text: e.textContent?.slice(0, 30) })).filter(e => e.right > innerWidth + 1).slice(0, 15))); await page.screenshot({ path: '/tmp/homealias-zoom-overflow.png', fullPage: true }) }
 assert.ok(zoomFits, '200% page scaling'); checks++
 await page.screenshot({ path: `/tmp/homealias-zoom-${width}.png`, fullPage: true })
 await page.evaluate(() => { document.documentElement.style.fontSize = '' })
 await page.locator('dialog[open]').getByRole('button', { name: 'Cancelar', exact: true }).click()
 // Empty responses should display empty states instead of exceptions.
 empty = true
 for (const path of ['/connections', '/hosts', '/tokens', '/history', '/alerts', '/admin/users', '/admin/audit']) {
  await page.goto(base + path); await page.waitForLoadState('networkidle')
  assert.ok(await page.locator('h1').count() > 0)
  assert.equal(await page.getByRole('button', { name: 'Tentar novamente', exact: true }).count(), 0); checks++
 }
 await context.close()
 console.log(`Detailed forms/tokens @ ${width} passed`)
}
await browser.close()
console.log(`${checks} additional checks passed`)
