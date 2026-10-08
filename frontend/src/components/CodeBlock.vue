<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { Download, Eye, EyeOff } from 'lucide-vue-next'
import CopyButton from './CopyButton.vue'

// `mask` é o trecho sensível (ex.: o token). Fica oculto com ••••; aparece ao passar o mouse,
// ao focar, ao copiar (por instantes) ou pelo botão do olho. Copiar e baixar usam sempre o valor real.
const props = defineProps<{ code: string; caption?: string; filename?: string; scroll?: boolean; mask?: string }>()

const DOTS = '••••••••••••'
const hover = ref(false)
const pinned = ref(false)
const flash = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

const revealed = computed(() => !props.mask || hover.value || pinned.value || flash.value)
const shown = computed(() => (revealed.value || !props.mask ? props.code : props.code.split(props.mask).join(DOTS)))

function onCopied() {
  flash.value = true
  clearTimeout(timer)
  timer = setTimeout(() => (flash.value = false), 2000)
}
onBeforeUnmount(() => clearTimeout(timer))

// Entrega o conteúdo real como arquivo, sem nova ida ao servidor.
function download() {
  const url = URL.createObjectURL(new Blob([props.code], { type: 'text/plain;charset=utf-8' }))
  const a = document.createElement('a')
  a.href = url
  a.download = props.filename ?? 'homealias.txt'
  document.body.appendChild(a)
  a.click()
  a.remove()
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}
</script>

<template>
  <div class="overflow-hidden rounded-btn border border-base-300 bg-base-200">
    <div class="flex flex-wrap items-center justify-between gap-2 border-b border-base-300 py-1 pl-3 pr-1">
      <span class="min-w-0 flex-1 basis-32 text-xs font-semibold text-base-content/70">{{ caption }}</span>
      <div class="flex items-center">
        <button
          v-if="mask"
          type="button"
          class="btn btn-sm btn-ghost btn-square"
          :aria-label="pinned ? 'Ocultar dado sensível' : 'Mostrar dado sensível'"
          :aria-pressed="pinned"
          @click="pinned = !pinned"
        >
          <component :is="revealed ? EyeOff : Eye" :size="16" aria-hidden="true" />
        </button>
        <button v-if="filename" type="button" class="btn btn-sm btn-ghost gap-1.5" @click="download">
          <Download :size="16" aria-hidden="true" />Baixar
        </button>
        <CopyButton :text="code" @copied="onCopied" />
      </div>
    </div>
    <pre
      class="font-data overflow-auto p-3 transition-[filter] duration-200"
      :class="scroll ? 'max-h-80 whitespace-pre text-xs' : 'whitespace-pre-wrap break-all'"
      tabindex="0"
      @mouseenter="hover = true"
      @mouseleave="hover = false"
      @focus="hover = true"
      @blur="hover = false"
    ><code>{{ shown }}</code></pre>
  </div>
</template>
