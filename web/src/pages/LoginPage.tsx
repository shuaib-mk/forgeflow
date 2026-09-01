import { FormEvent, useState } from 'react'
import { useAuth } from '../stores/AuthContext'

export function LoginPage() {
  const { login } = useAuth(); const [error, setError] = useState(''); const [submitting, setSubmitting] = useState(false)
  const submit = async (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); setSubmitting(true); setError(''); const data = new FormData(event.currentTarget); try { await login(String(data.get('email')), String(data.get('password'))) } catch (value) { setError(value instanceof Error ? value.message : 'Login failed') } finally { setSubmitting(false) } }
  return <main className="auth-page"><section className="auth-copy"><a className="brand light" href="/"><span className="brand-mark">F</span>Forge<span>Flow</span></a><div><span className="eyebrow">Local by design</span><h1>Ship better software,<br /><em>one flow at a time.</em></h1><p>Projects, quality checks, test runs, and engineering history in one dependable workspace.</p></div><small>Open source · Self-hosted · Built for small teams</small></section><section className="auth-form-wrap"><form className="auth-form" onSubmit={submit}><span className="eyebrow">Welcome back</span><h2>Sign in to ForgeFlow</h2><p>Use the account created for your local organization.</p>{error && <div className="inline-error" role="alert">{error}</div>}<label>Email<input name="email" type="email" autoComplete="email" required /></label><label>Password<input name="password" type="password" autoComplete="current-password" minLength={12} required /></label><button className="button primary wide" disabled={submitting}>{submitting ? 'Signing in…' : 'Sign in'}</button><p className="hint">Need an account? Register through <code>POST /api/v1/auth/register</code>.</p></form></section></main>
}

