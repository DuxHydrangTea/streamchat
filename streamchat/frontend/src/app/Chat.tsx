import { FormEvent, useEffect, useRef, useState } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { preserveScrollAnchor } from './virtualMessages'
import { MessageIcon } from './Icons'

type Conversation = { id: string; type: 'direct' | 'group'; name?: string }
type MediaJob = { id: string; status: string; error_message?: string; title?: string; duration_seconds?: number; width?: number; height?: number; size_bytes?: number; playback_url?: string }
type Message = { id: string; sender_id: string; sender_name: string; client_message_id: string; text_content: string; created_at: string; type?: string; media_job?: MediaJob }
type User = { id: string; display_name: string }
type Page = { messages: Message[]; next_cursor?: string; has_more: boolean }
export function filterConversations<T extends Conversation>(conversations: T[], query: string): T[] {
  const normalized = query.trim().toLowerCase()
  return normalized ? conversations.filter(conversation => (conversation.name || (conversation.type === 'direct' ? 'Direct conversation' : 'Group conversation')).toLowerCase().includes(normalized)) : conversations
}
export function videoURL(text: string): string | undefined {
  const trimmed = text.trim()
  return /^https:\/\/\S+$/i.test(trimmed) ? trimmed : undefined
}

async function api<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, { credentials: 'include', headers: { 'Content-Type': 'application/json' }, ...init })
  if (!response.ok) throw new Error((await response.json()).error ?? 'Request failed')
  return response.status === 204 ? undefined as T : response.json()
}

