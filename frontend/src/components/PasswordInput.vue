<script setup lang="ts">
import { inject, ref, useAttrs, watch } from 'vue'
import { formValidationCycle } from '../composables/formValidation'
import { Eye, EyeOff } from 'lucide-vue-next'

defineOptions({ inheritAttrs: false })
defineProps<{ disabled?: boolean; secretLabel?: string; placeholder?: string }>()
const value = defineModel<string>({ required: true })
const attrs = useAttrs()
const visible = ref(false)
watch(value, (next) => { if (!next) visible.value = false })
watch(inject(formValidationCycle, ref(0)), () => { visible.value = false })
</script>

<template>
  <div class="relative w-full min-w-0">
    <input v-bind="$attrs" v-model="value" :type="visible ? 'text' : 'password'" :placeholder="placeholder || 'Digite sua senha'" class="pr-14" :disabled="disabled" />
    <div class="absolute inset-y-0 right-1 flex items-center">
    <button
      type="button"
      class="btn btn-ghost btn-sm btn-square !min-h-[44px] !min-w-[44px]"
      :aria-label="`${visible ? 'Ocultar' : 'Mostrar'} ${secretLabel || 'senha'}`"
      :aria-controls="typeof attrs.id === 'string' ? attrs.id : undefined"
      :aria-pressed="visible"
      :disabled="disabled"
      @click="visible = !visible"
    >
      <component :is="visible ? EyeOff : Eye" :size="18" aria-hidden="true" />
    </button>
    </div>
  </div>
</template>
