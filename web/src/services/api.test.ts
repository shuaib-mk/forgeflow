import { beforeEach, describe, expect, it, vi } from 'vitest'
import { api } from './api'

describe('dashboard API client', () => {
  const fetchMock = vi.fn()

  beforeEach(() => {
    fetchMock.mockReset()
    fetchMock.mockResolvedValue({ ok: true, status: 200, json: async () => ({ items: [] }) })
    vi.stubGlobal('fetch', fetchMock)
    sessionStorage.clear()
    sessionStorage.setItem('forgeflow.token', 'session-token')
  })

  it('loads organization run history with the project filter', async () => {
    await api.runs('org-id', 'project-id')
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/runs?organizationId=org-id&projectId=project-id', expect.objectContaining({ headers: expect.objectContaining({ Authorization: 'Bearer session-token' }) }))
  })

  it('can request a managed repository', async () => {
    await api.createRepository('project-id', { name: 'Product', localPath: 'product', initialize: true })
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/projects/project-id/repositories', expect.objectContaining({ method: 'POST', body: JSON.stringify({ name: 'Product', localPath: 'product', initialize: true }) }))
  })
})
