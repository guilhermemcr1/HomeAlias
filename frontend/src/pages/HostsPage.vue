<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { CircleCheck, Globe, KeyRound, OctagonX, Plus, RefreshCw, Trash2, TriangleAlert } from 'lucide-vue-next'
import { ApiError, api, errorMessage } from '../api/client'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import LoadError from '../components/LoadError.vue'
import EmptyState from '../components/EmptyState.vue'
import FormField from '../components/FormField.vue'
import FormModal from '../components/FormModal.vue'
import PageHeader from '../components/PageHeader.vue'
import SkeletonRows from '../components/SkeletonRows.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { relativeTime } from '../composables/format'
import { useToast } from '../composables/useToast'
import * as v from '../composables/validation'

type Zone = { id: string; name: string }
type Connection = { id: string; name: string; zones?: string | Zone[] | null }
type Host = {
  id: string
  fqdn: string
  status: string
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
  connection: v.required('Conexão', form.value.connection_id),
  zone: v.required('Zona', form.value.zone_id),
  name: v.hostName(form.value.name) || (zone.value ? v.fqdn(form.value.name, zone.value.name) : ''),
  records: form.value.enable_a || form.value.enable_aaaa ? '' : 'Ative ao menos um tipo de registro (A ou AAAA).',
}))
const preview = computed(() => (form.value.name && zone.value ? form.value.name.trim() === '@' ? zone.value.name : `${form.value.name.trim().toLowerCase()}.${zone.value.name}` : ''))

async function runCheck() {
  const seq = ++checkSeq
  check.value = null
  checkError.value = ''
  form.value.adopt_existing = false
  if (!zone.value || errs.value.name || !form.value.connection_id) {
    checking.value = false
    return
  }
  checking.value = true
  try {
    const res = await api<Check>('/api/hosts/check', {
      method: 'POST',
      body: JSON.stringify({
        connection_id: form.value.connection_id,
        zone_id: form.value.zone_id,
        zone_name: zone.value.name,
        name: form.value.name.trim().toLowerCase(),
      }),
    })
    if (seq === checkSeq) check.value = res
  } catch (e) {
    if (seq === checkSeq) checkError.value = errorMessage(e, 'Não foi possível verificar a Cloudflare.')
  } finally {
    if (seq === checkSeq) checking.value = false
  }
}

