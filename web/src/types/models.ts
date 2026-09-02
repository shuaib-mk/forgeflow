export type User = { id: string; email: string; displayName: string }
export type Organization = { id: string; name: string; slug: string; createdAt: string }
export type OrganizationMembership = { organization: Organization; role: 'owner' | 'admin' | 'developer' | 'viewer' }
export type Project = {
  id: string
  organizationId: string
  name: string
  slug: string
  description: string
  createdAt: string
  updatedAt: string
}
export type Task = {
  id: string
  projectId: string
  title: string
  description: string
  status: 'open' | 'in_progress' | 'done' | 'canceled'
  createdAt: string
}
export type Repository = { id: string; projectId: string; name: string; localPath: string; defaultBranch: string }
export type WorkflowRun = {
  id: string
  workflowId: string
  projectId: string
  status: 'queued' | 'running' | 'succeeded' | 'failed' | 'canceled' | 'timed_out'
  createdAt: string
  error?: string
}
export type WorkflowStep = {
  id: string
  name: string
  command: string
  args?: string[]
  dependsOn?: string[]
  retries?: number
  continueOnFail?: boolean
}
export type Workflow = {
  id: string
  projectId: string
  name: string
  version: number
  steps: WorkflowStep[]
  createdAt: string
  updatedAt: string
}
export type AuditEvent = { id: string; action: string; resourceType: string; resourceId: string; requestId: string; createdAt: string }
export type Page<T> = { items: T[]; page: number; pageSize: number; totalItems: number; totalPages: number }
export type Analytics = { projects: number; openTasks: number; runningWorkflows: number; failedRuns: number }
export type Plugin = { id: string; name: string; version: string; description: string; enabled: boolean }
export type ApiFailure = { error: { code: string; message: string; requestId: string; fields?: Record<string, string> } }
