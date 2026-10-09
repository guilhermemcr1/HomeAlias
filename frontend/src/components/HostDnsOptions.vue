<script setup lang="ts">
import FormField from './FormField.vue'
import { useId } from 'vue'

const enableA = defineModel<boolean>('enableIpv4', { required: true })
const enableAAAA = defineModel<boolean>('enableIpv6', { required: true })
const proxied = defineModel<boolean>('proxied', { required: true })
const ttl = defineModel<number>('ttl', { required: true })
defineProps<{ error?: string }>()
const ttlOptions = [1, 60, 120, 300, 600, 1800, 3600, 86400]
const recordsErrorId = useId()
</script>

<template>
  <fieldset class="flex flex-col gap-2" :aria-describedby="error ? recordsErrorId : undefined">
    <legend class="text-sm font-semibold mb-1.5">Tipos de IP para atualizar</legend>
    <div class="flex flex-wrap gap-x-6 gap-y-2">
      <label class="flex items-center gap-3 min-h-11 cursor-pointer"><input v-model="enableA" type="checkbox" class="checkbox checkbox-primary checkbox-sm" :aria-invalid="!!error" :aria-describedby="error ? recordsErrorId : undefined" />IPv4 (registro A)</label>
      <label class="flex items-center gap-3 min-h-11 cursor-pointer"><input v-model="enableAAAA" type="checkbox" class="checkbox checkbox-primary checkbox-sm" :aria-invalid="!!error" :aria-describedby="error ? recordsErrorId : undefined" />IPv6 (registro AAAA)</label>
    </div>
    <p v-if="error" :id="recordsErrorId" class="text-sm text-error" role="alert">{{ error }}</p>
  </fieldset>
  <div class="form-grid">
    <FormField v-slot="{ id, describedBy }" label="Como acessar este endereço" hint="O acesso direto aponta para seu IP. Com proxy, o tráfego de sites (HTTP/HTTPS) passa pela Cloudflare. A escolha vale para os tipos de IP selecionados.">
      <select :id="id" v-model="proxied" class="select select-bordered w-full" :aria-describedby="describedBy">
        <option :value="false">Direto para seu IP (somente DNS)</option>
        <option :value="true">Pela Cloudflare (com proxy)</option>
      </select>
    </FormField>
    <FormField v-slot="{ id, describedBy }" label="Tempo de cache (TTL)" :hint="proxied ? 'Com proxy, esse tempo é definido automaticamente pela Cloudflare.' : 'Por quanto tempo uma consulta DNS pode reutilizar o IP antes de buscar o registro novamente.'">
      <select :id="id" :value="proxied ? 1 : ttl" class="select select-bordered w-full" :disabled="proxied" :aria-describedby="describedBy" @change="ttl = Number(($event.target as HTMLSelectElement).value)">
        <option v-if="!ttlOptions.includes(ttl)" :value="ttl">{{ ttl }} segundos</option>
        <option v-for="value in ttlOptions" :key="value" :value="value">{{ value === 1 ? 'Automático' : `${value} segundos` }}</option>
      </select>
    </FormField>
  </div>
</template>
