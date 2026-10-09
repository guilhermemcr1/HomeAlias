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
    <legend class="text-sm font-semibold mb-1.5">Registros DNS</legend>
    <label class="flex items-center gap-3 min-h-11 cursor-pointer"><input v-model="enableA" type="checkbox" class="checkbox checkbox-primary checkbox-sm" :aria-invalid="!!error" :aria-describedby="error ? recordsErrorId : undefined" />IPv4 (registro A)</label>
    <label class="flex items-center gap-3 min-h-11 cursor-pointer"><input v-model="enableAAAA" type="checkbox" class="checkbox checkbox-primary checkbox-sm" :aria-invalid="!!error" :aria-describedby="error ? recordsErrorId : undefined" />IPv6 (registro AAAA)</label>
    <p v-if="error" :id="recordsErrorId" class="text-sm text-error" role="alert">{{ error }}</p>
  </fieldset>
  <FormField v-slot="{ id, describedBy }" label="Proxy Cloudflare" hint="A escolha vale para os registros habilitados deste host. Só DNS conecta diretamente ao seu IP; com proxy, o tráfego HTTP/HTTPS passa pela Cloudflare.">
    <select :id="id" v-model="proxied" class="select select-bordered w-full" :aria-describedby="describedBy">
      <option :value="false">Só DNS (sem proxy)</option>
      <option :value="true">Com proxy Cloudflare</option>
    </select>
  </FormField>
  <FormField v-slot="{ id, describedBy }" label="TTL" :hint="proxied ? 'Com proxy, a Cloudflare usa TTL automático.' : 'Tempo de cache do registro DNS.'">
    <select :id="id" :value="proxied ? 1 : ttl" class="select select-bordered w-full" :disabled="proxied" :aria-describedby="describedBy" @change="ttl = Number(($event.target as HTMLSelectElement).value)">
      <option v-if="!ttlOptions.includes(ttl)" :value="ttl">{{ ttl }} segundos</option>
      <option v-for="value in ttlOptions" :key="value" :value="value">{{ value === 1 ? 'Automático' : `${value} segundos` }}</option>
    </select>
  </FormField>
</template>
