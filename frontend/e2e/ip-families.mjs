import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
const { chromium } = await import(process.env.HOMEALIAS_PLAYWRIGHT_MODULE || 'playwright')
const templates = Object.fromEntries(['sh', 'ps1'].map(kind => [kind, readFileSync(new URL(`../../backend/internal/clientfiles/update.${kind}`, import.meta.url), 'utf8')]))
const browser = await chromium.launch({ executablePath: process.env.E2E_CHROME_PATH || '/usr/bin/google-chrome', headless: true, args: ['--no-sandbox'] })
const base = process.env.E2E_BASE_URL || 'http://127.0.0.1:5173'
try {
 for (const width of [375, 1280]) {
  for (const [label, enableA, enableAAAA] of [['A', true, false], ['AAAA', false, true], ['dual', true, true]]) {
   const context = await browser.newContext({ viewport: { width, height: 844 }, reducedMotion: 'reduce' })
   const page = await context.newPage()
   page.on('pageerror', error => console.error(error.message))
   page.on('console', message => { if (message.type() === 'error') console.error(message.text()) })
   page.on('requestfailed', request => console.error(request.url(), request.failure()?.errorText))
   await page.route('**/api/**', route => {
    const path = new URL(route.request().url()).pathname
    if (!path.startsWith('/api/')) return route.continue()
    const data = path === '/api/auth/me' ? { id: 'dummy', name: 'Teste', role: 'user', email: 'test@example.com' }
     : path === '/api/hosts' ? [{ id: 'dummy-host', fqdn: 'test.example.com', enable_a: enableA, enable_aaaa: enableAAAA }]
     : path === '/api/tokens' ? { token: 'dummy_token', token_prefix: 'dummy', instructions: {} } : []
    return route.fulfill({ json: data })
   })
   await page.route('**/client/update.*', route => route.fulfill({ contentType: 'text/plain', body: templates[new URL(route.request().url()).pathname.endsWith('.ps1') ? 'ps1' : 'sh'] }))
   await page.goto(base + '/tokens')
   await page.waitForLoadState('networkidle')
   const dialog = page.locator('dialog[open]')
   try {
    await dialog.getByLabel('Nome do token', { exact: true }).fill('Teste', { timeout: 10000 })
   } catch (error) {
    console.error(page.url(), await page.locator('body').innerText())
    throw error
   }
   await dialog.getByRole('button', { name: 'Gerar token', exact: true }).click()
   await page.locator('#panel-shell').getByRole('button', { name: 'Baixar', exact: true }).waitFor()
   const shell = await page.locator('#panel-shell pre').first().textContent()
   assert.ok(shell.includes('${HOMEALIAS_IPV4:-' + Number(enableA) + '}'))
   assert.ok(shell.includes('${HOMEALIAS_IPV6:-' + Number(enableAAAA) + '}'))
   assert.ok(!shell.includes('__HOMEALIAS_'))
   assert.equal(await page.getByText('Sem baixar nada: um único comando', { exact: true }).count(), 0)
   await page.getByRole('tab', { name: 'Script Linux', exact: true }).focus()
   await page.keyboard.press('End')
   const curlTab = page.getByRole('tab', { name: 'cURL', exact: true })
   assert.equal(await curlTab.getAttribute('aria-selected'), 'true')
   const curl = await page.locator('#panel-curl pre').textContent()
   assert.equal(curl.includes('curl -4 '), enableA)
   assert.equal(curl.includes('curl -6 '), enableAAAA)
   await page.getByRole('tab', { name: 'Docker', exact: true }).click()
   const docker = await page.locator('#panel-docker pre').first().textContent()
   assert.equal(docker.includes('curl -4 '), enableA)
   assert.equal(docker.includes('curl -6 '), enableAAAA)
   await page.getByRole('tab', { name: 'Windows', exact: true }).click()
   const downloadButton = page.locator('#panel-windows').getByRole('button', { name: 'Baixar', exact: true })
   await downloadButton.waitFor()
   const [download] = await Promise.all([page.waitForEvent('download'), downloadButton.click()])
   assert.equal(download.suggestedFilename(), 'homealias-update.ps1')
   const windows = readFileSync(await download.path(), 'utf8')
   assert.ok(windows.includes('$Ipv4 = ("' + Number(enableA) + '" -ne "0")'))
   assert.ok(windows.includes('$Ipv6 = ("' + Number(enableAAAA) + '" -eq "1")'))
   assert.ok(windows.includes('dummy_token') && windows.includes('test.example.com'))
   assert.ok(!windows.includes('__HOMEALIAS_'))
   assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= document.documentElement.clientWidth + 1))
   await context.close()
   console.log(`PASS ${label} @ ${width}: Linux, cURL tab/keyboard, Docker, Windows download`)
  }
 }
} finally { await browser.close() }
