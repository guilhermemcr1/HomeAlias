<script setup lang="ts">
// Texto preenchido por linhas animadas desenhadas em canvas (porta do CanvasText).
import { onBeforeUnmount, onMounted, ref } from 'vue'

const props = withDefaults(
  defineProps<{
    text: string
    colors?: string[]
    animationDuration?: number
    lineWidth?: number
    lineGap?: number
    curveIntensity?: number
  }>(),
  {
    colors: () => ['#ff6b6b', '#4ecdc4', '#45b7d1', '#96ceb4', '#ffeaa7', '#dfe6e9'],
    animationDuration: 5,
    lineWidth: 1.5,
    lineGap: 10,
    curveIntensity: 60,
  },
)

const canvas = ref<HTMLCanvasElement | null>(null)
const sizer = ref<HTMLSpanElement | null>(null)
const bgProbe = ref<HTMLSpanElement | null>(null)
const size = ref({ width: 0, height: 0 })

let frame = 0
let font = ''
let bgColor = '#3b5bdb'
let resizeObserver: ResizeObserver | null = null
let themeObserver: MutationObserver | null = null
const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)')

function measure() {
  const el = sizer.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const cs = getComputedStyle(el)
  font = `${cs.fontStyle} ${cs.fontWeight} ${cs.fontSize} ${cs.fontFamily}`
  size.value = { width: Math.ceil(rect.width) || 200, height: Math.ceil(rect.height) || 48 }
  if (bgProbe.value) bgColor = getComputedStyle(bgProbe.value).backgroundColor
  start()
}

function start() {
  cancelAnimationFrame(frame)
  const el = canvas.value
  const ctx = el?.getContext('2d', { alpha: true })
  if (!el || !ctx || !font) return
  const { width, height } = size.value
  const dpr = window.devicePixelRatio || 1
  el.width = width * dpr
  el.height = height * dpr

  ctx.font = font
  const m = ctx.measureText(props.text)
  const baselineY = (height + m.actualBoundingBoxAscent - m.actualBoundingBoxDescent) / 2
  const lines = Math.floor(height / props.lineGap) + 10
  const t0 = performance.now()

  const draw = (now: number) => {
    const phase = (reduceMotion.matches ? 0.8 : ((now - t0) / 1000 / props.animationDuration) * Math.PI * 2)
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    ctx.clearRect(0, 0, width, height)
    ctx.globalCompositeOperation = 'source-over'
    ctx.font = font
    ctx.textBaseline = 'alphabetic'
    ctx.fillStyle = '#000'
    ctx.fillText(props.text, 0, baselineY)
    ctx.globalCompositeOperation = 'source-in'
    ctx.fillStyle = bgColor
    ctx.fillRect(0, 0, width, height)
    ctx.globalCompositeOperation = 'source-atop'
    ctx.lineWidth = props.lineWidth
    const c1 = Math.sin(phase) * props.curveIntensity
    const c2 = Math.sin(phase + 0.5) * props.curveIntensity * 0.6
    for (let i = 0; i < lines; i++) {
      const y = i * props.lineGap
      ctx.strokeStyle = props.colors[i % props.colors.length]
      ctx.beginPath()
      ctx.moveTo(0, y)
      ctx.bezierCurveTo(width * 0.33, y + c1, width * 0.66, y + c2, width, y)
      ctx.stroke()
    }
    if (!reduceMotion.matches) frame = requestAnimationFrame(draw)
  }
  frame = requestAnimationFrame(draw)
}

onMounted(() => {
  measure()
  // A fonte variável chega depois da primeira medição.
  void document.fonts?.ready.then(measure)
  if (sizer.value) {
    resizeObserver = new ResizeObserver(measure)
    resizeObserver.observe(sizer.value)
  }
  themeObserver = new MutationObserver(measure)
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme', 'class'] })
})

onBeforeUnmount(() => {
  cancelAnimationFrame(frame)
  resizeObserver?.disconnect()
  themeObserver?.disconnect()
})
</script>

<template>
  <span class="relative inline-block">
    <span ref="bgProbe" class="pointer-events-none absolute h-0 w-0 bg-primary opacity-0" aria-hidden="true" />
    <span ref="sizer" class="invisible inline-block" aria-hidden="true">{{ text }}</span>
    <canvas
      ref="canvas"
      class="pointer-events-none absolute left-0 top-0"
      :style="{ width: size.width ? size.width + 'px' : 'auto', height: size.height ? size.height + 'px' : 'auto' }"
      role="img"
      :aria-label="text"
    />
  </span>
</template>
