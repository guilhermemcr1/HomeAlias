import assert from 'node:assert/strict'
const { chromium } = await import(process.env.HOMEALIAS_PLAYWRIGHT_MODULE || 'playwright')
const browser = await chromium.launch({ executablePath: process.env.E2E_CHROME_PATH || '/usr/bin/google-chrome', args: ['--no-sandbox'], headless: true })
const base = process.env.E2E_BASE_URL || 'http://127.0.0.1:5173'
let checks = 0
try {
 for (const width of [320, 375, 1280]) {
  const page = await browser.newPage({ viewport: { width, height: 900 }, reducedMotion: 'reduce' })
  const errors = [], writes = []
  let anonymous = true, holdOldCheck, oldCheckStarted = false
  page.on('pageerror', error => errors.push(error.message))
  const connection = { id: 'connection', name: 'Conta pessoal', zones: [{ id: 'zone', name: 'example.com' }], status: 'valid' }
  const host = { id: 'host', fqdn: 'old.example.com', zone_name: 'example.com', enable_a: true, enable_aaaa: false, ttl: 1, proxied: false, status: 'online' }
  await page.route('**/api/**', async route => {
   const request = route.request(), path = new URL(request.url()).pathname
   if (!path.startsWith('/api/')) return route.continue()
   if (path === '/api/auth/me') return route.fulfill({ status: anonymous ? 401 : 200, json: anonymous ? {} : { id: 'me', name: 'Ana', email: 'ana@example.com', role: 'admin' } })
   if (path === '/api/hosts/check') {
    const body = request.postDataJSON()
    if (body.name === 'old') {
     oldCheckStarted = true
     await new Promise(resolve => { holdOldCheck = resolve })
     try { await route.fulfill({ json: { fqdn: 'old.example.com', in_app: true, records: [] } }) } catch { /* The browser correctly cancelled the stale request. */ }
     return
    }
    return route.fulfill({ json: { fqdn: `${body.name}.example.com`, in_app: false, records: [] } })
   }
   if (request.method() !== 'GET') { writes.push({ path, body: request.postDataJSON() }); return route.fulfill({ json: {} }) }
   const data = {
    '/api/connections': [connection], '/api/hosts': [host], '/api/users': [{ id: 'other', name: 'Outro usuário', email: 'other@example.com', role: 'user', status: 'active' }],
    '/api/history': [{ id: 1, created_at: new Date().toISOString(), result: 'updated', client_type: 'docker', changed: true, detected_ip: '192.0.2.1' }],
    '/api/audit': [{ id: 1, created_at: new Date().toISOString(), action: 'token_create', resource_type: 'ddns_token', summary: 'token emitted', ip: '192.0.2.1' }],
   }
   return route.fulfill({ json: data[path] || [] })
  })
  async function go(path, publicPage = false) { anonymous = publicPage; await page.goto(base + path); await page.waitForLoadState('networkidle') }
  async function check(condition, message) { assert.ok(condition, message); checks++ }
  // Blur validates only the field being left, then typing a correction clears the error.
  await go('/login', true)
  const email = page.getByLabel('E-mail', { exact: true })
  await email.fill('errado')
  await page.getByLabel('Senha', { exact: true }).focus()
  await check(await email.getAttribute('aria-invalid') === 'true', 'email blur validation')
  await check(await page.getByLabel('Senha', { exact: true }).getAttribute('aria-invalid') === 'false', 'untouched password remains quiet')
  await email.fill('ana@example.com')
  await check(await email.getAttribute('aria-invalid') === 'false', 'corrected email clears error')
  // Hints remain available while errors are shown; modal reopening resets feedback and secret visibility.
  await go('/connections')
  await page.getByRole('button', { name: 'Nova conexão', exact: true }).first().click()
  let dialog = page.locator('dialog[open]')
  let token = dialog.getByLabel('Token de API da Cloudflare', { exact: true })
  await token.fill('curto')
  await dialog.getByLabel('Nome', { exact: true }).focus()
  await check(await token.getAttribute('aria-invalid') === 'true', 'invalid token validated on blur')
  const descriptions = await token.getAttribute('aria-describedby')
  await check(descriptions.split(' ').length === 2, 'error and hint both associated')
  for (const id of descriptions.split(' ')) await check(await dialog.locator(`[id="${id}"]`).isVisible(), 'description remains visible')
  await dialog.getByRole('button', { name: 'Mostrar token', exact: true }).click()
  await check(await token.getAttribute('type') === 'text', 'token can be checked')
  await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
  await page.getByRole('button', { name: 'Nova conexão', exact: true }).first().click()
  await check(await token.inputValue() === '' && await token.getAttribute('type') === 'password', 'reopened token is empty and masked')
  await check(await token.getAttribute('aria-invalid') === 'false', 'reopened form clears touched state')
  await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
  // Password rules match the backend and retain the explanatory hint.
  await go('/conta')
  await page.getByRole('button', { name: 'Alterar senha', exact: true }).first().click()
  dialog = page.locator('dialog[open]')
  const next = dialog.getByLabel('Nova senha', { exact: true })
  await dialog.getByLabel('Senha atual', { exact: true }).fill('senha-anterior')
  await next.fill('Password-1234!')
  await dialog.getByLabel('Confirmar nova senha', { exact: true }).focus()
  await check(await next.getAttribute('aria-invalid') === 'true', 'common password is rejected locally')
  await dialog.getByText('Essa senha é muito comum. Escolha outra, de preferência uma frase longa.', { exact: true }).waitFor()
  await next.fill('Meu jardim tem 3 luas!')
  await dialog.getByLabel('Confirmar nova senha', { exact: true }).fill('outra senha')
  const before = writes.length
  await dialog.getByRole('button', { name: 'Alterar senha', exact: true }).click()
  await check(writes.length === before, 'mismatched password is not submitted')
  await check(await dialog.getByLabel('Confirmar nova senha', { exact: true }).getAttribute('aria-invalid') === 'true', 'confirmation explains correction')
  await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
  await page.getByRole('button', { name: 'Alterar senha', exact: true }).first().click()
  await check(await next.inputValue() === '', 'cancelled password is cleared')
  await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
  // New invite confirmation prevents accidental mismatches and malformed codes.
  await go('/convite?token=' + 'a'.repeat(40), true)
  await page.getByLabel('Nome', { exact: true }).fill('Ana')
  await page.getByLabel('Senha', { exact: true }).fill('Meu jardim tem 3 luas!')
  await page.getByLabel('Confirmar senha', { exact: true }).fill('Diferente')
  const inviteBefore = writes.length
  await page.getByRole('button', { name: 'Ativar conta', exact: true }).click()
  await check(writes.length === inviteBefore, 'invite mismatch does not submit')
  await page.getByLabel('Confirmar senha', { exact: true }).fill('Meu jardim tem 3 luas!')
  await page.getByRole('button', { name: 'Ativar conta', exact: true }).click()
  await page.waitForURL('**/login')
  await check(writes.length === inviteBefore + 1 && writes.at(-1).path === '/api/auth/invite/accept', 'matching invite submits exactly once')
  // An old Cloudflare response cannot show a conflict for the new name.
  await go('/hosts')
  await page.getByRole('button', { name: 'Novo host', exact: true }).first().click()
  dialog = page.locator('dialog[open]')
  await dialog.getByLabel('Subdomínio', { exact: true }).fill('old')
  for (let attempts = 0; !oldCheckStarted && attempts < 200; attempts++) await page.waitForTimeout(25)
  assert.ok(oldCheckStarted, 'old name check started')
  await dialog.getByLabel('Subdomínio', { exact: true }).fill('new')
  await dialog.getByText('new.example.com está livre.', { exact: true }).waitFor()
  holdOldCheck()
  await page.waitForTimeout(100)
  await check(await dialog.getByText('new.example.com está livre.', { exact: true }).isVisible(), 'stale conflict is ignored')
  await check(await dialog.getByRole('button', { name: 'Criar host', exact: true }).isEnabled(), 'new valid name is not blocked')
  await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
  // Technical event codes have readable labels.
  await go('/history'); await page.locator('tbody').getByText('IP atualizado', { exact: true }).waitFor()
  await go('/admin/audit'); await page.getByText('Token gerado', { exact: true }).waitFor()
  await go('/alerts'); await page.getByText(/O envio automático por regras ainda não está disponível/).waitFor()
  await page.getByRole('button', { name: 'Novo canal', exact: true }).first().click()
  dialog = page.locator('dialog[open]')
  await dialog.getByLabel('Tipo', { exact: true }).selectOption('telegram')
  await dialog.getByLabel('Conversa ou canal do Telegram', { exact: true }).fill('inválido')
  await dialog.getByLabel('Nome', { exact: true }).focus()
  await check(await dialog.getByLabel('Conversa ou canal do Telegram', { exact: true }).getAttribute('aria-invalid') === 'true', 'Telegram feedback explains accepted format')
  await check(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1), 'mobile page fits')
  await page.screenshot({ path: `/tmp/homealias-form-feedback-${width}.png`, fullPage: true })
  await check(errors.length === 0, `no browser errors: ${errors.join(', ')}`)
  await page.close()
  console.log(`PASS form feedback @ ${width}: blur, corrections, hints, password rules, invites, stale checks and natural language`)
 }
} finally { await browser.close() }
console.log(`${checks} form feedback checks passed`)
