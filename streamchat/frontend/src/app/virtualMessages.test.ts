import { describe, expect, it } from 'vitest'
import { preserveScrollAnchor, visibleRange } from './virtualMessages'

describe('virtual message helpers', () => {
  it('preserves the viewed pixel after prepending history', () => {
    expect(preserveScrollAnchor(220, 1000, 1600)).toBe(820)
  })
  it('clamps overscan to the list bounds', () => {
    expect(visibleRange(1, 4, 3, 5)).toEqual({ start: 0, end: 5 })
    expect(visibleRange(10, 20, 3, 30)).toEqual({ start: 7, end: 23 })
  })
})
