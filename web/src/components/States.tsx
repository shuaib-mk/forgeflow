import type { ReactNode } from 'react'

export function LoadingState({ label = 'Loading' }: { label?: string }) { return <div className="state" role="status"><span className="spinner" aria-hidden="true" />{label}…</div> }
export function ErrorState({ error, retry }: { error: Error; retry?: () => void }) { return <div className="state error" role="alert"><strong>Something went wrong</strong><p>{error.message}</p>{retry && <button className="button secondary" onClick={retry}>Try again</button>}</div> }
export function EmptyState({ title, body, action }: { title: string; body: string; action?: ReactNode }) { return <div className="state empty"><span className="empty-icon" aria-hidden="true">◇</span><strong>{title}</strong><p>{body}</p>{action}</div> }
export function StatusPill({ status }: { status: string }) { return <span className={`pill ${status}`}>{status.replace('_', ' ')}</span> }

