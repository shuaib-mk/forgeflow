import { type FormEvent, useState } from 'react'
import { useAuth } from '../stores/AuthContext'

export function LoginPage() {
  const { login, register } = useAuth()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setSubmitting(true)
    setError('')
    const data = new FormData(event.currentTarget)
    try {
      if (mode === 'login') {
        await login(String(data.get('email')), String(data.get('password')))
      } else {
        await register({
          email: String(data.get('email')),
          displayName: String(data.get('displayName')),
          password: String(data.get('password')),
          organizationName: String(data.get('organizationName')),
          organizationSlug: String(data.get('organizationSlug')),
        })
      }
    } catch (value) {
      setError(value instanceof Error ? value.message : mode === 'login' ? 'Login failed' : 'Registration failed')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <main className="auth-page">
      <section className="auth-copy">
        <a className="brand light" href="/">
          <span className="brand-mark">F</span>Forge<span>Flow</span>
        </a>
        <div>
          <span className="eyebrow">Local by design</span>
          <h1>
            Ship better software,
            <br />
            <em>one flow at a time.</em>
          </h1>
          <p>Projects, quality checks, test runs, and engineering history in one dependable workspace.</p>
        </div>
        <small>Open source · Self-hosted · Built for small teams</small>
      </section>
      <section className="auth-form-wrap">
        <form className="auth-form" onSubmit={submit}>
          <div className="auth-switch" role="tablist" aria-label="Account access">
            <button
              type="button"
              className={mode === 'login' ? 'active' : ''}
              onClick={() => {
                setMode('login')
                setError('')
              }}
            >
              Sign in
            </button>
            <button
              type="button"
              className={mode === 'register' ? 'active' : ''}
              onClick={() => {
                setMode('register')
                setError('')
              }}
            >
              Create account
            </button>
          </div>
          <span className="eyebrow">{mode === 'login' ? 'Welcome back' : 'First-time setup'}</span>
          <h2>{mode === 'login' ? 'Sign in to ForgeFlow' : 'Create your workspace'}</h2>
          <p>{mode === 'login' ? 'Use your local ForgeFlow account.' : 'Your organization is connected automatically—no IDs to copy.'}</p>
          {error && (
            <div className="inline-error" role="alert">
              {error}
            </div>
          )}
          {mode === 'register' && (
            <>
              <label>
                Your name
                <input name="displayName" autoComplete="name" maxLength={100} required />
              </label>
              <label>
                Organization name
                <input name="organizationName" maxLength={100} required />
              </label>
              <label>
                Organization slug
                <input name="organizationSlug" placeholder="my-team" pattern="[a-z0-9]+(-[a-z0-9]+)*" required />
                <small>Lowercase letters, numbers, and single hyphens.</small>
              </label>
            </>
          )}
          <label>
            Email
            <input name="email" type="email" autoComplete="email" required />
          </label>
          <label>
            Password
            <input
              name="password"
              type="password"
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              minLength={12}
              maxLength={128}
              required
            />
          </label>
          <button className="button primary wide" disabled={submitting}>
            {submitting ? 'Please wait…' : mode === 'login' ? 'Sign in' : 'Create account'}
          </button>
        </form>
      </section>
    </main>
  )
}
