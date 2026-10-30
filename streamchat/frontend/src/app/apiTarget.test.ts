import { describe, expect, it } from 'vitest'
describe('API proxy target', () => {
  it('uses the Compose API service hostname by default', async () => {
    const { apiTarget } = await import('./apiTarget')
    expect(apiTarget).toBe('http://api:8080')
  })
})
