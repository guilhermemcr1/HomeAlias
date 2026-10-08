<script setup lang="ts">
import { computed } from 'vue'
import { CircleCheck, CircleX, Hourglass, TriangleAlert } from 'lucide-vue-next'
import { statusInfo } from '../composables/format'

const props = defineProps<{ status: string }>()
const info = computed(() => statusInfo(props.status))
const icon = computed(() => {
  switch (props.status) {
    case 'online':
    case 'valid':
      return CircleCheck
    case 'warning':
      return TriangleAlert
    case 'offline':
    case 'invalid':
      return CircleX
    default:
      return Hourglass
  }
})
</script>

<template>
  <span class="badge badge-outline gap-1.5 h-7 px-2.5 font-semibold whitespace-nowrap" :class="info.cls">
    <component :is="icon" :size="14" :stroke-width="2.25" aria-hidden="true" />
    {{ info.label }}
  </span>
</template>
