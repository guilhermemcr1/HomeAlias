<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useToast } from '../composables/useToast'
import { Check, Copy } from 'lucide-vue-next'

// Campo clicável: clicar copia o valor. Se `secret`, mostra ••••• e revela ao passar o mouse,
// focar ou copiar (por instantes).
const props = defineProps<{ label: string; value: string; hint?: string; secret?: boolean }>()
const hover = ref(false)
const copied = ref(false)
const toast = useToast()
let timer: ReturnType<typeof setTimeout> | undefined

const revealed = computed(() => !props.secret || hover.value || copied.value)

async function copy() {
  try {
    await navigator.clipboard.writeText(props.value)
  } catch {
    const focused = document.activeElement as HTMLElement | null
    const ta = document.createElement('textarea')
    ta.value = props.value
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    try {
      ta.select()
      if (!document.execCommand('copy')) throw new Error('clipboard unavailable')
    } catch {
      toast.error('Não foi possível copiar. Selecione o texto e copie manualmente.')
      return
    } finally {
      ta.remove()
      focused?.focus({ preventScroll: true })
    }
  }
  copied.value = true
  clearTimeout(timer)
  timer = setTimeout(() => (copied.value = false), 1800)
}
onBeforeUnmount(() => clearTimeout(timer))
</script>

<template>
  <button
    type="button"
    class="group flex min-w-0 w-full items-center gap-3 rounded-btn border border-base-300 bg-base-100 px-3 py-2 text-left transition-colors duration-150 hover:border-primary/50 hover:bg-base-200 active:scale-[0.99]"
    :title="hint"
    :aria-label="`Copiar ${label}`"
    @click="copy"
    @mouseenter="hover = true"
    @mouseleave="hover = false"
    @focus="hover = true"
    @blur="hover = false"
  >
    <span class="min-w-0 flex-1">
      <span class="block text-xs text-base-content/70">{{ label }}</span>
      <span class="font-data block truncate text-sm font-semibold">{{ revealed ? value : '••••••••••••' }}</span>
    </span>
    <span class="shrink-0 text-base-content/55 transition-colors group-hover:text-primary" :class="copied && '!text-success'" aria-live="polite">
      <component :is="copied ? Check : Copy" :size="16" aria-hidden="true" />
      <span class="sr-only">{{ copied ? 'Copiado' : '' }}</span>
    </span>
  </button>
</template>
