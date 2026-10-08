<script setup lang="ts">
import { CircleCheck, CircleX, Info, X } from 'lucide-vue-next'
import { useToast } from '../composables/useToast'

const { toasts, dismiss } = useToast()
const icons = { success: CircleCheck, error: CircleX, info: Info }
const tone = { success: 'text-success', error: 'text-error', info: 'text-info' }
</script>

<template>
  <div class="fixed bottom-4 right-4 left-4 sm:left-auto z-50 flex flex-col gap-2 sm:w-96" aria-live="polite">
    <TransitionGroup name="toast">
    <div
      v-for="t in toasts"
      :key="t.id"
      class="flex items-start gap-3 rounded-box border border-base-300 bg-base-100 p-3 shadow-lg"
      :role="t.kind === 'error' ? 'alert' : 'status'"
    >
      <component :is="icons[t.kind]" :size="20" class="mt-0.5 shrink-0" :class="tone[t.kind]" aria-hidden="true" />
      <p class="flex-1 text-sm">{{ t.text }}</p>
      <button type="button" class="btn btn-ghost btn-xs btn-square" aria-label="Dispensar" @click="dismiss(t.id)">
        <X :size="14" />
      </button>
    </div>
    </TransitionGroup>
  </div>
</template>
