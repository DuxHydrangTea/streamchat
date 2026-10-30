export function validateRegistration(username: string, password: string, displayName: string) {
  if (!/^[a-zA-Z0-9_]{3,32}$/.test(username)) return 'Username must be 3-32 letters, numbers, or underscores.'
  if (password.length < 10 || password.length > 72) return 'Password must be 10-72 characters.'
  if (displayName.length > 80) return 'Display name must be 80 characters or fewer.'
  return null
}
