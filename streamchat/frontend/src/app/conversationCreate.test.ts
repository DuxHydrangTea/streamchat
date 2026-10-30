import { describe, expect, it } from 'vitest'
import { createConversationPayload } from './Chat'

describe('createConversationPayload', () => {
  it('sends the username instead of a user ID', () => {
    expect(createConversationPayload('direct', '', '  casey_1 ')).toEqual({
      type: 'direct',
      name: '',
      member_usernames: ['casey_1'],
    })
  })

  it('includes the group name and omits empty usernames', () => {
    expect(createConversationPayload('group', 'Product team', '  ')).toEqual({
      type: 'group',
      name: 'Product team',
      member_usernames: [],
    })
  })
})
