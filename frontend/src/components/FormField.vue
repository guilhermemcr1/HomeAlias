<script setup lang="ts">
import { useId } from 'vue'

const props = defineProps<{ label: string; hint?: string; error?: string }>()
const id = useId()
</script>

<template>
  <div class="min-w-0 flex flex-col gap-1.5">
    <label :for="id" class="text-sm font-semibold">{{ label }}</label>
    <slot :id="id" :described-by="props.error || props.hint ? `${id}-d` : undefined" :invalid="!!error" />
    <p v-if="error" :id="`${id}-d`" class="text-sm text-error" role="alert">{{ error }}</p>
    <p v-else-if="hint" :id="`${id}-d`" class="text-sm text-base-content/65">{{ hint }}</p>
  </div>
</template>
