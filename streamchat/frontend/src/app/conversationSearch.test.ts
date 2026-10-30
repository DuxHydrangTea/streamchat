import { describe, expect, it } from 'vitest'
import { filterConversations } from './Chat'

const conversations = [
  { id: '1', type: 'direct' as const, name: 'Casey' },
  { id: '2', type: 'group' as const, name: 'Product team' },
  { id: '3', type: 'direct' as const },
]

describe('filterConversations', () => {
  it('matches names case-insensitively and ignores surrounding spaces', () => {
    expect(filterConversations(conversations, '  PRODUCT ')).toEqual([conversations[1]])
  })

  it('searches fallback labels and returns all conversations for an empty query', () => {
    expect(filterConversations(conversations, 'direct')).toEqual([conversations[2]])
    expect(filterConversations(conversations, '  ')).toEqual(conversations)
  })
})
