<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { CircleCheck, Globe, KeyRound, OctagonX, Pencil, Plus, RefreshCw, Trash2, TriangleAlert } from 'lucide-vue-next'
import { ApiError, api, errorMessage } from '../api/client'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import LoadError from '../components/LoadError.vue'
import EmptyState from '../components/EmptyState.vue'
import FormField from '../components/FormField.vue'
import FormModal from '../components/FormModal.vue'
import PageHeader from '../components/PageHeader.vue'
import SkeletonRows from '../components/SkeletonRows.vue'
import StatusBadge from '../components/StatusBadge.vue'
import HostDnsOptions from '../components/HostDnsOptions.vue'
import { relativeTime } from '../composables/format'
import { useToast } from '../composables/useToast'
import * as v from '../composables/validation'

type Zone = { id: string; name: string }
type Connection = { id: string; name: string; zones?: string | Zone[] | null }
type Host = {
  id: string
  fqdn: string
  zone_name: string
  status: string
  enable_a: boolean
  enable_aaaa: boolean
  proxied: boolean
  ttl: number
  last_ipv4?: string | null
  last_ipv6?: string | null
  last_seen_v4?: string | null
  last_seen_v6?: string | null
}

const toast = useToast()
const hosts = ref<Host[]>([])
const connections = ref<Connection[]>([])
const loading = ref(true)
const loadError = ref('')
let reloading = false
const showForm = ref(false)
const busy = ref(false)
const error = ref('')
const syncing = ref('')
const attempted = ref(false)

const form = ref({
  connection_id: '',
  zone_id: '',
  name: '',
  enable_a: true,
  enable_aaaa: false,
  proxied: false,
  ttl: 1,
  adopt_existing: false,
})

// Verificação prévia: o nome já existe no app ou na Cloudflare?
type Existing = { type: string; content: string; ttl: number; proxied: boolean }
type Check = { fqdn: string; in_app: boolean; records: Existing[] }
const check = ref<Check | null>(null)
const checking = ref(false)
const checkError = ref('')
let checkTimer: ReturnType<typeof setTimeout> | undefined
let checkSeq = 0
let checkController: AbortController | undefined

function parseZones(c?: Connection): Zone[] {
  if (!c?.zones) return []
  if (Array.isArray(c.zones)) return c.zones
  try {
    const zones: unknown = JSON.parse(c.zones)
    return Array.isArray(zones) ? zones.filter((z): z is Zone => typeof z?.id === 'string' && typeof z?.name === 'string') : []
  } catch {
    return []
  }
}
const zones = computed(() => parseZones(connections.value.find((c) => c.id === form.value.connection_id)))
const zone = computed(() => zones.value.find((z) => z.id === form.value.zone_id))
const errs = computed(() => ({
  connection: connections.value.some(c => c.id === form.value.connection_id) ? '' : 'Escolha uma conexão disponível.',
  zone: zones.value.some(z => z.id === form.value.zone_id) ? '' : 'Escolha um domínio disponível.',
  name: v.hostName(form.value.name) || (zone.value ? v.fqdn(form.value.name, zone.value.name) : ''),
  records: form.value.enable_a || form.value.enable_aaaa ? '' : 'Selecione IPv4, IPv6 ou ambos.',
}))
const preview = computed(() => (form.value.name && zone.value ? form.value.name.trim() === '@' ? zone.value.name : `${form.value.name.trim().toLowerCase()}.${zone.value.name}` : ''))

async function runCheck() {
  const seq = ++checkSeq
  checkController?.abort()
  check.value = null
  checkError.value = ''
  form.value.adopt_existing = false
  if (!showForm.value || !zone.value || errs.value.name || !form.value.connection_id) {
    checking.value = false
    return
  }
  checking.value = true
  const controller = new AbortController()
  checkController = controller
  try {
    const res = await api<Check>('/api/hosts/check', {
      method: 'POST',
      signal: controller.signal,
      body: JSON.stringify({
        connection_id: form.value.connection_id,
        zone_id: form.value.zone_id,
        zone_name: zone.value.name,
        name: form.value.name.trim().toLowerCase(),
      }),
    })
    if (seq === checkSeq) check.value = res
  } catch (e) {
    if (seq === checkSeq && !controller.signal.aborted) checkError.value = errorMessage(e, 'Não foi possível verificar a Cloudflare.')
  } finally {
    if (seq === checkSeq) checking.value = false
  }
}

