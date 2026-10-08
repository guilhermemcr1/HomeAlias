<script setup lang="ts">
import { nextTick, onMounted, ref, useId, watch } from 'vue'
import { focusFirstInvalid } from '../composables/formFocus'
import { X } from 'lucide-vue-next'

const props = withDefaults(defineProps<{
  open: boolean
  title: string
  description?: string
  submitLabel: string
  busyLabel?: string
  busy?: boolean
  error?: string
  canSubmit?: boolean
}>(), { canSubmit: true })
const emit = defineEmits<{ submit: []; cancel: [] }>()
const dialog = ref<HTMLDialogElement | null>(null)
const errorBox = ref<HTMLElement | null>(null)
const titleId = useId()

function syncDialog() {
  if (props.open && !dialog.value?.open) dialog.value?.showModal()
  if (!props.open && dialog.value?.open) dialog.value.close()
}
onMounted(syncDialog)
watch(() => props.open, syncDialog, { flush: 'post' })

// Erro do servidor entra numa região role=alert; garante que ela esteja visível.
watch(
  () => props.error,
  async (e) => {
    if (!e) return
    await nextTick()
    errorBox.value?.scrollIntoView({ block: 'nearest' })
  },
)

async function submit(event: Event) {
  if (props.busy || !props.canSubmit) return
  emit('submit')
  await focusFirstInvalid(event.target as HTMLFormElement)
}

function close() {
  if (!props.busy) emit('cancel')
}
</script>

<template>
  <dialog ref="dialog" class="modal modal-bottom sm:modal-middle" :aria-labelledby="titleId" @cancel.prevent="close">
    <div class="modal-box w-full max-w-lg">
      <div class="flex items-start justify-between gap-4">
        <div class="min-w-0">
          <h3 :id="titleId" class="font-display text-lg font-bold break-words">{{ title }}</h3>
          <p v-if="description" class="mt-1 text-sm text-base-content/70">{{ description }}</p>
        </div>
        <button type="button" class="btn btn-ghost btn-sm btn-square -mr-2 -mt-1" aria-label="Fechar" :disabled="busy" @click="close">
          <X :size="18" aria-hidden="true" />
        </button>
      </div>
      <form class="mt-5 flex flex-col gap-4" novalidate :aria-busy="busy" @submit.prevent="submit">
        <fieldset class="min-w-0 flex flex-col gap-4" :disabled="busy"><slot /></fieldset>
        <p v-if="error" ref="errorBox" class="rounded-btn border border-error/50 bg-error/10 px-3 py-2.5 text-sm break-words" role="alert">{{ error }}</p>
        <div class="modal-action mt-1">
          <button type="button" class="btn btn-ghost" :disabled="busy" @click="close">Cancelar</button>
          <button type="submit" class="btn btn-primary" :disabled="busy || !canSubmit">
            <span v-if="busy" class="loading loading-spinner loading-sm" />
            {{ busy && busyLabel ? busyLabel : submitLabel }}
          </button>
        </div>
      </form>
    </div>
    <div class="modal-backdrop"><button type="button" tabindex="-1" aria-label="Fechar" @click="close" /></div>
  </dialog>
</template>
