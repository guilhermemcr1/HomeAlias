import assert from 'node:assert/strict'
const { chromium } = await import(process.env.HOMEALIAS_PLAYWRIGHT_MODULE || 'playwright')
const browser = await chromium.launch({ executablePath: process.env.E2E_CHROME_PATH || '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] })
const base = process.env.E2E_BASE_URL || 'http://127.0.0.1:5173'
try {
 for (const width of [375, 1280]) {
  const context = await browser.newContext({ viewport: { width, height: 900 }, reducedMotion: 'reduce' })
  const page = await context.newPage()
  let connection = { id: 'connection', name: 'Conta pessoal', status: 'valid', api_token_suffix: '…old1' }
  let host = { id: 'host', name: 'old', zone_name: 'example.com', fqdn: 'old.example.com', status: 'online', enable_a: true, enable_aaaa: false, proxied: false, ttl: 1 }
  const connectionWrites = [], hostWrites = []
  let denyToken = false, denyHostname = false
  await page.route('**/api/**', async route => {
   const path = new URL(route.request().url()).pathname
   if (!path.startsWith('/api/')) return route.continue()
   if (path === '/api/auth/me') return route.fulfill({ json: { id: 'dummy', name: 'Teste', role: 'user', email: 'test@example.com' } })
   if (path === '/api/connections/connection') {
    const body = route.request().postDataJSON()
    connectionWrites.push(body)
    if (denyToken) return route.fulfill({ status: 400, body: 'O novo token precisa ter acesso a todos os domínios usados pelos hosts desta conexão.' })
    connection = { ...connection, name: body.name, api_token_suffix: body.api_token ? '…new1' : connection.api_token_suffix }
    return route.fulfill({ json: { id: connection.id } })
   }
   if (path === '/api/hosts/host') {
    const body = route.request().postDataJSON()
    hostWrites.push(body)
    if (denyHostname) return route.fulfill({ status: 409, body: 'O novo endereço já possui registros na Cloudflare. Escolha outro nome.' })
    host = { ...host, ...body, fqdn: body.name ? `${body.name}.example.com` : host.fqdn }
    return route.fulfill({ json: { id: host.id } })
   }
   return route.fulfill({ json: path === '/api/connections' ? [connection] : path === '/api/hosts' ? [host] : [] })
  })
  await page.goto(base + '/connections')
  await page.getByRole('button', { name: 'Editar conexão', exact: true }).click()
  const dialog = page.locator('dialog[open]')
  const token = dialog.getByLabel('Token de API da Cloudflare', { exact: true })
  assert.equal(await token.inputValue(), '')
  await dialog.getByLabel('Nome', { exact: true }).fill('Conta renomeada')
  await dialog.getByRole('button', { name: 'Validar e salvar', exact: true }).click()
  await dialog.waitFor({ state: 'hidden' })
  assert.equal('api_token' in connectionWrites[0], false)
  await page.getByRole('button', { name: 'Editar conexão', exact: true }).click()
  await token.fill('dummy_replacement_token_new1')
  denyToken = true
  await dialog.getByRole('button', { name: 'Validar e salvar', exact: true }).click()
  await dialog.getByText('O novo token precisa ter acesso a todos os domínios usados pelos hosts desta conexão.', { exact: true }).waitFor()
  assert.equal(await token.inputValue(), 'dummy_replacement_token_new1')
  assert.equal(connection.api_token_suffix, '…old1')
  denyToken = false
  await dialog.getByRole('button', { name: 'Validar e salvar', exact: true }).click()
  await dialog.waitFor({ state: 'hidden' })
  await page.getByText('…new1', { exact: true }).waitFor()
  assert.ok(!(await page.locator('body').innerText()).includes('dummy_replacement_token_new1'))
  await page.goto(base + '/hosts')
  await page.getByRole('button', { name: 'Editar host', exact: true }).click()
  const hostname = dialog.getByLabel('Endereço completo', { exact: true })
  await hostname.fill('outside.other.com')
  await dialog.getByRole('button', { name: 'Salvar alterações', exact: true }).click()
  await dialog.getByText('Use um endereço do domínio example.com.', { exact: true }).waitFor()
  assert.equal(hostWrites.length, 0)
  await hostname.fill('new.example.com')
  denyHostname = true
  await dialog.getByRole('button', { name: 'Salvar alterações', exact: true }).click()
  await dialog.getByText('O novo endereço já possui registros na Cloudflare. Escolha outro nome.', { exact: true }).waitFor()
  assert.equal(await hostname.inputValue(), 'new.example.com')
  await page.screenshot({ path: `/tmp/homealias-edit-resources-${width}.png`, fullPage: true })
  denyHostname = false
  await dialog.getByRole('button', { name: 'Salvar alterações', exact: true }).click()
  await dialog.waitFor({ state: 'hidden' })
  assert.equal(hostWrites.at(-1).name, 'new')
  assert.equal(host.id, 'host')
  await page.getByText('new.example.com', { exact: true }).waitFor()
  assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1))
  await context.close()
  console.log(`PASS edit resources @ ${width}: retain token, replace/retry, secret hidden, rename/conflict/zone validation`)
 }
} finally { await browser.close() }