// Espera o usuário parar de digitar antes de consultar a Cloudflare.
watch(
  () => [showForm.value, form.value.connection_id, form.value.zone_id, form.value.name.trim().toLowerCase()],
  () => {
    ++checkSeq
    checkController?.abort()
    clearTimeout(checkTimer)
    check.value = null
    checkError.value = ''
    form.value.adopt_existing = false
    checking.value = !!(showForm.value && zone.value && !errs.value.name)
    if (checking.value) checkTimer = setTimeout(runCheck, 500)
  },
)
onBeforeUnmount(() => { ++checkSeq; clearTimeout(checkTimer); checkController?.abort() })

const conflict = computed(() => !!check.value && (check.value.in_app || check.value.records.length > 0))
const blocked = computed(() => !!check.value?.in_app || checking.value || (!!check.value?.records.length && !form.value.adopt_existing))

async function reload() {
  if (reloading) return
  reloading = true
  loading.value = true
  try {
    ;[hosts.value, connections.value] = await Promise.all([
      api<Host[]>('/api/hosts').then((r) => r ?? []),
      api<Connection[]>('/api/connections').then((r) => r ?? []),
    ])
    if (connections.value.length === 1 && !form.value.connection_id) form.value.connection_id = connections.value[0].id
    loadError.value = ''
  } catch (e) {
    loadError.value = errorMessage(e)
  } finally {
    loading.value = false
    reloading = false
  }
}
onMounted(reload)

function pickConnection() {
  form.value.zone_id = zones.value.length === 1 ? zones.value[0].id : ''
}

function openForm() {
  error.value = ''
  attempted.value = false
  form.value.name = ''
  form.value.proxied = false
  form.value.ttl = 1
  check.value = null
  checkError.value = ''
  form.value.adopt_existing = false
  if (connections.value.length === 1) {
    form.value.connection_id = connections.value[0].id
    pickConnection()
  }
  showForm.value = true
}

async function create() {
  if (busy.value) return
  error.value = ''
  attempted.value = true
  if (Object.values(errs.value).some(Boolean) || !zone.value) return
  busy.value = true
  try {
    await api('/api/hosts', {
      method: 'POST',
      body: JSON.stringify({ ...form.value, ttl: form.value.proxied ? 1 : form.value.ttl, name: form.value.name.trim().toLowerCase(), zone_name: zone.value.name }),
    })
    toast.success(`${preview.value} criado. Gere um token para começar a atualizar.`)
    showForm.value = false
    await reload()
  } catch (e) {
    // Alguém criou o registro entre a verificação e o envio: mostra a pergunta de novo.
    if (e instanceof ApiError && e.status === 409) await runCheck()
    error.value = e instanceof ApiError && e.status === 409 && conflict.value ? '' : errorMessage(e, 'Não foi possível criar o host.')
  } finally {
    busy.value = false
  }
}

async function sync(h: Host) {
  if (syncing.value) return
  syncing.value = h.id
  try {
    await api(`/api/hosts/${h.id}/sync`, { method: 'POST' })
    toast.success(`${h.fqdn} sincronizado com a Cloudflare.`)
  } catch (e) {
    toast.error(errorMessage(e))
  } finally {
    syncing.value = ''
  }
}

const removing = ref<Host | null>(null)
const alsoDeleteDns = ref(true)
const removeBusy = ref(false)

function askRemove(h: Host) {
  alsoDeleteDns.value = true
  removing.value = h
}

async function confirmRemove() {
  if (removeBusy.value) return
  if (!removing.value) return
  removeBusy.value = true
  try {
    await api(`/api/hosts/${removing.value.id}?delete_dns=${alsoDeleteDns.value}`, { method: 'DELETE' })
    toast.success(`${removing.value.fqdn} removido.`)
    removing.value = null
    await reload()
  } catch (e) {
    toast.error(errorMessage(e))
  } finally {
    removeBusy.value = false
  }
}

