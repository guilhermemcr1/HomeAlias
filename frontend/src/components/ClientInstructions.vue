<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { Container, Monitor, Router, Terminal } from 'lucide-vue-next'
import CodeBlock from './CodeBlock.vue'
import FieldCopy from './FieldCopy.vue'
import * as snip from '../composables/clientSnippets'
import LoadError from './LoadError.vue'

const props = defineProps<{ token: string; hostname: string; ipv4Enabled: boolean; ipv6Enabled: boolean }>()

const cfg = computed<snip.ClientConfig>(() => ({ origin: window.location.origin, hostname: props.hostname, token: props.token, enableA: props.ipv4Enabled, enableAAAA: props.ipv6Enabled }))

const tabs = [
  { key: 'shell', label: 'Script Linux', icon: Terminal },
  { key: 'docker', label: 'Docker', icon: Container },
  { key: 'router', label: 'Roteador (DDNS)', icon: Router },
  { key: 'windows', label: 'Windows', icon: Monitor },
  { key: 'curl', label: 'cURL', icon: Terminal },
] as const
type Tab = (typeof tabs)[number]['key']
const active = ref<Tab>('shell')
const tabButtons = ref<HTMLButtonElement[]>([])

// Setas, Home e End trocam de aba (padrão WAI-ARIA de tabs).
function onKey(e: KeyboardEvent, i: number) {
  const last = tabs.length - 1
  const next = { ArrowRight: i === last ? 0 : i + 1, ArrowLeft: i === 0 ? last : i - 1, Home: 0, End: last }[e.key]
  if (next === undefined) return
  e.preventDefault()
  active.value = tabs[next].key
  tabButtons.value[next]?.focus()
}

const fields = computed(() => snip.routerFields(cfg.value))

// Busca o modelo de /client/*, preenche URL, token e hostname aqui no navegador e exibe o script pronto.
const scripts = ref<{ sh: string; ps1: string }>({ sh: '', ps1: '' })
const pending = ref({ sh: false, ps1: false })
const scriptErrors = ref({ sh: '', ps1: '' })
const controllers = new Set<AbortController>()
onBeforeUnmount(() => { for (const controller of controllers) controller.abort() })

async function loadScript(kind: 'sh' | 'ps1') {
  if (scripts.value[kind] || pending.value[kind]) return
  pending.value[kind] = true
  scriptErrors.value[kind] = ''
  const controller = new AbortController()
  controllers.add(controller)
  let timedOut = false
  const timer = setTimeout(() => { timedOut = true; controller.abort() }, 30_000)
  try {
    const res = await fetch(`/client/update.${kind}`, { signal: controller.signal })
    if (!res.ok) throw new Error(String(res.status))
    const text = await res.text()
    if (!snip.isScriptTemplate(text, kind)) throw new Error('invalid script')
    scripts.value[kind] = text
  } catch {
    if (!controller.signal.aborted || timedOut) scriptErrors.value[kind] = 'Não foi possível carregar o script. Verifique sua conexão e tente novamente.'
  } finally {
    clearTimeout(timer)
    controllers.delete(controller)
    pending.value[kind] = false
  }
}

const shScript = computed(() => (scripts.value.sh ? snip.configureScript(scripts.value.sh, cfg.value) : ''))
const ps1Script = computed(() => (scripts.value.ps1 ? snip.configureScript(scripts.value.ps1, cfg.value) : ''))

watch(
  active,
  (t) => {
    if (t === 'shell') loadScript('sh')
    else if (t === 'windows') loadScript('ps1')
  },
  { immediate: true },
)
</script>

