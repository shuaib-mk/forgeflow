import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  auth: {
    user: { id: 'user-id', email: 'dev@example.test', displayName: 'Local Developer' },
    organizations: [{ organization: { id: 'org-id', name: 'Example', slug: 'example', createdAt: '2026-01-01T00:00:00Z' }, role: 'owner' }],
    organizationId: 'org-id',
    loading: false,
    login: vi.fn(),
    register: vi.fn(),
    selectOrganization: vi.fn(),
    logout: vi.fn(),
  },
  api: {
    analytics: vi.fn(),
    projects: vi.fn(),
    project: vi.fn(),
    createProject: vi.fn(),
    tasks: vi.fn(),
    createTask: vi.fn(),
    updateTask: vi.fn(),
    repositories: vi.fn(),
    createRepository: vi.fn(),
    workflows: vi.fn(),
    createWorkflow: vi.fn(),
    runWorkflow: vi.fn(),
    runs: vi.fn(),
    run: vi.fn(),
    logs: vi.fn(),
    audit: vi.fn(),
    plugins: vi.fn(),
  },
}))

vi.mock('../stores/AuthContext', () => ({ useAuth: () => mocks.auth }))
vi.mock('../services/api', () => ({ api: mocks.api }))

import { DashboardPage } from './DashboardPage'
import { AnalyticsPage, AuditPage } from './InsightsPages'
import { LoginPage } from './LoginPage'
import { ProjectDetailPage } from './ProjectDetailPage'
import { NewProjectPage, ProjectsPage } from './ProjectsPage'
import { RunDetailPage, RunsPage } from './RunPage'
import { PluginsPage, SettingsPage } from './UtilityPages'
import { WorkflowEditorPage } from './WorkflowEditorPage'

const project = {
  id: 'project-id',
  organizationId: 'org-id',
  name: 'Demo Project',
  slug: 'demo',
  description: 'A project',
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}
const run = { id: 'run-id', workflowId: 'workflow-id', projectId: 'project-id', status: 'succeeded', createdAt: '2026-01-01T00:00:00Z' }

function route(path: string, element: React.ReactNode) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route
          path={path.includes('project-id') ? '/projects/:projectId' : path.includes('run-id') ? '/runs/:runId' : path}
          element={element}
        />
      </Routes>
    </MemoryRouter>,
  )
}

