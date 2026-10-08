import { describe, expect, it } from 'vitest'
import { relativeTime, statusInfo } from '../src/composables/format'

describe('statusInfo', () => {
  it('maps known statuses to label and class', () => {
    expect(statusInfo('online')).toEqual({ label: 'Online', cls: 'badge-success' })
    expect(statusInfo('warning').cls).toBe('badge-warning')
    expect(statusInfo('offline').cls).toBe('badge-error')
    expect(statusInfo('never_seen').cls).toBe('badge-ghost')
  })

  it('falls back to the raw status for unknown values', () => {
    expect(statusInfo('xyz')).toEqual({ label: 'xyz', cls: 'badge-ghost' })
  })
})

describe('relativeTime', () => {
  it('handles missing and recent timestamps', () => {
    expect(relativeTime(null)).toBe('nunca')
    expect(relativeTime(new Date().toISOString())).toBe('agora há pouco')
  })
})