function SearchIcon() { return <svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="10.8" cy="10.8" r="6.8" /><path d="m16 16 4.5 4.5" /></svg> }
function PlusIcon() { return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14M5 12h14" /></svg> }
function SignOutIcon() { return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M10 17l5-5-5-5M15 12H3" /><path d="M12 3h6a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-6" /></svg> }
function MenuIcon() { return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 6h16M4 12h16M4 18h16" /></svg> }
function PaperclipIcon() { return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m8.5 12.5 7.1-7.1a4 4 0 0 1 5.7 5.7l-9.2 9.2a6 6 0 0 1-8.5-8.5l9.2-9.2" /></svg> }
function SendIcon() { return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m21 3-7.2 18-3.6-7.2L3 10.2 21 3Z" /><path d="M10.2 13.8 15 9" /></svg> }
function CloseIcon() { return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m6 6 12 12M18 6 6 18" /></svg> }

export function Chat({ user, onSignOut }: { user: User; onSignOut: () => void }) {
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [selected, setSelected] = useState('')
  const [messages, setMessages] = useState<Message[]>([])
  const [draft, setDraft] = useState('')
  const [error, setError] = useState('')
  const [newType, setNewType] = useState<'direct' | 'group'>('direct')
  const [newName, setNewName] = useState('')
  const [memberID, setMemberID] = useState('')
  const [olderCursor, setOlderCursor] = useState<string>()
  const [hasOlder, setHasOlder] = useState(false)
  const [unread, setUnread] = useState(0)
  const [conversationQuery, setConversationQuery] = useState('')
  const [mobileSidebarOpen, setMobileSidebarOpen] = useState(false)
  const [loadingMessages, setLoadingMessages] = useState(false)
  const list = useRef<HTMLDivElement>(null)
  const filteredConversations = filterConversations(conversations, conversationQuery)
  const nearBottom = useRef(true)
  const virtualizer = useVirtualizer({ count: messages.length, getScrollElement: () => list.current, estimateSize: () => 72, overscan: 8, getItemKey: index => messages[index]?.id ?? index })

  useEffect(() => {
    void api<{ conversations: Conversation[] }>('/api/v1/conversations').then(result => {
      setConversations(result.conversations)
      if (result.conversations[0]) setSelected(result.conversations[0].id)
    }).catch(e => setError(e.message))
  }, [])

  useEffect(() => {
    if (!selected) { setMessages([]); return }
    let active = true
    setMessages([]); setUnread(0); setOlderCursor(undefined); setHasOlder(false); setLoadingMessages(true)
    api<Page>(`/api/v1/conversations/${selected}/messages?limit=30`).then(page => {
      if (active) { setMessages(page.messages); setOlderCursor(page.next_cursor); setHasOlder(page.has_more); requestAnimationFrame(() => { if (list.current) list.current.scrollTop = list.current.scrollHeight }) }
    }).catch(e => setError(e.message)).finally(() => { if (active) setLoadingMessages(false) })
    return () => { active = false }
  }, [selected])

  useEffect(() => {
    const pending = messages.filter(m => m.media_job?.id && !['completed', 'failed', 'cancelled'].includes(m.media_job.status))
    if (!pending.length) return
    let active = true
    const timer = window.setInterval(() => {
      for (const message of pending) {
        void api<MediaJob>(`/api/v1/media/jobs/${message.media_job!.id}`).then(job => {
          if (!active) return
          setMessages(old => old.map(item => item.id === message.id ? { ...item, media_job: job } : item))
        }).catch(() => undefined)
      }
    }, 2000)
    return () => { active = false; window.clearInterval(timer) }
  }, [messages])

  useEffect(() => {
    const scheme = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const socket = new WebSocket(`${scheme}//${location.host}/ws`)
    socket.onopen = () => { if (selected) socket.send(JSON.stringify({ action: 'join', conversation_id: selected })) }
    socket.onmessage = e => {
      try {
        const event = JSON.parse(e.data)
        if (event.event_type !== 'message.created' || event.conversation_id !== selected) return
        setMessages(old => old.some(m => m.id === event.payload.id || m.client_message_id === event.payload.client_message_id) ? old : [...old, event.payload])
        if (nearBottom.current) requestAnimationFrame(() => { if (list.current) list.current.scrollTop = list.current.scrollHeight })
        else setUnread(n => n + 1)
      } catch { setError('Received an unreadable realtime event') }
    }
    return () => socket.close()
  }, [selected])

  async function send(event: FormEvent) {
    event.preventDefault(); if (!selected || !draft.trim()) return
    const text = draft.trim(), clientID = crypto.randomUUID()
    const source = videoURL(text)
    const optimistic: Message = { id: `pending-${clientID}`, sender_id: user.id, sender_name: user.display_name, client_message_id: clientID, text_content: text, created_at: new Date().toISOString(), type: source ? 'video' : 'text', media_job: source ? { id: '', status: 'queued' } : undefined }
    setMessages(old => [...old, optimistic]); setDraft(''); setError('')
    requestAnimationFrame(() => { if (list.current) list.current.scrollTop = list.current.scrollHeight })
    try {
      const saved = source
        ? await api<Message>(`/api/v1/conversations/${selected}/media-jobs`, { method: 'POST', body: JSON.stringify({ client_message_id: clientID, url: source }) })
        : await api<Message>(`/api/v1/conversations/${selected}/messages`, { method: 'POST', body: JSON.stringify({ client_message_id: clientID, text_content: text }) })
      setMessages(old => old.map(m => m.client_message_id === clientID ? { ...saved, sender_name: user.display_name } : m))
    } catch (e) { setMessages(old => old.filter(m => m.client_message_id !== clientID)); setDraft(text); setError(e instanceof Error ? e.message : 'Message failed') }
  }

  async function uploadFile(file: File) {
    if (!selected) return
    const data = new FormData(); data.set('file', file)
    try { await api(`/api/v1/conversations/${selected}/attachments`, { method: 'POST', body: data, headers: {} }); setError('Upload complete. Attachments will be attached to messages in the next phase.') }
    catch (e) { setError(e instanceof Error ? e.message : 'Upload failed') }
  }

  async function createConversation(event: FormEvent) {
    event.preventDefault()
    try {
      const created = await api<Conversation>('/api/v1/conversations', { method: 'POST', body: JSON.stringify({ type: newType, name: newName, member_ids: memberID ? [memberID] : [] }) })
      setConversations(old => [created, ...old]); setSelected(created.id); setNewName(''); setMemberID('')
    } catch (e) { setError(e instanceof Error ? e.message : 'Unable to create conversation') }
  }

  async function loadOlder() {
    if (!selected || !olderCursor || !hasOlder || !list.current) return
    const element = list.current, previousHeight = element.scrollHeight, previousTop = element.scrollTop
    try {
      const page = await api<Page>(`/api/v1/conversations/${selected}/messages?before=${encodeURIComponent(olderCursor)}&limit=30`)
      setMessages(old => [...page.messages, ...old]); setOlderCursor(page.next_cursor); setHasOlder(page.has_more)
      requestAnimationFrame(() => { if (list.current) list.current.scrollTop = preserveScrollAnchor(previousTop, previousHeight, list.current.scrollHeight) })
    } catch (e) { setError(e instanceof Error ? e.message : 'Unable to load older messages') }
  }

  return <main className="chat-shell">
    {mobileSidebarOpen && <button className="mobile-scrim" aria-label="Close conversations" onClick={() => setMobileSidebarOpen(false)} />}
    <aside className={`chat-sidebar${mobileSidebarOpen ? ' is-open' : ''}`}>
      <div className="side-head"><a className="brand-mark" href="#" aria-label="StreamChat home"><span className="logo">S</span><span><b>StreamChat</b><small>Messages</small></span></a><button className="icon-button sign-out" aria-label="Sign out" title="Sign out" onClick={onSignOut}><SignOutIcon /></button></div>
      <div className="sidebar-content">
        <div className="sidebar-title-row"><div><h1>Messages</h1><p>{conversations.length} conversations</p></div><button className="icon-button new-chat-shortcut" aria-label="New conversation" title="New conversation" onClick={() => document.querySelector<HTMLInputElement>('.new-conversation input')?.focus()}><PlusIcon /></button></div>
        <label className="conversation-search"><SearchIcon /><input aria-label="Search conversations" placeholder="Search conversations" value={conversationQuery} onChange={e => setConversationQuery(e.target.value)} /></label>
        <div className="conversation-list" aria-label="Conversations">
          {filteredConversations.map(c => <button key={c.id} className={selected === c.id ? 'conversation active' : 'conversation'} onClick={() => { setSelected(c.id); setMobileSidebarOpen(false) }}><span className={`conversation-avatar ${c.type}`}>{(c.name || (c.type === 'direct' ? 'D' : 'G')).slice(0, 1).toUpperCase()}</span><span className="conversation-copy"><strong>{c.name || (c.type === 'direct' ? 'Direct conversation' : 'Group conversation')}</strong><small>{c.type === 'direct' ? 'Direct message' : 'Group conversation'}</small></span><span className="conversation-chevron" aria-hidden="true">›</span></button>)}
          {!filteredConversations.length && <p className="conversation-empty">{conversationQuery ? 'No conversations match your search.' : 'Your conversations will appear here.'}</p>}
        </div>
        <form className="new-conversation" onSubmit={createConversation}>
          <h2>Start a conversation</h2>
          <label>Conversation type<select value={newType} onChange={e => setNewType(e.target.value as 'direct' | 'group')}><option value="direct">Direct message</option><option value="group">Group conversation</option></select></label>
          {newType === 'group' && <label>Group name<input placeholder="e.g. Product team" value={newName} onChange={e => setNewName(e.target.value)} required /></label>}
          <label>{newType === 'group' ? 'First member user ID' : 'User ID'}<input placeholder="Enter a user ID" value={memberID} onChange={e => setMemberID(e.target.value)} required /></label>
          <button className="secondary-action" type="submit"><PlusIcon /> Create conversation</button>
        </form>
      </div>
      <div className="sidebar-footer"><span className="user-avatar">{user.display_name.slice(0, 1).toUpperCase()}</span><span className="user-copy"><strong>{user.display_name}</strong><small>Signed in</small></span><button className="icon-button" onClick={onSignOut} aria-label="Sign out" title="Sign out"><SignOutIcon /></button></div>
    </aside>
    <section className="chat-main">
      <header className="chat-header"><button className="icon-button mobile-menu" aria-label="Open conversations" title="Open conversations" onClick={() => setMobileSidebarOpen(true)}><MenuIcon /></button><span className={`conversation-avatar header-avatar ${conversations.find(c => c.id === selected)?.type || 'direct'}`}>{(conversations.find(c => c.id === selected)?.name || 'C').slice(0, 1).toUpperCase()}</span><div className="chat-heading"><strong>{conversations.find(c => c.id === selected)?.name || 'Select a conversation'}</strong><small>{selected ? 'Conversation' : 'Your messages'}</small></div></header>
      <div className="message-list" ref={list} onScroll={e => { const el = e.currentTarget; nearBottom.current = el.scrollHeight - el.scrollTop - el.clientHeight < 100 }}>
        {hasOlder && <button className="older" onClick={() => void loadOlder()}>Load earlier messages</button>}
        {loadingMessages ? <div className="message-loading" aria-label="Loading messages"><span /><span /><span /></div> : <>
          <div className="virtual-messages" style={{ height: `${virtualizer.getTotalSize()}px` }}>
            {virtualizer.getVirtualItems().map(item => { const m = messages[item.index]; return <article data-index={item.index} ref={virtualizer.measureElement} className={`${m.sender_id === user.id ? 'message mine' : 'message'} virtual-message`} key={m.id} style={{ transform: `translateY(${item.start}px)` }}><small>{m.sender_id === user.id ? 'You' : m.sender_name}</small>{m.type === 'video' && m.media_job ? <div className="video-message">{m.media_job.status === 'completed' && m.media_job.playback_url ? <video controls preload="metadata" src={m.media_job.playback_url} /> : <div className="video-status"><strong>{m.media_job.status === 'failed' ? 'Video processing failed' : m.media_job.status === 'cancelled' ? 'Video processing cancelled' : 'Preparing video'}</strong><span>{m.media_job.status === 'failed' ? 'The original link is still available.' : m.media_job.status === 'cancelled' ? 'You can send the link again to retry.' : 'This usually takes a short while.'}</span>{m.media_job.status === 'failed' && <button onClick={() => void api(`/api/v1/media/jobs/${m.media_job!.id}/retry`, { method: 'POST' }).then(() => setMessages(old => old.map(item => item.id === m.id ? { ...item, media_job: { ...item.media_job!, status: 'queued', error_message: undefined } } : item))).catch(e => setError(e instanceof Error ? e.message : 'Retry failed'))}>Retry download</button>}</div>}{m.media_job.title && <strong className="video-title">{m.media_job.title}</strong>}{m.media_job.duration_seconds ? <small className="video-duration">{Math.floor(m.media_job.duration_seconds / 60)}:{String(Math.floor(m.media_job.duration_seconds % 60)).padStart(2, '0')}</small> : null}<a href={m.text_content} target="_blank" rel="noreferrer">Open source link</a></div> : <p>{m.text_content}</p>}<time>{new Date(m.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</time></article> })}
          </div>
          {!messages.length && !loadingMessages && <div className="chat-empty"><span className="empty-icon"><MessageIcon /></span><h2>{selected ? 'Start the conversation' : 'Your conversations, together'}</h2><p>{selected ? 'Send a message or share a video link to get things moving.' : 'Choose a conversation from the sidebar, or create a new one to begin.'}</p>{!selected && <button onClick={() => setMobileSidebarOpen(true)}>Browse conversations</button>}</div>}
        </>}
      </div>
      {unread > 0 && <button className="unread" onClick={() => { if (list.current) list.current.scrollTop = list.current.scrollHeight; setUnread(0) }}>{unread} new {unread === 1 ? 'message' : 'messages'} <span aria-hidden="true">↓</span></button>}
      {error && <div className="chat-error" role="alert"><span>{error}</span><button onClick={() => setError('')} aria-label="Dismiss error"><CloseIcon /></button></div>}
      <form className="composer" onSubmit={send}><label className="attach-button" title="Upload an attachment" aria-label="Upload an attachment"><PaperclipIcon /><input type="file" hidden onChange={e => { const file = e.target.files?.[0]; if (file) void uploadFile(file); e.currentTarget.value = '' }} /></label><textarea aria-label="Write a message" value={draft} onChange={e => setDraft(e.target.value)} onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); e.currentTarget.form?.requestSubmit() } }} placeholder={selected ? 'Write a message…' : 'Select a conversation to start'} maxLength={4000} disabled={!selected} /><button className="send-button" disabled={!draft.trim() || !selected} aria-label="Send message"><span>Send</span><SendIcon /></button></form>
    </section>
  </main>
}
