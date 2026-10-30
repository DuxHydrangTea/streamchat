import { describe, expect, it } from 'vitest'
import { videoURL } from './Chat'

describe('videoURL', () => {
  it('recognizes a standalone HTTPS URL', () => {
    expect(videoURL(' https://youtu.be/example?t=4 ')).toBe('https://youtu.be/example?t=4')
  })

  it('does not treat plain text or embedded links as automatic downloads', () => {
    expect(videoURL('hello https://youtube.com/watch?v=1')).toBeUndefined()
    expect(videoURL('hello world')).toBeUndefined()
    expect(videoURL('http://youtube.com/watch?v=1')).toBeUndefined()
  })
})
