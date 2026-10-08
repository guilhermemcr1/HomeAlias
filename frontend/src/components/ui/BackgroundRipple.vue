<script setup lang="ts">
// Grade de células que ondula a partir do clique (porta do BackgroundRippleEffect).
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const props = withDefaults(defineProps<{ cellSize?: number }>(), { cellSize: 56 })

const cols = ref(0)
const rows = ref(0)
const clicked = ref<{ row: number; col: number } | null>(null)
const rippleKey = ref(0)

function fit() {
  cols.value = Math.ceil(window.innerWidth / props.cellSize) + 1
  rows.value = Math.ceil(window.innerHeight / props.cellSize) + 1
}

function ripple(row: number, col: number) {
  clicked.value = { row, col }
  rippleKey.value++
}

onMounted(() => {
  fit()
  window.addEventListener('resize', fit)
  // Onda de abertura a partir do topo, no centro.
  ripple(0, Math.floor(cols.value / 2))
})
onBeforeUnmount(() => window.removeEventListener('resize', fit))

const cells = computed(() =>
  Array.from({ length: rows.value * cols.value }, (_, idx) => {
    const row = Math.floor(idx / cols.value)
    const col = idx % cols.value
    if (!clicked.value) return { idx, row, col, style: undefined }
    const d = Math.hypot(clicked.value.row - row, clicked.value.col - col)
    return { idx, row, col, style: { '--delay': `${d * 55}ms`, '--duration': `${200 + d * 80}ms` } }
  }),
)
</script>

<template>
  <div class="ripple-bg absolute inset-0 overflow-hidden" aria-hidden="true">
    <div
      :key="rippleKey"
      class="ripple-grid"
      :style="{ gridTemplateColumns: `repeat(${cols}, ${cellSize}px)`, gridAutoRows: `${cellSize}px` }"
    >
      <div
        v-for="c in cells"
        :key="c.idx"
        class="ripple-cell"
        :class="{ 'is-rippling': clicked }"
        :style="c.style"
        @click="ripple(c.row, c.col)"
      />
    </div>
  </div>
</template>