const editing = ref<Host | null>(null)
const editForm = ref({ enable_a: true, enable_aaaa: false, proxied: false, ttl: 1 })
const editBusy = ref(false)
const editError = ref('')
const editAttempted = ref(false)
const editHostname = ref('')
const editName = computed(() => {
  const hostname = editHostname.value.trim().toLowerCase()
  const zoneName = editing.value?.zone_name ?? ''
  return hostname === zoneName ? '@' : hostname.endsWith(`.${zoneName}`) ? hostname.slice(0, -(zoneName.length + 1)) : ''
})
const editHostnameError = computed(() => !editing.value?.zone_name ? 'Não foi possível identificar o domínio. Recarregue a página.' : !editName.value ? `Use um endereço do domínio ${editing.value.zone_name}.` : v.hostName(editName.value) || v.fqdn(editName.value, editing.value.zone_name))
const editRecordsError = computed(() => editForm.value.enable_a || editForm.value.enable_aaaa ? '' : 'Selecione IPv4, IPv6 ou ambos.')

function openEdit(h: Host) {
  editForm.value = { enable_a: h.enable_a ?? true, enable_aaaa: h.enable_aaaa ?? false, proxied: h.proxied ?? false, ttl: h.ttl ?? 1 }
  editError.value = ''
  editAttempted.value = false
  editing.value = h
  editHostname.value = h.fqdn
}

async function saveEdit() {
  if (!editing.value || editBusy.value) return
  editAttempted.value = true
  if (editRecordsError.value || editHostnameError.value) return
  editBusy.value = true
  editError.value = ''
  try {
    const renamed = editHostname.value.trim().toLowerCase() !== editing.value.fqdn
    await api(`/api/hosts/${editing.value.id}`, { method: 'PATCH', body: JSON.stringify({ ...editForm.value, ...(renamed ? { name: editName.value } : {}), ttl: editForm.value.proxied ? 1 : editForm.value.ttl }) })
    toast.success(renamed ? 'Endereço atualizado. O registro antigo foi mantido na Cloudflare. Use o novo endereço nos scripts e roteadores.' : 'Configuração salva. Se mudou os tipos de IP, confira também os scripts de atualização.')
    editing.value = null
    await reload()
  } catch (e) {
    editError.value = errorMessage(e, 'Não foi possível atualizar o host.')
  } finally {
    editBusy.value = false
  }
}

function lastSeen(h: Host) {
  const t = Math.max(h.last_seen_v4 ? +new Date(h.last_seen_v4) : 0, h.last_seen_v6 ? +new Date(h.last_seen_v6) : 0)
  return t ? new Date(t).toISOString() : null
}
</script>

