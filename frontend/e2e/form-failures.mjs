import assert from 'node:assert/strict'
const { chromium } = await import(process.env.HOMEALIAS_PLAYWRIGHT_MODULE || 'playwright')
const browser = await chromium.launch({ executablePath: process.env.E2E_CHROME_PATH || '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] })
const base = process.env.E2E_BASE_URL || 'http://127.0.0.1:5173'
let checks = 0
const scenarios = [
 ['slow', 'O servidor está indisponível. Tente novamente em instantes.'],
 [400, 'Revise os dados informados.'],
 [403, 'Você não tem permissão para fazer isso.'],
 [404, 'Item não encontrado. Atualize a lista e tente novamente.'],
 [409, 'Este nome já está em uso. Escolha outro.'],
 [429, 'Muitas tentativas. Aguarde um minuto e tente de novo.'],
 [500, 'Erro no servidor. Tente novamente em instantes.'],
 [502, 'O servidor está indisponível. Tente novamente em instantes.'],
 [503, 'O servidor está indisponível. Tente novamente em instantes.'],
 [504, 'O servidor demorou para responder. Tente novamente em instantes.'],
 ['malformed', 'O servidor enviou uma resposta inválida. Tente novamente.'],
 ['offline', 'Não foi possível falar com o servidor. Verifique sua conexão e tente novamente.'],
]
try {
 for (const width of [320, 375, 1280]) {
  const context = await browser.newContext({ viewport: { width, height: 844 }, reducedMotion: 'reduce' })
  const page = await context.newPage()
  const errors = []
  page.on('pageerror', error => errors.push(error.message))
  let mode = 500, writes = 0
  await page.route('**/api/**', async route => {
   const path = new URL(route.request().url()).pathname
   if (!path.startsWith('/api/')) return route.continue()
   if (route.request().method() !== 'GET') {
    writes++
    if (mode === 'offline') return route.abort('failed')
    if (mode === 'slow') await new Promise(resolve => setTimeout(resolve, 1000))
    const status = mode === 'malformed' ? 200 : mode === 'slow' ? 503 : Number(mode)
    return route.fulfill({ status, body: status === 400 ? 'Revise os dados informados.' : status === 409 ? 'Este nome já está em uso. Escolha outro.' : '<html>private server trace</html>' })
   }
   const fixtures = {
    '/api/auth/me': { id: 'dummy', name: 'Teste', email: 'test@example.com', role: 'admin' },
    '/api/connections': [{ id: 'connection', name: 'Conta', status: 'valid', zones: [{ id: 'zone', name: 'example.com' }] }],
    '/api/hosts': [{ id: 'host', fqdn: 'old.example.com', zone_name: 'example.com', status: 'online', enable_a: true, enable_aaaa: false, proxied: false, ttl: 1 }],
   }
   return route.fulfill({ json: fixtures[path] ?? [] })
  })
  for (const [path, opener, submit, field, value] of [
   ['/connections', 'Editar conexão', 'Validar e salvar', 'Token de API da Cloudflare', 'dummy_replacement_token_1234'],
   ['/hosts', 'Editar host', 'Salvar alterações', 'Endereço completo', 'new.example.com'],
   ['/tokens', null, 'Gerar token', 'Nome do token', 'Teste'],
   ['/alerts', 'Novo canal', 'Cadastrar canal', 'Endereço de e-mail', 'test@example.com'],
  ]) {
   await page.goto(base + path)
   await page.waitForLoadState('networkidle')
   if (opener) await page.getByRole('button', { name: opener, exact: true }).first().click()
   const dialog = page.locator('dialog[open]')
   const input = dialog.getByLabel(field, { exact: true })
   await input.fill(value)
   if (path === '/alerts') await dialog.getByLabel('Nome', { exact: true }).fill('Meu canal')
   for (const [nextMode, message] of scenarios) {
    mode = nextMode
    const before = writes
    if (mode === 'slow') {
     await dialog.locator('form').evaluate(form => { form.requestSubmit(); form.requestSubmit(); form.requestSubmit() })
     assert.equal(await input.isDisabled(), true)
     assert.equal(await dialog.getByRole('button', { name: 'Cancelar', exact: true }).isDisabled(), true)
     await page.keyboard.press('Escape')
     assert.equal(await dialog.count(), 1)
     checks += 3
    } else await dialog.getByRole('button', { name: submit, exact: true }).click()
    await dialog.getByText(message, { exact: true }).waitFor()
    assert.equal(writes, before + 1)
    assert.equal(await input.inputValue(), value)
    assert.equal(await input.isDisabled(), false)
    assert.ok(!(await dialog.innerText()).includes('private server trace'))
    assert.ok(await dialog.evaluate(el => el.scrollWidth <= el.clientWidth + 1))
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1))
    checks += 6
   }
   await page.screenshot({ path: `/tmp/homealias-form-failures-${path.slice(1)}-${width}.png`, fullPage: true })
   await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
   console.log(`PASS failures ${path} @ ${width}: 12 error/slow/offline scenarios`)
  }
  mode = 401
  await page.goto(base + '/tokens')
  const dialog = page.locator('dialog[open]')
  await dialog.getByLabel('Nome do token', { exact: true }).fill('Teste')
  await dialog.getByRole('button', { name: 'Gerar token', exact: true }).click()
  await page.waitForURL(url => url.pathname === '/login' && url.searchParams.get('next') === '/tokens')
  assert.equal(await page.locator('.app-header').count(), 0)
  assert.equal(await page.locator('dialog[open]').count(), 0)
  assert.equal(errors.length, 0, errors.join('\n'))
  checks += 3
  await context.close()
 }
} finally { await browser.close() }
console.log(`${checks} form failure checks passed`)