describe('dashboard pages', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mocks.api.analytics.mockResolvedValue({ projects: 1, openTasks: 1, runningWorkflows: 0, failedRuns: 0 })
    mocks.api.projects.mockResolvedValue({ items: [project], page: 1, pageSize: 20, totalItems: 1, totalPages: 1 })
    mocks.api.project.mockResolvedValue(project)
    mocks.api.tasks.mockResolvedValue({
      items: [
        {
          id: 'task-id',
          projectId: 'project-id',
          title: 'Ship it',
          description: 'Test it',
          status: 'open',
          createdAt: '2026-01-01T00:00:00Z',
        },
      ],
      page: 1,
      pageSize: 20,
      totalItems: 1,
      totalPages: 1,
    })
    mocks.api.repositories.mockResolvedValue({
      items: [{ id: 'repo-id', projectId: 'project-id', name: 'Demo', localPath: '/data/workspaces/demo', defaultBranch: 'main' }],
    })
    mocks.api.workflows.mockResolvedValue({
      items: [
        {
          id: 'workflow-id',
          projectId: 'project-id',
          name: 'Check',
          version: 1,
          steps: [{ id: 'test', name: 'Test', command: 'git' }],
          createdAt: '2026-01-01T00:00:00Z',
          updatedAt: '2026-01-01T00:00:00Z',
        },
      ],
    })
    mocks.api.runs.mockResolvedValue({ items: [run], page: 1, pageSize: 20, totalItems: 1, totalPages: 1 })
    mocks.api.run.mockResolvedValue(run)
    mocks.api.logs.mockResolvedValue({ items: ['done\n'] })
    mocks.api.audit.mockResolvedValue({
      items: [
        {
          id: 'event-id',
          action: 'project.created',
          resourceType: 'project',
          resourceId: 'project-id',
          requestId: 'request-id',
          createdAt: '2026-01-01T00:00:00Z',
        },
      ],
      page: 1,
      pageSize: 20,
      totalItems: 1,
      totalPages: 1,
    })
    mocks.api.plugins.mockResolvedValue({
      items: [{ id: 'plugin-id', name: 'Summary', version: '1.0.0', description: 'Summaries', enabled: true }],
    })
  })

  it('renders overview, projects, analytics, audit, plugins, and settings', async () => {
    const overview = route('/', <DashboardPage />)
    expect(await screen.findByText('Demo Project')).toBeVisible()
    overview.unmount()
    const projects = route('/projects', <ProjectsPage />)
    expect(await screen.findByRole('heading', { name: 'Projects' })).toBeVisible()
    projects.unmount()
    const analytics = route('/analytics', <AnalyticsPage />)
    expect(await screen.findByText('open Tasks')).toBeVisible()
    analytics.unmount()
    const audit = route('/audit', <AuditPage />)
    expect(await screen.findByText('project.created')).toBeVisible()
    audit.unmount()
    const plugins = route('/plugins', <PluginsPage />)
    expect(await screen.findByText('Summary')).toBeVisible()
    plugins.unmount()
    route('/settings', <SettingsPage />)
    expect(screen.getByRole('combobox', { name: /Active organization/ })).toHaveValue('org-id')
  })

  it('renders project operations and run history', async () => {
    const detail = route('/projects/project-id', <ProjectDetailPage />)
    expect(await screen.findByRole('heading', { name: 'Demo Project' })).toBeVisible()
    expect(screen.getByRole('button', { name: 'Add task' })).toBeEnabled()
    expect(screen.getByRole('button', { name: 'Run' })).toBeEnabled()
    detail.unmount()
    const runs = route('/runs', <RunsPage />)
    expect(await screen.findByText('run-id')).toBeVisible()
    runs.unmount()
    route('/runs/run-id', <RunDetailPage />)
    expect(await screen.findByText('done')).toBeVisible()
  })

  it('renders project and workflow creation forms', async () => {
    const create = route('/projects/new', <NewProjectPage />)
    expect(screen.getByRole('button', { name: 'Create project' })).toBeEnabled()
    create.unmount()
    route('/workflows/new', <WorkflowEditorPage />)
    expect(await screen.findByRole('combobox', { name: 'Project' })).toBeVisible()
    expect(screen.getByRole('button', { name: 'Save workflow' })).toBeDisabled()
  })
})

describe('account onboarding', () => {
  beforeEach(() => vi.clearAllMocks())

  it('supports sign in and self-contained registration', async () => {
    const view = route('/', <LoginPage />)
    fireEvent.change(screen.getByRole('textbox', { name: 'Email' }), { target: { value: 'dev@example.test' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'a-secure-password' } })
    fireEvent.click(screen.getAllByRole('button', { name: 'Sign in' }).at(-1)!)
    await waitFor(() => expect(mocks.auth.login).toHaveBeenCalledWith('dev@example.test', 'a-secure-password'))
    view.unmount()

    route('/', <LoginPage />)
    fireEvent.click(screen.getAllByRole('button', { name: 'Create account' })[0])
    fireEvent.change(screen.getByRole('textbox', { name: 'Your name' }), { target: { value: 'Dev' } })
    fireEvent.change(screen.getByRole('textbox', { name: 'Organization name' }), { target: { value: 'Example' } })
    fireEvent.change(screen.getByRole('textbox', { name: /Organization slug/ }), { target: { value: 'example' } })
    fireEvent.change(screen.getByRole('textbox', { name: 'Email' }), { target: { value: 'dev@example.test' } })
    fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'a-secure-password' } })
    fireEvent.click(screen.getAllByRole('button', { name: 'Create account' }).at(-1)!)
    await waitFor(() => expect(mocks.auth.register).toHaveBeenCalledWith(expect.objectContaining({ organizationSlug: 'example' })))
  })
})
