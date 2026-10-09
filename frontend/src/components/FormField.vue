<script setup lang="ts">
import { computed, inject, ref, useId, watch } from 'vue'
import { formValidationCycle } from '../composables/formValidation'

const props = withDefaults(defineProps<{ label: string; hint?: string; error?: string; submitted?: boolean }>(), { submitted: true })
const id = useId()
const touched = ref(false)
const cycle = inject(formValidationCycle, ref(0))
watch(cycle, () => { touched.value = false })
const visibleError = computed(() => props.error && (props.submitted || touched.value) ? props.error : '')
const describedBy = computed(() => [visibleError.value ? `${id}-error` : '', props.hint ? `${id}-hint` : ''].filter(Boolean).join(' ') || undefined)

function onBlur(event: FocusEvent) {
  if (!(event.relatedTarget instanceof Node) || !(event.currentTarget as HTMLElement).contains(event.relatedTarget)) touched.value = true
}
</script>

<template>
  <div class="min-w-0 flex flex-col gap-1.5" @focusout="onBlur">
    <label :for="id" class="text-sm font-semibold">{{ label }}</label>
    <slot :id="id" :described-by="describedBy" :invalid="!!visibleError" />
    <p v-if="visibleError" :id="`${id}-error`" class="text-sm text-error" role="alert">{{ visibleError }}</p>
    <p v-if="hint" :id="`${id}-hint`" class="text-sm text-base-content/70">{{ hint }}</p>
  </div>
</template>
