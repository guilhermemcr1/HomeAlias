<script setup lang="ts">
import { onMounted, ref, useId, watch } from 'vue'

const props = defineProps<{
  open: boolean
  title: string
  confirmLabel: string
  danger?: boolean
  busy?: boolean
}>()
const emit = defineEmits<{ confirm: []; cancel: [] }>()
const titleId = useId()
function close() {
  if (!props.busy) emit('cancel')
}
const dialog = ref<HTMLDialogElement | null>(null)

function syncDialog() {
  if (props.open && !dialog.value?.open) dialog.value?.showModal()
  if (!props.open && dialog.value?.open) dialog.value.close()
}
onMounted(syncDialog)
watch(() => props.open, syncDialog, { flush: 'post' })
</script>

<template>
  <dialog ref="dialog" class="modal" :aria-labelledby="titleId" :aria-busy="busy" @cancel.prevent="close">
    <div class="modal-box">
      <h3 :id="titleId" class="font-display text-lg font-bold">{{ title }}</h3>
      <div class="py-3 text-base-content/80"><slot /></div>
      <div class="modal-action">
        <button type="button" class="btn btn-ghost" :disabled="busy" @click="close">Cancelar</button>
        <button
          type="button"
          class="btn"
          :class="danger ? 'btn-error' : 'btn-primary'"
          :disabled="busy"
          @click="!busy && emit('confirm')"
        >
          <span v-if="busy" class="loading loading-spinner loading-xs" />
          {{ confirmLabel }}
        </button>
      </div>
    </div>
    <div class="modal-backdrop"><button type="button" tabindex="-1" aria-label="Fechar" :disabled="busy" @click="close" /></div>
  </dialog>
</template>