<template>
  <div>
    <PageHeader title="Hosts" description="Um host é um endereço, como casa.seudominio.com, que continua apontando para sua rede quando o IP muda.">
      <template #actions>
        <button v-if="connections.length" type="button" class="btn btn-primary gap-2" @click="openForm">
          <Plus :size="18" aria-hidden="true" />Novo host
        </button>
      </template>
    </PageHeader>
    <LoadError v-if="loadError" :message="loadError" :busy="loading" @retry="reload" />
    <template v-if="!loadError">

      <EmptyState
        v-if="!loading && !connections.length"
        class="surface"
        title="Antes de criar um host, conecte sua Cloudflare"
        description="Adicione uma conexão para escolher um domínio da sua conta Cloudflare."
      >
        <RouterLink to="/connections" class="btn btn-primary">Adicionar conexão</RouterLink>
      </EmptyState>

      <template v-else>
        <section class="surface" aria-label="Hosts cadastrados">
          <SkeletonRows v-if="loading" :rows="3" />
          <EmptyState v-else-if="!hosts.length" title="Nenhum host ainda" description="Adicione um endereço para sua rede. Depois, configure um dispositivo para mantê-lo atualizado.">
            <template #icon><Globe :size="28" /></template>
            <button type="button" class="btn btn-primary" @click="openForm">Novo host</button>
          </EmptyState>
          <ul v-else class="divide-y divide-base-300">
            <li v-for="h in hosts" :key="h.id" class="flex flex-wrap items-center gap-x-4 gap-y-3 p-4">
              <div class="min-w-0 flex-1 basis-56">
                <p class="font-data font-semibold break-all">{{ h.fqdn }}</p>
                <p class="text-sm text-base-content/65">
                  <span class="font-data">{{ h.last_ipv4 || h.last_ipv6 || 'IP ainda não recebido' }}</span>
                  · último contato: {{ lastSeen(h) ? relativeTime(lastSeen(h)) : 'ainda não recebido' }}
                </p>
                <p class="text-sm text-base-content/65">{{ h.proxied ? 'Pela Cloudflare' : 'Acesso direto' }} · cache: {{ h.ttl === 1 ? 'automático' : Number.isInteger(h.ttl) && h.ttl > 1 ? `${h.ttl} s` : 'não informado' }}</p>
              </div>
              <StatusBadge :status="h.status" />
              <div class="flex flex-wrap gap-1">
                <button type="button" class="btn btn-sm btn-ghost gap-1.5" @click="openEdit(h)"><Pencil :size="14" aria-hidden="true" />Editar host</button>
                <RouterLink :to="{ path: '/tokens', query: { host: h.id } }" class="btn btn-sm btn-ghost gap-1.5"><KeyRound :size="14" aria-hidden="true" />Gerar token</RouterLink>
                <button type="button" class="btn btn-sm btn-ghost gap-1.5" :disabled="syncing === h.id" @click="sync(h)">
                  <RefreshCw :size="14" :class="{ 'animate-spin': syncing === h.id }" aria-hidden="true" />Atualizar na Cloudflare
                </button>
                <button type="button" class="btn btn-sm btn-ghost text-error gap-1.5" @click="askRemove(h)"><Trash2 :size="14" aria-hidden="true" />Remover</button>
              </div>
            </li>
          </ul>
        </section>
      </template>

    </template>

    <FormModal
      :open="showForm"
      title="Novo host"
      size="large"
      description="Escolha o domínio e o nome do endereço que vai apontar para sua rede."
      submit-label="Criar host"
      busy-label="Criando…"
      :busy="busy"
      :error="error"
      :can-submit="!blocked"
      @submit="create"
      @cancel="showForm = false"
    >
      <div class="form-grid">
        <FormField v-slot="{ id, describedBy, invalid }" label="Conexão" :error="errs.connection" :submitted="attempted">
          <select :id="id" v-model="form.connection_id" class="select select-bordered w-full" required :aria-invalid="invalid" :aria-describedby="describedBy" @change="pickConnection">
            <option disabled value="">Escolha uma conexão…</option>
            <option v-for="c in connections" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </FormField>
        <FormField
          v-slot="{ id, describedBy, invalid }"
          label="Domínio"
          :hint="form.connection_id && !zones.length ? 'Nenhum domínio disponível. Confira o token da Cloudflare na página Conexões.' : undefined"
          :error="errs.zone" :submitted="attempted"
        >
          <select :id="id" v-model="form.zone_id" class="select select-bordered w-full" required :disabled="!zones.length" :aria-invalid="invalid" :aria-describedby="describedBy">
            <option disabled value="">Escolha um domínio…</option>
            <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
          </select>
        </FormField>
      </div>
      <FormField v-slot="{ id, describedBy, invalid }" label="Subdomínio" :hint="preview ? `Endereço final: ${preview}` : 'Digite a parte antes do domínio, como casa. Para usar o próprio domínio, digite @.'" :error="errs.name" :submitted="attempted">
        <input
          :id="id"
          v-model="form.name"
          class="input input-bordered w-full font-data"
          placeholder="Ex.: casa"
          maxlength="120"
          autocomplete="off"
          autocapitalize="none"
          spellcheck="false"
          required
          :aria-invalid="invalid"
          :aria-describedby="describedBy"
        />
      </FormField>
      <HostDnsOptions v-model:enable-ipv4="form.enable_a" v-model:enable-ipv6="form.enable_aaaa" v-model:proxied="form.proxied" v-model:ttl="form.ttl" :error="attempted ? errs.records : ''" />

      <p v-if="checking" class="flex items-center gap-2 text-sm text-base-content/70" role="status">
        <span class="loading loading-spinner loading-xs" />Verificando se {{ preview }} já existe…
      </p>
      <p v-else-if="checkError" class="text-sm text-warning" role="status">{{ checkError }}</p>
      <p v-else-if="check && !conflict" class="flex items-center gap-2 text-sm text-success" role="status">
        <CircleCheck :size="16" aria-hidden="true" />{{ check.fqdn }} está livre.
      </p>

      <div v-if="check?.in_app" class="flex items-start gap-3 rounded-btn border border-error/50 bg-error/10 p-3" role="alert">
        <OctagonX :size="20" class="mt-0.5 shrink-0 text-error" aria-hidden="true" />
        <p class="text-sm"><strong>{{ check.fqdn }}</strong> já é um host do HomeAlias. Escolha outro subdomínio.</p>
      </div>

      <div v-else-if="check?.records.length" class="space-y-3 rounded-btn border border-warning/60 bg-warning/10 p-3" role="alert">
        <div class="flex items-start gap-3">
          <TriangleAlert :size="20" class="mt-0.5 shrink-0 text-warning" aria-hidden="true" />
          <p class="text-sm"><strong>{{ check.fqdn }}</strong> já existe na Cloudflare:</p>
        </div>
        <ul class="space-y-1">
          <li v-for="r in check.records" :key="r.type" class="flex flex-wrap items-center gap-x-3 rounded-btn bg-base-100 px-3 py-1.5 text-sm">
            <span class="badge badge-sm badge-outline">{{ r.type }}</span>
            <span class="font-data">{{ r.content }}</span>
            <span class="text-xs text-base-content/65">TTL {{ r.ttl === 1 ? 'auto' : r.ttl }} · {{ r.proxied ? 'com proxy' : 'sem proxy' }}</span>
          </li>
        </ul>
        <label class="flex cursor-pointer items-start gap-3 text-sm">
          <input v-model="form.adopt_existing" type="checkbox" class="checkbox checkbox-warning checkbox-sm mt-0.5" />
          <span>
            <strong>Permitir que o HomeAlias atualize este registro.</strong>
            Na próxima atualização, este registro passará a apontar para o IP enviado pelo seu dispositivo. Para manter o registro atual, escolha outro nome.
          </span>
        </label>
      </div>
    </FormModal>

    <FormModal :open="!!editing" title="Editar host" size="large" :description="editing?.fqdn ?? ''" submit-label="Salvar alterações" busy-label="Salvando…" :busy="editBusy" :error="editError" @submit="saveEdit" @cancel="editing = null">
      <FormField v-slot="{ id, describedBy, invalid }" label="Endereço completo" :hint="`Use um endereço dentro de ${editing?.zone_name ?? 'seu domínio'}. Ao mudar o nome, o registro antigo fica na Cloudflare.`" :error="editHostnameError" :submitted="editAttempted">
        <input :id="id" v-model="editHostname" type="text" class="input input-bordered w-full font-data" placeholder="Ex.: casa.seudominio.com" maxlength="253" autocomplete="off" autocapitalize="none" spellcheck="false" required :aria-invalid="invalid" :aria-describedby="describedBy" />
      </FormField>
      <p v-if="editing && editHostname.trim().toLowerCase() !== editing.fqdn" class="text-sm text-warning" role="status">Seus tokens continuam funcionando, mas você precisa trocar o endereço nos scripts e roteadores. O nome antigo não será mais aceito nas atualizações.</p>
      <HostDnsOptions v-model:enable-ipv4="editForm.enable_a" v-model:enable-ipv6="editForm.enable_aaaa" v-model:proxied="editForm.proxied" v-model:ttl="editForm.ttl" :error="editAttempted ? editRecordsError : ''" />
      <p class="text-sm text-base-content/70">Ao desativar IPv4 ou IPv6, o HomeAlias deixa de atualizar esse tipo de IP. O registro já criado permanece na Cloudflare. Confira também os scripts dos seus dispositivos.</p>
    </FormModal>

    <ConfirmDialog
      :open="!!removing"
      :title="removing ? `Remover ${removing.fqdn}?` : 'Remover host'"
      confirm-label="Remover host"
      danger
      :busy="removeBusy"
      @confirm="confirmRemove"
      @cancel="removing = null"
    >
      <p>Os tokens ligados a este host deixam de funcionar para ele.</p>
      <label class="mt-4 flex items-start gap-3 cursor-pointer">
        <input v-model="alsoDeleteDns" type="checkbox" class="checkbox checkbox-sm mt-0.5" />
        <span>Apagar também o registro DNS na Cloudflare</span>
      </label>
    </ConfirmDialog>
  </div>
</template>
