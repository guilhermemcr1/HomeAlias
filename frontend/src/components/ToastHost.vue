<script setup lang="ts">
import { CircleCheck, CircleX, Info, TriangleAlert, X } from 'lucide-vue-next'
import { useToast } from '../composables/useToast'

const { toasts, dismiss } = useToast()
const icons = { success: CircleCheck, error: CircleX, warning: TriangleAlert, info: Info }
const variants = { success: 'alert-success', error: 'alert-error', warning: 'alert-warning', info: 'alert-info' }
</script>

<template>
  <div class="toast toast-end toast-bottom z-50 w-full min-w-0 max-w-md whitespace-normal pointer-events-none">
    <TransitionGroup name="toast" tag="div" class="relative flex w-full max-h-[calc(100dvh-2rem)] animate-none flex-col gap-3 overflow-y-auto pointer-events-auto">
    <div
      v-for="t in toasts"
      :key="t.id"
      class="alert grid-cols-[auto_minmax(0,1fr)_auto] items-start gap-3 p-3 text-left shadow-lg"
      :class="variants[t.kind]"
      :role="t.kind === 'error' || t.kind === 'warning' ? 'alert' : 'status'"
      aria-atomic="true"
    >
      <component :is="icons[t.kind]" :size="20" class="mt-3 shrink-0" aria-hidden="true" />
      <p class="min-w-0 py-3 text-sm leading-5 [overflow-wrap:anywhere]">{{ t.text }}</p>
      <button type="button" class="btn btn-ghost btn-sm btn-square !min-h-[44px] !min-w-[44px] text-inherit" aria-label="Dispensar" @click="dismiss(t.id)">
        <X :size="18" aria-hidden="true" />
      </button>
    </div>
    </TransitionGroup>
  </div>
</template>
