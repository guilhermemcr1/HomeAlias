import { ref } from 'vue'

export type Toast = { id: number; kind: 'success' | 'error' | 'info'; text: string }

const toasts = ref<Toast[]>([])
let seq = 0

function push(kind: Toast['kind'], text: string) {
  const id = ++seq
  toasts.value.push({ id, kind, text })
  setTimeout(() => dismiss(id), kind === 'error' ? 7000 : 4000)
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
    info: (t: string) => push('info', t),
  }
}
