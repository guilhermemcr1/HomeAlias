import { nextTick } from 'vue'

export async function focusFirstInvalid(form: HTMLFormElement | null) {
  await nextTick()
  form?.querySelector<HTMLElement>('[aria-invalid="true"]:not(:disabled)')?.focus()
}
