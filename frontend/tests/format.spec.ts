import { expect, it } from 'vitest'
import { dateTime, relativeTime } from '../src/composables/format'
it('handles absent and invalid dates without crashing lists', () => {
  expect(relativeTime(null)).toBe('nunca')
  expect(relativeTime('invalid')).toBe('data indisponível')
  expect(dateTime('invalid')).toBe('—')
  expect(dateTime(null)).toBe('—')
})