<template>
  <div class="space-y-4">
    <p class="text-sm text-base-content/80">
      Atualização automática: {{ props.ipv4Enabled && props.ipv6Enabled ? 'IPv4 (A) e IPv6 (AAAA), com uma atualização para cada tipo.' : props.ipv4Enabled ? 'IPv4 (A).' : 'IPv6 (AAAA).' }}
      Seu dispositivo precisa ter acesso aos tipos de IP selecionados.
    </p>
    <div class="overflow-x-auto pb-1">
      <div role="tablist" aria-label="Forma de configuração" class="inline-flex min-w-full gap-1 rounded-box bg-base-200 p-1 sm:min-w-0">
        <button
          v-for="(t, i) in tabs"
          :id="`tab-${t.key}`"
          :key="t.key"
          :ref="(el) => (tabButtons[i] = el as HTMLButtonElement)"
          type="button"
          role="tab"
          :aria-selected="active === t.key"
          :aria-controls="`panel-${t.key}`"
          :tabindex="active === t.key ? 0 : -1"
          class="flex h-11 flex-1 items-center justify-center gap-2 whitespace-nowrap rounded-btn border px-4 text-sm font-medium transition-colors sm:flex-none"
          :class="active === t.key ? 'border-base-300 bg-base-100 text-base-content' : 'border-transparent text-base-content/65 hover:text-base-content'"
          @click="active = t.key"
          @keydown="onKey($event, i)"
        >
          <component :is="t.icon" :size="16" aria-hidden="true" />{{ t.label }}
        </button>
      </div>
    </div>

    <!-- Script Linux -->
    <div v-if="active === 'shell'" id="panel-shell" role="tabpanel" aria-labelledby="tab-shell" class="space-y-4">
      <p class="text-base-content/80">Para Linux, macOS, NAS e Raspberry Pi. O arquivo já vem com o endereço do seu HomeAlias, o host e o token preenchidos.</p>
      <ol class="list-decimal space-y-3 pl-5 marker:font-semibold">
        <li>
          Copie ou baixe o script abaixo e salve como <code class="font-data">homealias-update.sh</code>.
          <CodeBlock v-if="shScript" class="mt-2" :code="shScript" caption="homealias-update.sh" filename="homealias-update.sh" scroll />
          <LoadError v-else-if="scriptErrors.sh" class="mt-2" :message="scriptErrors.sh" :busy="pending.sh" @retry="loadScript('sh')" />
          <div v-else class="skeleton mt-2 h-32 w-full" role="status" aria-label="Carregando script" aria-busy="true" />
          <p class="mt-1 text-sm text-base-content/70">O script contém o token: guarde-o com permissão só para você (<code class="font-data">chmod 700</code>).</p>
        </li>
        <li>Teste agora:<CodeBlock class="mt-2" code="chmod 700 homealias-update.sh && ./homealias-update.sh" caption="Terminal" /></li>
        <li>Agende a cada 5 minutos (<code class="font-data">crontab -e</code>):<CodeBlock class="mt-2" :code="snip.cronLine()" caption="crontab" /></li>
      </ol>
    </div>

    <!-- Docker -->
    <div v-else-if="active === 'docker'" id="panel-docker" role="tabpanel" aria-labelledby="tab-docker" class="space-y-4">
      <p class="text-base-content/80">Este comando usa a imagem oficial <code class="font-data">curlimages/curl</code> e repete a atualização a cada 5 minutos.</p>
      <CodeBlock :code="snip.dockerRun(cfg)" :mask="props.token" caption="docker run" />
      <details class="rounded-btn border border-base-300 p-3">
        <summary class="cursor-pointer font-medium">Prefere Docker Compose?</summary>
        <CodeBlock class="mt-3" :code="snip.dockerCompose(cfg)" :mask="props.token" caption="docker-compose.yml" />
      </details>
      <p class="text-sm text-base-content/70">Ver o resultado: <code class="font-data">docker logs -f homealias-ddns</code></p>
    </div>

    <!-- Roteador -->
    <div v-else-if="active === 'router'" id="panel-router" role="tabpanel" aria-labelledby="tab-router" class="space-y-4">
      <p class="text-base-content/80">
        No roteador ou firewall, procure por DDNS / "serviço personalizado" e preencha os campos abaixo. O IP é detectado pelo
        servidor, então o roteador não precisa informá-lo.
      </p>
      <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
        <FieldCopy v-for="f in fields" :key="f.label" v-bind="f" :class="f.wide && 'sm:col-span-2'" />
      </div>
      <p class="text-xs text-base-content/65">Clique em um campo para copiar. Dados sensíveis ficam ocultos até você passar o mouse ou copiar.</p>
      <details class="rounded-btn border border-base-300 p-3">
        <summary class="cursor-pointer font-medium">Testar pelo computador</summary>
        <p class="mt-2 text-sm text-base-content/70">A resposta good confirma a atualização. nochg indica que o IP já estava correto.</p>
        <CodeBlock class="mt-3" :code="snip.routerTest(cfg)" :mask="props.token" caption="cURL" />
      </details>
      <details class="rounded-btn border border-base-300 p-3">
        <summary class="cursor-pointer font-medium">OpenWrt (/etc/config/ddns)</summary>
        <CodeBlock class="mt-3" :code="snip.openwrtConfig(cfg)" :mask="props.token" caption="ddns-scripts" />
      </details>
      <details class="rounded-btn border border-base-300 p-3">
        <summary class="cursor-pointer font-medium">Ubiquiti EdgeRouter (EdgeOS)</summary>
        <CodeBlock class="mt-3" :code="snip.edgeosConfig(cfg)" :mask="props.token" caption="CLI" />
      </details>
    </div>

    <!-- Windows -->
    <div v-else-if="active === 'windows'" id="panel-windows" role="tabpanel" aria-labelledby="tab-windows" class="space-y-4">
      <p class="text-base-content/80">Um arquivo só: ele já vem configurado e cria a tarefa agendada sozinho.</p>
      <ol class="list-decimal space-y-3 pl-5 marker:font-semibold">
        <li>
          Copie ou baixe o script abaixo e salve como <code class="font-data">homealias-update.ps1</code>.
          <CodeBlock v-if="ps1Script" class="mt-2" :code="ps1Script" caption="homealias-update.ps1" filename="homealias-update.ps1" scroll />
          <LoadError v-else-if="scriptErrors.ps1" class="mt-2" :message="scriptErrors.ps1" :busy="pending.ps1" @retry="loadScript('ps1')" />
          <div v-else class="skeleton mt-2 h-32 w-full" role="status" aria-label="Carregando script" aria-busy="true" />
          <p class="mt-1 text-sm text-base-content/70">O script contém o token: guarde-o numa pasta só sua, como <code class="font-data">C:\HomeAlias</code>.</p>
        </li>
        <li>No PowerShell, na pasta do arquivo, crie a tarefa (roda a cada 5 minutos e já atualiza agora):<CodeBlock class="mt-2" :code="snip.windowsInstall()" caption="PowerShell" /></li>
        <li>Para remover depois: <code class="font-data">.\homealias-update.ps1 -Uninstall</code></li>
      </ol>
    </div>

    <!-- cURL -->
    <div v-else-if="active === 'curl'" id="panel-curl" role="tabpanel" aria-labelledby="tab-curl" class="space-y-4">
      <p class="text-base-content/80">Execute no terminal para atualizar agora. Os comandos usam IPv4 ou IPv6 conforme os registros habilitados no host.</p>
      <CodeBlock :code="snip.curlCommand(cfg)" :mask="props.token" caption="cURL" />
      <p class="text-sm text-base-content/70">Cada vez que você executar o comando, ele enviará uma atualização. Para fazer isso automaticamente, use um dos scripts ou o Docker.</p>
    </div>
  </div>
</template>
