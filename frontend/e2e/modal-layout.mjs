import assert from 'node:assert/strict'
const { chromium } = await import(process.env.HOMEALIAS_PLAYWRIGHT_MODULE || 'playwright')
const base = process.env.E2E_BASE_URL || 'http://127.0.0.1:5173'
const browser=await chromium.launch({executablePath:process.env.E2E_CHROME_PATH || '/usr/bin/google-chrome',args:['--no-sandbox'],headless:true})
try{
 for(const width of [320,375,768,1280]){
  const errors=[]
  const page=await browser.newPage({viewport:{width,height:900},reducedMotion:'reduce'})
  page.on('pageerror', error => errors.push(error.message))
  await page.route('**/api/**',route=>{
   const path=new URL(route.request().url()).pathname
   if(!path.startsWith('/api/'))return route.continue()
   const data={
    '/api/auth/me':{id:'user',name:'Ana',email:'ana@example.com',role:'admin'},
    '/api/connections':[{id:'connection',name:'Conta pessoal',zones:[{id:'zone',name:'example.com'}],status:'valid'}],
    '/api/hosts':[{id:'host',fqdn:'casa.example.com',zone_name:'example.com',enable_a:true,enable_aaaa:true,proxied:false,ttl:300,status:'online'}],
    '/api/users':[{id:'other',email:'other@example.com',name:'Outro usuário',role:'user',status:'active'}],
   }
   return route.fulfill({json:data[path]||[]})
  })
  for(const [path,opener,name]of[['/tokens',null,'token'],['/connections','Nova conexão','cloudflare'],['/hosts','Editar host','host'],['/conta','Editar perfil','profile'],['/admin/users','Novo usuário','user'],['/alerts','Novo canal','alert']]){
   await page.goto(base+path)
   await page.waitForLoadState('networkidle')
   if(opener)await page.getByRole('button',{name:opener,exact:true}).first().click()
   const box=page.locator('dialog[open] .modal-box')
   await box.waitFor(); await page.waitForFunction(()=>{const el=document.querySelector('dialog[open] .modal-box');return el&&new DOMMatrix(getComputedStyle(el).transform).a>0.999})
   const initial=await box.evaluate(el=>({height:el.scrollHeight,width:el.getBoundingClientRect().width,columns:[...el.querySelectorAll('.form-grid')].map(g=>getComputedStyle(g).gridTemplateColumns.split(' ').filter(x=>parseFloat(x)>0).length)}))
   assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=document.documentElement.clientWidth+1))
   assert.ok(await box.evaluate(el=>el.scrollWidth<=el.clientWidth+1))
   if(width>=768){assert.ok(initial.width>=672,JSON.stringify(initial));if(name!=='cloudflare')assert.ok(initial.columns.every(c=>c===2))}
   else assert.ok(initial.columns.every(c=>c===1))
   if(width===375||width===1280)await page.screenshot({path:`/tmp/homealias-modal-${name}-${width}.png`,fullPage:true})
   await page.evaluate(()=>document.documentElement.style.fontSize='200%')
   assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=document.documentElement.clientWidth+1))
   assert.ok(await box.evaluate(el=>el.scrollWidth<=el.clientWidth+1))
   await page.locator('dialog[open]').getByRole('button',{name:'Cancelar',exact:true}).scrollIntoViewIfNeeded()
   assert.ok(await page.locator('dialog[open]').getByRole('button',{name:'Cancelar',exact:true}).isVisible())
   await page.evaluate(()=>document.documentElement.style.fontSize='')
   await page.setViewportSize({width,height:500})
   const cancel=page.locator('dialog[open]').getByRole('button',{name:'Cancelar',exact:true})
   await cancel.scrollIntoViewIfNeeded()
   const rect=await cancel.boundingBox()
   assert.ok(rect && rect.y>=0 && rect.y+rect.height<=500)
   await page.setViewportSize({width,height:900})
   console.log(`PASS modal ${name} @ ${width}: width, columns, long hints, 200% text, scroll`)
  }
  assert.deepEqual(errors,[])
  await page.close()
 }
}finally{await browser.close()}
