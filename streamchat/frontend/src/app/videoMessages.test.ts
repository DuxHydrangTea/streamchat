import { describe, expect, it } from 'vitest'
import { mediaJobProgress, videoURL } from './Chat'

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

describe('mediaJobProgress', () => {
  it.each([
    ['queued', 'Queued for preparation'],
    ['fetching_metadata', 'Reading video information'],
    ['downloading', 'Downloading video'],
    ['processing', 'Preparing video'],
    ['uploading', 'Saving video'],
  ])('describes the %s state', (status, title) => {
    expect(mediaJobProgress(status)?.title).toBe(title)
    expect(mediaJobProgress(status)?.detail).toBeTruthy()
  })

  it.each(['completed', 'failed', 'cancelled'])('does not show progress for %s', status => {
    expect(mediaJobProgress(status)).toBeUndefined()
  })
})
