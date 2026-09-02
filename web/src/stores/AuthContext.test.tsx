import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const api = vi.hoisted(() => ({
  me: vi.fn(),
  organizations: vi.fn(),
  login: vi.fn(),
  register: vi.fn(),
  logout: vi.fn(),
}))
vi.mock('../services/api', () => ({ api }))

import { AuthProvider, useAuth } from './AuthContext'

const user = { id: 'user-id', email: 'dev@example.test', displayName: 'Developer' }
const membership = { organization: { id: 'org-id', name: 'Example', slug: 'example', createdAt: '2026-01-01T00:00:00Z' }, role: 'owner' }

function Harness() {
  const auth = useAuth()
  return (
    <div>
      <span>{auth.loading ? 'loading' : 'ready'}</span>
      <span>{auth.user?.displayName ?? 'signed-out'}</span>
      <span>{auth.organizationId || 'no-org'}</span>
      <button onClick={() => void auth.login('dev@example.test', 'a-secure-password')}>login</button>
      <button
        onClick={() =>
          void auth.register({
            email: 'new@example.test',
            displayName: 'New',
            password: 'a-secure-password',
            organizationName: 'New Org',
            organizationSlug: 'new-org',
          })
        }
      >
        register
      </button>
      <button onClick={() => auth.selectOrganization('second-org')}>select</button>
      <button onClick={() => void auth.logout()}>logout</button>
    </div>
  )
}

describe('AuthProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    sessionStorage.clear()
    localStorage.clear()
    api.me.mockResolvedValue(user)
    api.organizations.mockResolvedValue({ items: [membership] })
    api.login.mockResolvedValue({ token: 'session-token', expiresAt: '2026-01-02T00:00:00Z', user })
    api.register.mockResolvedValue({ user, organization: membership.organization })
    api.logout.mockResolvedValue(undefined)
  })

  it('logs in, selects an organization, and logs out', async () => {
    render(
      <AuthProvider>
        <Harness />
      </AuthProvider>,
    )
    expect(await screen.findByText('ready')).toBeVisible()
    fireEvent.click(screen.getByRole('button', { name: 'login' }))
    expect(await screen.findByText('Developer')).toBeVisible()
    expect(screen.getByText('org-id')).toBeVisible()
    expect(sessionStorage.getItem('forgeflow.token')).toBe('session-token')
    fireEvent.click(screen.getByRole('button', { name: 'select' }))
    expect(screen.getByText('second-org')).toBeVisible()
    expect(localStorage.getItem('forgeflow.organization')).toBe('second-org')
    fireEvent.click(screen.getByRole('button', { name: 'logout' }))
    await waitFor(() => expect(screen.getByText('signed-out')).toBeVisible())
    expect(sessionStorage.getItem('forgeflow.token')).toBeNull()
  })

  it('registers and automatically authenticates a new account', async () => {
    render(
      <AuthProvider>
        <Harness />
      </AuthProvider>,
    )
    expect(await screen.findByText('ready')).toBeVisible()
    fireEvent.click(screen.getByRole('button', { name: 'register' }))
    await waitFor(() => expect(api.register).toHaveBeenCalledWith(expect.objectContaining({ organizationSlug: 'new-org' })))
    expect(await screen.findByText('Developer')).toBeVisible()
    expect(api.login).toHaveBeenCalledWith('new@example.test', 'a-secure-password')
  })

  it('restores an existing session on startup', async () => {
    sessionStorage.setItem('forgeflow.token', 'existing-token')
    render(
      <AuthProvider>
        <Harness />
      </AuthProvider>,
    )
    expect(await screen.findByText('Developer')).toBeVisible()
    expect(api.me).toHaveBeenCalledOnce()
    expect(api.organizations).toHaveBeenCalledOnce()
  })
})
