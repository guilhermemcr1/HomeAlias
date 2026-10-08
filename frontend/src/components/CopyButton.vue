<script setup lang="ts">
import { onUnmounted, ref } from 'vue'
import { useToast } from '../composables/useToast'
import { Check, Copy } from 'lucide-vue-next'

const emit = defineEmits<{ copied: [] }>()
const props = defineProps<{ text: string; label?: string; iconOnly?: boolean }>()
const copied = ref(false)
const copying = ref(false)
const toast = useToast()
let timer: ReturnType<typeof setTimeout> | undefined
onUnmounted(() => clearTimeout(timer))

async function copy() {
  if (copying.value) return
  copying.value = true
  const focused = document.activeElement as HTMLElement | null
  try {
    try {
      await navigator.clipboard.writeText(props.text)
    } catch {
      const ta = document.createElement('textarea')
      ta.value = props.text
      ta.style.position = 'fixed'
      ta.style.opacity = '0'
      // A modal makes the rest of the document inert; keep the fallback inside it.
      ;(focused?.closest('dialog') ?? document.body).appendChild(ta)
      try {
        ta.select()
        if (!document.execCommand('copy')) throw new Error('clipboard unavailable')
      } finally {
        ta.remove()
        focused?.focus({ preventScroll: true })
      }
    }
    copied.value = true
    emit('copied')
    clearTimeout(timer)
    timer = setTimeout(() => (copied.value = false), 1800)
  } catch {
    toast.error('Não foi possível copiar. Selecione o texto e copie manualmente.')
  } finally {
    copying.value = false
  }
}

</script>

<template>
  <button type="button" class="btn btn-sm btn-ghost gap-1.5" :class="iconOnly && 'btn-square'" :aria-label="iconOnly ? (copied ? 'Copiado' : (label ?? 'Copiar')) : undefined" :title="iconOnly ? (label ?? 'Copiar') : undefined" :disabled="copying" @click="copy">
    <component :is="copied ? Check : Copy" :size="16" aria-hidden="true" />
    <span :class="iconOnly && 'sr-only'" aria-live="polite">{{ copied ? 'Copiado' : (label ?? 'Copiar') }}</span>
  </button>
</template>
