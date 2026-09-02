import type {
  Analytics,
  ApiFailure,
  AuditEvent,
  Organization,
  OrganizationMembership,
  Page,
  Plugin,
  Project,
  Repository,
  Task,
  User,
  Workflow,
  WorkflowRun,
} from '../types/models'

const baseURL = import.meta.env.VITE_API_URL ?? ''

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public requestId = '',
    public fields?: Record<string, string>,
  ) {
    super(message)
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = sessionStorage.getItem('forgeflow.token')
  const response = await fetch(`${baseURL}${path}`, {
    ...options,
    headers: {
      Accept: 'application/json',
      ...(options.body ? { 'Content-Type': 'application/json' } : {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  })
  if (!response.ok) {
    const payload = (await response.json().catch(() => ({ error: { message: response.statusText, requestId: '' } }))) as ApiFailure
    throw new ApiError(payload.error.message, response.status, payload.error.requestId, payload.error.fields)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export const api = {
  register: (input: { email: string; displayName: string; password: string; organizationName: string; organizationSlug: string }) =>
    request<{ user: User; organization: Organization }>('/api/v1/auth/register', { method: 'POST', body: JSON.stringify(input) }),
  login: (email: string, password: string) =>
    request<{ token: string; expiresAt: string; user: User }>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),
  me: () => request<User>('/api/v1/auth/me'),
  organizations: () => request<{ items: OrganizationMembership[] }>('/api/v1/organizations'),
  logout: () => request<void>('/api/v1/auth/logout', { method: 'POST' }),
  projects: (organizationId: string, search = '') =>
    request<Page<Project>>(
      `/api/v1/projects?organizationId=${encodeURIComponent(organizationId)}&search=${encodeURIComponent(search)}&order=desc`,
    ),
  project: (id: string) => request<Project>(`/api/v1/projects/${encodeURIComponent(id)}`),
  createProject: (input: { organizationId: string; name: string; slug: string; description: string }) =>
    request<Project>('/api/v1/projects', { method: 'POST', body: JSON.stringify(input) }),
  tasks: (projectId: string) => request<Page<Task>>(`/api/v1/projects/${encodeURIComponent(projectId)}/tasks`),
  createTask: (projectId: string, input: { title: string; description: string }) =>
    request<Task>(`/api/v1/projects/${encodeURIComponent(projectId)}/tasks`, { method: 'POST', body: JSON.stringify(input) }),
  updateTask: (projectId: string, taskId: string, status: Task['status']) =>
    request<Task>(`/api/v1/projects/${encodeURIComponent(projectId)}/tasks/${encodeURIComponent(taskId)}`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    }),
  repositories: (projectId: string) => request<{ items: Repository[] }>(`/api/v1/projects/${encodeURIComponent(projectId)}/repositories`),
  createRepository: (projectId: string, input: { name: string; localPath: string; initialize?: boolean }) =>
    request<Repository>(`/api/v1/projects/${encodeURIComponent(projectId)}/repositories`, { method: 'POST', body: JSON.stringify(input) }),
  workflows: (projectId: string) => request<{ items: Workflow[] }>(`/api/v1/projects/${encodeURIComponent(projectId)}/workflows`),
  createWorkflow: (projectId: string, definition: unknown) =>
    request<Workflow>(`/api/v1/projects/${encodeURIComponent(projectId)}/workflows`, { method: 'POST', body: JSON.stringify(definition) }),
  runWorkflow: (workflowId: string) => request<WorkflowRun>(`/api/v1/workflows/${encodeURIComponent(workflowId)}/runs`, { method: 'POST' }),
  runs: (organizationId: string, projectId = '') =>
    request<Page<WorkflowRun>>(
      `/api/v1/runs?organizationId=${encodeURIComponent(organizationId)}&projectId=${encodeURIComponent(projectId)}`,
    ),
  run: (id: string) => request<WorkflowRun>(`/api/v1/runs/${encodeURIComponent(id)}`),
  logs: (id: string) => request<{ items: string[] }>(`/api/v1/runs/${encodeURIComponent(id)}/logs?after=-1`),
  analytics: (organizationId: string) => request<Analytics>(`/api/v1/analytics?organizationId=${encodeURIComponent(organizationId)}`),
  audit: (organizationId: string) => request<Page<AuditEvent>>(`/api/v1/audit?organizationId=${encodeURIComponent(organizationId)}`),
  plugins: () => request<{ items: Plugin[] }>('/api/v1/plugins'),
}
