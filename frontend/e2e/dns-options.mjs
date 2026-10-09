import assert from 'node:assert/strict'
const { chromium } = await import(process.env.HOMEALIAS_PLAYWRIGHT_MODULE || 'playwright')
const browser = await chromium.launch({ executablePath: process.env.E2E_CHROME_PATH || '/usr/bin/google-chrome', args: ['--no-sandbox'], headless: true })
const base = process.env.E2E_BASE_URL || 'http://127.0.0.1:5173'
try {
 for (const width of [375, 1280]) {
  const context = await browser.newContext({ viewport: { width, height: 900 }, reducedMotion: 'reduce' })
  const page = await context.newPage()
  let hosts = [], writes = [], failPatch = false
  await page.route('**/api/**', async route => {
   const path = new URL(route.request().url()).pathname
   if (!path.startsWith('/api/')) return route.continue()
   const method = route.request().method()
   if (path === '/api/auth/me') return route.fulfill({ json: { id: 'dummy-user', name: 'Teste', role: 'user', email: 'test@example.com' } })
   if (path === '/api/connections') return route.fulfill({ json: [{ id: 'dummy-connection', name: 'Teste', zones: [{ id: 'dummy-zone', name: 'example.com' }] }] })
   if (path === '/api/hosts/check') return route.fulfill({ json: { fqdn: 'test.example.com', in_app: false, records: [] } })
   if (path === '/api/hosts' && method === 'POST') {
    const body = route.request().postDataJSON()
    writes.push(body)
    hosts = [{ ...body, id: 'dummy-host', fqdn: 'test.example.com', status: 'never_seen' }]
    return route.fulfill({ status: 201, json: { id: 'dummy-host' } })
   }
   if (path === '/api/hosts/dummy-host' && method === 'PATCH') {
    if (failPatch) return route.fulfill({ status: 502, body: 'Cloudflare unavailable' })
    const body = route.request().postDataJSON()
    writes.push(body)
    hosts = [{ ...hosts[0], ...body }]
    return route.fulfill({ json: { id: 'dummy-host' } })
   }
   return route.fulfill({ json: path === '/api/hosts' ? hosts : [] })
  })
  await page.goto(base + '/hosts')
  await page.getByRole('button', { name: 'Novo host', exact: true }).first().click()
  const dialog = page.locator('dialog[open]')
  await dialog.getByLabel('Subdomínio', { exact: true }).fill('test')
  await dialog.getByLabel('Como acessar este endereço', { exact: true }).selectOption('true')
  assert.equal(await dialog.getByLabel('Tempo de cache (TTL)', { exact: true }).isDisabled(), true)
  await dialog.getByRole('button', { name: 'Criar host', exact: true }).click()
  await page.getByRole('button', { name: 'Editar host', exact: true }).waitFor()
  assert.equal(writes[0].proxied, true)
  assert.equal(writes[0].ttl, 1)
  await page.getByRole('button', { name: 'Editar host', exact: true }).click()
  await dialog.getByLabel('Como acessar este endereço', { exact: true }).selectOption('false')
  await dialog.getByLabel('Tempo de cache (TTL)', { exact: true }).selectOption('600')
  await dialog.getByLabel('IPv4 (registro A)', { exact: true }).uncheck()
  await dialog.getByRole('button', { name: 'Salvar alterações', exact: true }).click()
  await dialog.getByText('Selecione IPv4, IPv6 ou ambos.', { exact: true }).waitFor()
  const ipv4 = dialog.getByLabel('IPv4 (registro A)', { exact: true })
  assert.equal(await ipv4.getAttribute('aria-invalid'), 'true')
  assert.ok(await ipv4.evaluate(el => el === document.activeElement))
  assert.equal(writes.length, 1)
  await dialog.getByLabel('IPv4 (registro A)', { exact: true }).check()
  await dialog.getByLabel('IPv6 (registro AAAA)', { exact: true }).check()
  failPatch = true
  await dialog.getByRole('button', { name: 'Salvar alterações', exact: true }).click()
  await dialog.getByText('O servidor está indisponível. Tente novamente em instantes.', { exact: true }).waitFor()
  assert.equal(await dialog.getByLabel('Tempo de cache (TTL)', { exact: true }).inputValue(), '600')
  await page.screenshot({ path: `/tmp/homealias-dns-options-${width}.png`, fullPage: true })
  failPatch = false
  await dialog.getByRole('button', { name: 'Salvar alterações', exact: true }).click()
  await dialog.waitFor({ state: 'hidden' })
  assert.equal(writes[1].proxied, false)
  assert.equal(writes[1].ttl, 600)
  assert.equal(writes[1].enable_aaaa, true)
  await page.getByText('Acesso direto · cache: 600 s', { exact: true }).waitFor()
  assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1))
  await context.close()
  console.log(`PASS DNS options @ ${width}: create proxy, edit DNS-only, TTL, family validation, failure/retry`)
 }
} finally { await browser.close() }
