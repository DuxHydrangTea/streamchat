import { FormEvent, useEffect, useState } from 'react'
import { validateRegistration } from './validation'
import { Chat } from './Chat'
import { loadCurrentUser, type SessionUser } from './authSession'
import { MessageIcon } from './Icons'


async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, { credentials: 'include', headers: { 'Content-Type': 'application/json' }, ...init })
  if (!response.ok) throw new Error((await response.json()).error ?? 'Request failed')
  return response.status === 204 ? (undefined as T) : response.json()
}

export function App() {
  const [user, setUser] = useState<SessionUser | null>(null)
  const [checkingSession, setCheckingSession] = useState(true)
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [error, setError] = useState('')
  const [errorContext, setErrorContext] = useState<'session' | 'auth'>('auth')

  useEffect(() => {
    let active = true
    loadCurrentUser().then(currentUser => {
      if (active) setUser(currentUser)
    }).catch(err => {
      if (active) { setError(err instanceof Error ? err.message : 'Could not restore session'); setErrorContext('session') }
    }).finally(() => {
      if (active) setCheckingSession(false)
    })
    return () => { active = false }
  }, [])

  async function submit(event: FormEvent) {
    event.preventDefault(); setError(''); setErrorContext('auth')
    if (mode === 'register') {
      const validationError = validateRegistration(username, password, displayName)
      if (validationError) { setError(validationError); return }
    }
    try {
      const next = await request<SessionUser>(`/api/v1/auth/${mode}`, { method: 'POST', body: JSON.stringify({ username, password, display_name: displayName }) })
      setUser(next)
    } catch (err) { setError(err instanceof Error ? err.message : 'Unable to authenticate') }
  }

  if (checkingSession) return <main className="auth shell"><div className="auth-brand"><span className="logo">S</span><span><strong>StreamChat</strong><small>Real-time conversations</small></span></div><section className="auth-panel auth-loading" aria-label="Restoring session"><span className="loading-mark" /><div><strong>Checking your session</strong><p className="muted">One moment while we get your messages.</p></div></section></main>
  if (user) return <Chat user={user} onSignOut={async () => { await request('/api/v1/auth/logout', { method: 'POST' }); setUser(null) }} />
  return <main className="auth shell"><div className="auth-brand"><span className="logo">S</span><span><strong>StreamChat</strong><small>Real-time conversations</small></span></div><section className="auth-panel"><div className="auth-heading"><span className="auth-icon"><MessageIcon /></span><h1>{mode === 'login' ? 'Welcome back' : 'Create your account'}</h1><p className="muted">{mode === 'login' ? 'Sign in to pick up where you left off.' : 'Create an account to start a conversation.'}</p></div><form onSubmit={submit}>{mode === 'register' && <label htmlFor="display-name">Display name<input id="display-name" value={displayName} onChange={e => setDisplayName(e.target.value)} placeholder="Your name" autoComplete="name" /></label>}<label htmlFor="username">Username<input id="username" required value={username} onChange={e => setUsername(e.target.value)} placeholder="Enter your username" autoComplete="username" /></label><label htmlFor="password">Password<input id="password" required type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="At least 10 characters" autoComplete={mode === 'login' ? 'current-password' : 'new-password'} /></label>{error && <div className="error" role="alert"><strong>{errorContext === 'session' ? 'Could not check your session' : mode === 'login' ? 'We couldn’t sign you in' : 'We couldn’t create your account'}</strong><span>{error}</span></div>}<button className="primary" type="submit">{mode === 'login' ? 'Sign in' : 'Create account'}</button><button type="button" className="link" onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); setError('') }}>{mode === 'login' ? 'New to StreamChat? Create an account' : 'Already have an account? Sign in'}</button></form></section><p className="auth-footnote">Real-time conversations, right where you left them.</p></main>
}