// Espera o usuário parar de digitar antes de consultar a Cloudflare.
watch(
  () => [form.value.connection_id, form.value.zone_id, form.value.name.trim().toLowerCase()],
  () => {
    clearTimeout(checkTimer)
    checking.value = !!(zone.value && !errs.value.name)
    checkTimer = setTimeout(runCheck, 500)
  },
)
onBeforeUnmount(() => clearTimeout(checkTimer))

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
      body: JSON.stringify({ ...form.value, name: form.value.name.trim().toLowerCase(), zone_name: zone.value.name }),
    })
    toast.success(`${preview.value} criado. Gere um token para começar a atualizar.`)
    showForm.value = false
    await reload()
  } catch (e) {
    // Alguém criou o registro entre a verificação e o envio: mostra a pergunta de novo.
    if (e instanceof ApiError && e.status === 409) await runCheck()
    error.value = e instanceof ApiError && e.status === 409 ? '' : errorMessage(e, 'Não foi possível criar o host.')
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

function lastSeen(h: Host) {
  const t = Math.max(h.last_seen_v4 ? +new Date(h.last_seen_v4) : 0, h.last_seen_v6 ? +new Date(h.last_seen_v6) : 0)
  return t ? new Date(t).toISOString() : null
}
</script>

<template>
  <div>
    <PageHeader title="Hosts" description="Cada host é um nome (como casa.seudominio.com) que acompanha o seu IP.">
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
        description="O host precisa de uma zona DNS, e as zonas vêm da conexão."
      >
        <RouterLink to="/connections" class="btn btn-primary">Adicionar conexão</RouterLink>
      </EmptyState>

      <template v-else>
        <section class="surface" aria-label="Hosts cadastrados">
          <SkeletonRows v-if="loading" :rows="3" />
          <EmptyState v-else-if="!hosts.length" title="Nenhum host ainda" description="Crie o primeiro para poder gerar um token.">
            <template #icon><Globe :size="28" /></template>
            <button type="button" class="btn btn-primary" @click="openForm">Novo host</button>
          </EmptyState>
          <ul v-else class="divide-y divide-base-300">
            <li v-for="h in hosts" :key="h.id" class="flex flex-wrap items-center gap-x-4 gap-y-3 p-4">
              <div class="min-w-0 flex-1 basis-56">
                <p class="font-data font-semibold break-all">{{ h.fqdn }}</p>
                <p class="text-sm text-base-content/65">
                  <span class="font-data">{{ h.last_ipv4 || h.last_ipv6 || 'sem IP' }}</span>
                  · contato {{ relativeTime(lastSeen(h)) }}
                </p>
              </div>
              <StatusBadge :status="h.status" />
              <div class="flex flex-wrap gap-1">
                <RouterLink :to="{ path: '/tokens', query: { host: h.id } }" class="btn btn-sm btn-ghost gap-1.5"><KeyRound :size="14" aria-hidden="true" />Token</RouterLink>
                <button type="button" class="btn btn-sm btn-ghost gap-1.5" :disabled="syncing === h.id" @click="sync(h)">
                  <RefreshCw :size="14" :class="{ 'animate-spin': syncing === h.id }" aria-hidden="true" />Sincronizar
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
      description="O host é o nome que vai acompanhar o seu IP."
      submit-label="Criar host"
      busy-label="Criando…"
      :busy="busy"
      :error="error"
      :can-submit="!blocked"
      @submit="create"
      @cancel="showForm = false"
    >
      <FormField v-slot="{ id, describedBy, invalid }" label="Conexão" :error="attempted ? errs.connection : ''">
        <select :id="id" v-model="form.connection_id" class="select select-bordered w-full" required :aria-invalid="invalid" :aria-describedby="describedBy" @change="pickConnection">
          <option disabled value="">Escolha uma conexão…</option>
          <option v-for="c in connections" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
      </FormField>
      <FormField
        v-slot="{ id, describedBy, invalid }"
        label="Zona (domínio)"
        :hint="form.connection_id && !zones.length ? 'Esta conexão não listou zonas. Teste-a em Conexões.' : undefined"
        :error="attempted ? errs.zone : ''"
      >
        <select :id="id" v-model="form.zone_id" class="select select-bordered w-full" required :disabled="!zones.length" :aria-invalid="invalid" :aria-describedby="describedBy">
          <option disabled value="">Escolha uma zona…</option>
          <option v-for="z in zones" :key="z.id" :value="z.id">{{ z.name }}</option>
        </select>
      </FormField>
      <FormField v-slot="{ id, describedBy, invalid }" label="Subdomínio" :hint="preview ? `Endereço final: ${preview}` : 'Só a parte antes do domínio.'" :error="attempted ? errs.name : ''">
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
      <fieldset class="flex flex-col gap-2">
        <legend class="text-sm font-semibold mb-1.5">Registros e opções</legend>
        <label class="flex items-center gap-3 min-h-8 cursor-pointer"><input v-model="form.enable_a" type="checkbox" class="checkbox checkbox-primary checkbox-sm" />IPv4 (registro A)</label>
        <label class="flex items-center gap-3 min-h-8 cursor-pointer"><input v-model="form.enable_aaaa" type="checkbox" class="checkbox checkbox-primary checkbox-sm" />IPv6 (registro AAAA)</label>
        <p v-if="attempted && errs.records" class="text-sm text-error" role="alert">{{ errs.records }}</p>
      </fieldset>

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
            <strong>Assumir e sincronizar pelo app.</strong>
            O HomeAlias passa a gerenciar este registro e o substitui pelo seu IP atual na próxima atualização. Para manter como está, escolha outro subdomínio.
          </span>
        </label>
      </div>
    </FormModal>

    <ConfirmDialog
      :open="!!removing"
      :title="`Remover ${removing?.fqdn}?`"
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
