import { describe, expect, it, vi } from 'vitest'
import { loadCurrentUser } from './authSession'

describe('loadCurrentUser', () => {
  it('restores the authenticated user from the session cookie', async () => {
    const user = { id: 'user-1', username: 'casey', display_name: 'Casey' }
    const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify(user), { status: 200 }))

    await expect(loadCurrentUser(fetcher)).resolves.toEqual(user)
    expect(fetcher).toHaveBeenCalledWith('/api/v1/auth/me', { credentials: 'include' })
  })

  it('returns null when there is no valid session', async () => {
    const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ error: 'authentication required' }), { status: 401 }))

    await expect(loadCurrentUser(fetcher)).resolves.toBeNull()
  })
})
