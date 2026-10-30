import { describe, expect, it } from 'vitest'
import { validateRegistration } from './validation'

describe('registration validation', () => {
  it('accepts valid registration fields', () => {
    expect(validateRegistration('casey_1', 'correct horse battery staple', 'Casey')).toBeNull()
  })

  it('rejects invalid usernames and weak passwords', () => {
    expect(validateRegistration('x', 'correct horse battery staple', '')).toMatch(/Username/)
    expect(validateRegistration('casey', 'short', '')).toMatch(/Password/)
  })

  it('rejects oversized display names', () => {
    expect(validateRegistration('casey', 'correct horse battery staple', 'x'.repeat(81))).toMatch(/Display name/)
  })
})
