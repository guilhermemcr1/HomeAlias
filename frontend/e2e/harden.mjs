// Run: HOMEALIAS_PLAYWRIGHT_MODULE=/path/to/playwright/index.mjs node e2e/harden.mjs
// Isolated API fixtures exercise UI failures without modifying real data.
import assert from 'node:assert/strict'
const { chromium } = await import(process.env.HOMEALIAS_PLAYWRIGHT_MODULE || 'playwright')
const browser = await chromium.launch({ executablePath: process.env.E2E_CHROME_PATH || '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] })
const base = process.env.E2E_BASE_URL || 'http://127.0.0.1:5173'
const long = 'Casa 🏠 العربية 中文 '.repeat(8)
const host = { id: 'h1', fqdn: `${'a'.repeat(63)}.${'b'.repeat(63)}.exemplo.com`, status: 'offline', last_ipv6: '2001:db8:abcd:1234:1234:1234:1234:1234', last_seen_v6: 'invalid', proxied: false }
const fixtures = {
 '/api/auth/me': { id: 'me', name: long, email: 'admin@exemplo.com', role: 'admin' },
 '/api/hosts': [host],
 '/api/connections': [{ id: 'c1', name: long, status: 'invalid', last_error: 'Erro'.repeat(150), zones: [{ id: 'z1', name: 'exemplo.com' }] }],
 '/api/alert-channels': [{ id: 'a1', name: long, type: 'email' }],
 '/api/alert-rules': [],
 '/api/users': [{ id: 'u1', name: long, email: 'usuario@exemplo.com', role: 'user', status: 'active' }],
 '/api/history': [{ id: 1, created_at: 'invalid', result: 'error', client_type: 'docker', changed: true, detected_ip: host.last_ipv6, error_reason: 'Falha'.repeat(100) }],
 '/api/audit': [{ id: 1, created_at: 'invalid', action: 'user.update', resource_type: 'user', summary: long, ip: host.last_ipv6 }],
}
let checks = 0
async function check(condition, message) { assert.ok(condition, message); checks++ }
for (const width of (process.env.E2E_WIDTHS ? process.env.E2E_WIDTHS.split(',').map(Number) : [320, 375, 390, 768, 1280])) {
 const context = await browser.newContext({ viewport: { width, height: 844 }, reducedMotion: 'reduce' })
 const page = await context.newPage()
 const errors = []
 page.on('pageerror', e => errors.push(e.message))
 let failing = ''; let writes = 0
 await page.route('**/api/**', async route => {
  const path = new URL(route.request().url()).pathname
  if (!path.startsWith('/api/')) return route.continue()
  if (route.request().method() !== 'GET') { writes++; await new Promise(r => setTimeout(r, 300)); return route.fulfill({ status: 500, body: 'failure' }) }
  if (path === failing) return route.fulfill({ status: 503, body: '<html>unavailable</html>' })
  return route.fulfill({ json: fixtures[path] ?? [] })
 })
 const routes = ['/', '/connections', '/hosts', '/tokens', '/history', '/alerts', '/admin/users', '/admin/audit', '/conta', '/ajuda', '/unknown']
 for (const path of routes) {
  await page.goto(base + path); await page.waitForLoadState('networkidle')
  await check(errors.length === 0, errors.join(', '))
  const content = await page.locator('body').innerText()
  if (content.includes('undefined')) console.error(content)
  await check(!content.includes('undefined'), `Missing data has a readable fallback: ${path} @ ${width}`)
  if (path === '/tokens') {
    if (await page.locator('dialog[open]').count()) await page.getByRole('button', { name: 'Cancelar', exact: true }).click()
  }
  const fits = await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1)
  if (!fits) {
    console.log(await page.locator('body').evaluate(el => [...el.querySelectorAll('*')].map(e => ({ tag: e.tagName, cls: e.className?.toString(), right: e.getBoundingClientRect().right, text: e.textContent?.slice(0, 50) })).filter(e => e.right > innerWidth + 1).slice(0, 20)))
    await page.screenshot({ path: '/tmp/homealias-overflow.png', fullPage: true })
  }
  await check(fits, `Page overflow: ${path} @ ${width}`)
 }
 // Mobile navigation must close on route change.
 if (width < 1280) {
  await page.getByRole('button', { name: 'Abrir menu', exact: true }).click()
  await page.locator('#mobile-nav').getByRole('link', { name: 'Hosts', exact: true }).click()
  await page.locator('#mobile-nav').waitFor({ state: 'hidden' })
  await check(await page.locator('#mobile-nav').count() === 0, 'Mobile menu closes')
 }
 const forms = [
  ['/connections', 'Nova conexão', 'Validar e salvar', 'Nome', 'name'],
  ['/hosts', 'Novo host', 'Criar host', 'Subdomínio', 'name'],
  ['/alerts', 'Novo canal', 'Cadastrar canal', 'Nome', 'name'],
  ['/admin/users', 'Novo usuário', 'Criar usuário', 'E-mail', 'email'],
  ['/conta', 'Alterar senha', 'Alterar senha', 'Senha atual', 'password'],
 ]
 for (const [path, open, submit, first] of forms) {
  await page.goto(base + path); await page.waitForLoadState('networkidle')
  await page.getByRole('button', { name: open, exact: true }).first().click()
  const dialog = page.locator('dialog[open]')
  const before = writes
  await dialog.getByRole('button', { name: submit, exact: true }).click()
  await page.waitForTimeout(50)
  await check(writes === before, `Invalid form didn't submit: ${path}`)
  await check(await dialog.getByLabel(first, { exact: true }).getAttribute('aria-invalid') === 'true', `Inline validation: ${path}`)
  await check(await dialog.getByLabel(first, { exact: true }).evaluate(el => el === document.activeElement), `Focus recovery: ${path}`)
  await check(await dialog.evaluate(el => el.scrollWidth <= el.clientWidth + 1), `Dialog overflow: ${path} @ ${width}`)
  if (path === '/connections') {
   await dialog.getByLabel('Nome', { exact: true }).fill('Conexão teste')
   await dialog.getByLabel('Token de API da Cloudflare').fill('a'.repeat(40))
   await dialog.locator('form').evaluate(form => { form.requestSubmit(); form.requestSubmit(); form.requestSubmit() })
   await page.waitForTimeout(500)
   await check(writes === before + 1, 'Duplicate submission blocked')
   await check(await dialog.getByLabel('Nome', { exact: true }).inputValue() === 'Conexão teste', 'Input retained after server error')
   await check(await dialog.getByRole('alert').count() > 0, 'Server error visible')
  }
  await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
 }
 // Every data page displays persistent errors and recovers through retry.
 for (const [path, api] of [['/', '/api/hosts'], ['/connections', '/api/connections'], ['/hosts', '/api/hosts'], ['/tokens', '/api/hosts'], ['/history', '/api/history'], ['/alerts', '/api/alert-channels'], ['/admin/users', '/api/users'], ['/admin/audit', '/api/audit']]) {
  failing = api
  await page.goto(base + path); await page.waitForLoadState('networkidle')
  await check(await page.getByRole('button', { name: 'Tentar novamente', exact: true }).count() === 1, `Retry available: ${path}`)
  failing = ''
  await page.getByRole('button', { name: 'Tentar novamente', exact: true }).click()
  await page.waitForLoadState('networkidle')
  await check(await page.getByRole('button', { name: 'Tentar novamente', exact: true }).count() === 0, `Retry recovers: ${path}`)
 }
 await page.goto(base + '/hosts'); await page.waitForLoadState('networkidle')
 await page.screenshot({ path: `/tmp/homealias-harden-${width}.png`, fullPage: true })
 await check(errors.length === 0, `No browser errors @ ${width}: ${errors.join(', ')}`)
 console.log(`Viewport ${width} passed`)
 await context.close()
}
// Public forms and offline login recovery.
const context = await browser.newContext({ viewport: { width: 375, height: 812 } })
const page = await context.newPage(); let calls = 0
await page.route('**/api/**', route => {
 if (!new URL(route.request().url()).pathname.startsWith('/api/')) return route.continue()
 if (route.request().method() === 'GET') return route.fulfill({ status: 401, body: 'unauthorized' })
 calls++; return route.abort('internetdisconnected')
})
await page.goto(base + '/login'); await page.waitForLoadState('networkidle')
await page.getByRole('button', { name: 'Entrar', exact: true }).click()
await check(calls === 0, 'Empty login rejected locally')
await page.getByLabel('E-mail', { exact: true }).fill('inválido')
await page.getByLabel('Senha', { exact: true }).fill('12345678')
await page.getByRole('button', { name: 'Entrar', exact: true }).click()
await check(calls === 0, 'Invalid email rejected locally')
await page.getByLabel('E-mail', { exact: true }).fill('ana@exemplo.com')
await page.getByRole('button', { name: 'Entrar', exact: true }).click()
await page.getByRole('alert').filter({ hasText: 'Verifique sua conexão' }).waitFor()
await check(await page.getByLabel('Senha', { exact: true }).inputValue() === '12345678', 'Password retained after network error')
await page.goto(base + '/convite'); await page.waitForLoadState('networkidle')
await page.getByRole('button', { name: 'Ativar conta', exact: true }).click()
await check(await page.getByLabel('Código do convite').getAttribute('aria-invalid') === 'true', 'Invite validation')
await context.close(); await browser.close()
console.log(`${checks} browser checks passed.`)
