import assert from 'node:assert/strict'
const { chromium } = await import(process.env.HOMEALIAS_PLAYWRIGHT_MODULE || 'playwright')
const browser = await chromium.launch({ executablePath: process.env.E2E_CHROME_PATH || '/usr/bin/google-chrome', args: ['--no-sandbox'], headless: true })
const base = process.env.E2E_BASE_URL || 'http://127.0.0.1:5173'
try {
  for (const width of [320, 375, 1280]) {
    const context = await browser.newContext({ viewport: { width, height: 800 }, reducedMotion: 'reduce' })
    await context.addCookies([{ name: 'homealias_csrf', value: 'test-csrf', url: base }])
    const page = await context.newPage()
    const errors = [], writes = []
    page.on('pageerror', error => errors.push(error.message))
    let connections = [{ id: 'connection', name: 'Conta da Cloudflare ' + 'A'.repeat(60), status: 'valid', api_token_suffix: '…abcd' }]
    let outcome = 409, release, started
    await page.route('**/api/**', async route => {
      const request = route.request(), path = new URL(request.url()).pathname
      if (!path.startsWith('/api/')) return route.continue()
      if (path === '/api/auth/me') return route.fulfill({ json: { id: 'owner', name: 'Ana', email: 'ana@example.com', role: 'user' } })
      if (path === '/api/connections') return route.fulfill({ json: connections })
      if (path === '/api/connections/connection' && request.method() === 'DELETE') {
        writes.push(request.headers()['x-csrf-token'])
        if (outcome === 'offline') return route.abort('failed')
        if (outcome === 'slow') {
          started = true
          await new Promise(resolve => { release = resolve })
          connections = []
          return route.fulfill({ status: 204 })
        }
        return route.fulfill({ status: outcome, body: outcome === 409 ? 'Esta conexão ainda tem hosts vinculados. Remova esses hosts na aba Hosts antes de excluir a conexão.' : 'internal details' })
      }
      return route.fulfill({ json: [] })
    })
    await page.goto(base + '/connections')
    const row = page.getByRole('listitem').filter({ has: page.getByRole('button', { name: 'Excluir conexão', exact: true }) })
    const dialog = page.getByRole('dialog', { name: 'Excluir conexão?' })
    await row.getByRole('button', { name: 'Excluir conexão', exact: true }).click()
    await dialog.waitFor()
    assert.ok((await dialog.innerText()).includes(connections[0].name))
    await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
    assert.equal(writes.length, 0)
    const messages = {
      409: 'Esta conexão ainda tem hosts vinculados.',
      403: 'Você não tem permissão',
      404: 'Item não encontrado',
      500: 'Erro no servidor',
      offline: 'Não foi possível falar com o servidor',
    }
    for (const result of [409, 403, 404, 500, 'offline']) {
      outcome = result
      await row.getByRole('button', { name: 'Excluir conexão', exact: true }).click()
      await dialog.getByRole('button', { name: 'Excluir conexão', exact: true }).click()
      await dialog.getByRole('alert').filter({ hasText: messages[result] }).waitFor()
      assert.equal(await row.count(), 1)
      assert.equal(await dialog.getByRole('button', { name: 'Cancelar', exact: true }).isEnabled(), true)
      assert.ok(await dialog.locator('.modal-box').evaluate(el => el.scrollWidth <= el.clientWidth + 1))
      assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1))
      if (result === 409) {
        await page.screenshot({ path: `/tmp/homealias-delete-connection-${width}.png`, fullPage: true })
        await page.evaluate(() => document.documentElement.style.fontSize = '200%')
        assert.ok(await dialog.locator('.modal-box').evaluate(el => el.scrollWidth <= el.clientWidth + 1))
        await page.evaluate(() => document.documentElement.style.fontSize = '')
      }
      await dialog.getByRole('button', { name: 'Cancelar', exact: true }).click()
    }
    await row.getByRole('button', { name: 'Excluir conexão', exact: true }).click()
    assert.equal(await dialog.getByRole('alert').count(), 0)
    outcome = 'slow'
    const countBefore = writes.length
    await dialog.getByRole('button', { name: 'Excluir conexão', exact: true }).click()
    for (let attempt = 0; !started && attempt < 100; attempt++) await page.waitForTimeout(25)
    assert.ok(started)
    assert.equal(await dialog.getByRole('button', { name: 'Cancelar', exact: true }).isDisabled(), true)
    const submit = dialog.getByRole('button', { name: 'Excluindo…', exact: true })
    assert.equal(await submit.isDisabled(), true)
    await submit.evaluate(button => { for (let i = 0; i < 10; i++) button.click() })
    await page.keyboard.press('Escape')
    assert.equal(await dialog.isVisible(), true)
    assert.equal(writes.length, countBefore + 1)
    release()
    await dialog.waitFor({ state: 'hidden' })
    await page.getByRole('heading', { name: 'Nenhuma conexão ainda', exact: true }).waitFor()
    await page.getByRole('status').filter({ hasText: 'Conexão excluída.' }).waitFor()
    assert.ok(writes.every(token => token === 'test-csrf'))
    assert.deepEqual(errors, [])
    await context.close()
    console.log(`PASS delete connection @ ${width}: cancel, conflict, forbidden, missing, server/offline failure, retry, duplicate clicks, CSRF, empty state and 200% text`)
  }
} finally { await browser.close() }
