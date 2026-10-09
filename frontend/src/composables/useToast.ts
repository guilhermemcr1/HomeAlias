import { ref } from 'vue'

export type Toast = { id: number; kind: 'success' | 'error' | 'warning' | 'info'; text: string }

const toasts = ref<Toast[]>([])
let seq = 0

function push(kind: Toast['kind'], text: string) {
  const id = ++seq
  toasts.value.push({ id, kind, text })
  setTimeout(() => dismiss(id), kind === 'error' || kind === 'warning' ? 7000 : 4000)
}

function dismiss(id: number) {
  toasts.value = toasts.value.filter((t) => t.id !== id)
}

export function useToast() {
  return {
    toasts,
    dismiss,
    success: (t: string) => push('success', t),
    error: (t: string) => push('error', t),
    warning: (t: string) => push('warning', t),
    info: (t: string) => push('info', t),
  }
}
