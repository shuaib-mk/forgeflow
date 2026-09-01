import type { Analytics, ApiFailure, AuditEvent, Page, Plugin, Project, Repository, Task, User, WorkflowRun } from '../types/models'

const baseURL = import.meta.env.VITE_API_URL ?? ''

export class ApiError extends Error {
  constructor(message: string, public status: number, public requestId = '', public fields?: Record<string, string>) { super(message) }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = sessionStorage.getItem('forgeflow.token')
  const response = await fetch(`${baseURL}${path}`, { ...options, headers: { Accept: 'application/json', ...(options.body ? { 'Content-Type': 'application/json' } : {}), ...(token ? { Authorization: `Bearer ${token}` } : {}), ...options.headers } })
  if (!response.ok) {
    const payload = (await response.json().catch(() => ({ error: { message: response.statusText, requestId: '' } }))) as ApiFailure
    throw new ApiError(payload.error.message, response.status, payload.error.requestId, payload.error.fields)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export const api = {
  login: (email: string, password: string) => request<{ token: string; expiresAt: string; user: User }>('/api/v1/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),
  me: () => request<User>('/api/v1/auth/me'),
  logout: () => request<void>('/api/v1/auth/logout', { method: 'POST' }),
  projects: (organizationId: string, search = '') => request<Page<Project>>(`/api/v1/projects?organizationId=${encodeURIComponent(organizationId)}&search=${encodeURIComponent(search)}&order=desc`),
  project: (id: string) => request<Project>(`/api/v1/projects/${encodeURIComponent(id)}`),
  createProject: (input: { organizationId: string; name: string; slug: string; description: string }) => request<Project>('/api/v1/projects', { method: 'POST', body: JSON.stringify(input) }),
  tasks: (projectId: string) => request<Page<Task>>(`/api/v1/projects/${encodeURIComponent(projectId)}/tasks`),
  repositories: (projectId: string) => request<{ items: Repository[] }>(`/api/v1/projects/${encodeURIComponent(projectId)}/repositories`),
  createWorkflow: (projectId: string, definition: unknown) => request(`/api/v1/projects/${encodeURIComponent(projectId)}/workflows`, { method: 'POST', body: JSON.stringify(definition) }),
  run: (id: string) => request<WorkflowRun>(`/api/v1/runs/${encodeURIComponent(id)}`),
  logs: (id: string) => request<{ items: string[] }>(`/api/v1/runs/${encodeURIComponent(id)}/logs?after=-1`),
  analytics: (organizationId: string) => request<Analytics>(`/api/v1/analytics?organizationId=${encodeURIComponent(organizationId)}`),
  audit: (organizationId: string) => request<Page<AuditEvent>>(`/api/v1/audit?organizationId=${encodeURIComponent(organizationId)}`),
  plugins: () => request<{ items: Plugin[] }>('/api/v1/plugins'),
}
