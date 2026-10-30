export type SessionUser = { id: string; username: string; display_name: string }

export async function loadCurrentUser(fetcher: typeof fetch = fetch): Promise<SessionUser | null> {
  const response = await fetcher('/api/v1/auth/me', { credentials: 'include' })
  if (response.status === 401) return null
  if (!response.ok) throw new Error(`Could not restore session (${response.status})`)
  return response.json() as Promise<SessionUser>
}
