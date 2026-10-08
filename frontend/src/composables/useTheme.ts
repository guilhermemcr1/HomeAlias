import { ref } from 'vue'

type ThemeName = 'homealias-dark' | 'homealias-light'

const theme = ref<ThemeName>(
  (document.documentElement.getAttribute('data-theme') as ThemeName) || 'homealias-dark',
)

export function useTheme() {
  function toggle() {
    theme.value = theme.value === 'homealias-dark' ? 'homealias-light' : 'homealias-dark'
    document.documentElement.setAttribute('data-theme', theme.value)
    try {
      localStorage.setItem('homealias-theme', theme.value)
    } catch {
      /* armazenamento indisponível: o tema vale só nesta sessão */
    }
  }
  return { theme, toggle }
}
