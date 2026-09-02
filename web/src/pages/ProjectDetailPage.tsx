import { type FormEvent, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api } from '../services/api'
import { useAsync } from '../hooks/useAsync'
import { EmptyState, ErrorState, LoadingState, StatusPill } from '../components/States'
import { Page } from './DashboardPage'
import type { Task } from '../types/models'
import { useAuth } from '../stores/AuthContext'

export function ProjectDetailPage() {
  const { projectId = '' } = useParams()
  const { organizationId } = useAuth()
  const navigate = useNavigate()
  const [actionError, setActionError] = useState('')
  const [busy, setBusy] = useState('')
  const project = useAsync(() => api.project(projectId), [projectId])
  const tasks = useAsync(() => api.tasks(projectId), [projectId])
  const repositories = useAsync(() => api.repositories(projectId), [projectId])
  const workflows = useAsync(() => api.workflows(projectId), [projectId])
  const runs = useAsync(() => api.runs(organizationId, projectId), [organizationId, projectId])

  const createTask = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setBusy('task')
    setActionError('')
    const form = event.currentTarget
    const data = new FormData(form)
    try {
      await api.createTask(projectId, { title: String(data.get('title')), description: String(data.get('description')) })
      form.reset()
      tasks.reload()
    } catch (value) {
      setActionError(value instanceof Error ? value.message : 'Could not create task')
    } finally {
      setBusy('')
    }
  }

  const updateTask = async (task: Task, status: Task['status']) => {
    setBusy(task.id)
    setActionError('')
    try {
      await api.updateTask(projectId, task.id, status)
      tasks.reload()
    } catch (value) {
      setActionError(value instanceof Error ? value.message : 'Could not update task')
    } finally {
      setBusy('')
    }
  }

  const connectRepository = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setBusy('repository')
    setActionError('')
    const form = event.currentTarget
    const data = new FormData(form)
    try {
      await api.createRepository(projectId, { name: String(data.get('name')), localPath: String(data.get('localPath')), initialize: data.get('initialize') === 'on' })
      form.reset()
      repositories.reload()
    } catch (value) {
      setActionError(value instanceof Error ? value.message : 'Could not connect repository')
    } finally {
      setBusy('')
    }
  }

  const runWorkflow = async (workflowId: string) => {
    setBusy(workflowId)
    setActionError('')
    try {
      const run = await api.runWorkflow(workflowId)
      navigate(`/runs/${run.id}`)
    } catch (value) {
      setActionError(value instanceof Error ? value.message : 'Could not start workflow')
      setBusy('')
    }
  }

  if (project.loading) return <Page><LoadingState label="Loading project" /></Page>
  if (project.error) return <Page><ErrorState error={project.error} retry={project.reload} /></Page>

  return <Page>
    <Link to="/projects" className="back">← Projects</Link>
    <div className="page-title"><div><span className="eyebrow">Project / {project.data?.slug}</span><h1>{project.data?.name}</h1><p>{project.data?.description || 'No project description.'}</p></div><Link className="button primary" to={`/workflows/new?project=${projectId}`}>＋ New workflow</Link></div>
    {actionError && <div className="inline-error" role="alert">{actionError}</div>}
    <div className="two-column">
      <section className="panel panel-stack">
        <div className="panel-head"><div><span className="eyebrow">Delivery queue</span><h2>Tasks</h2></div></div>
        <form className="compact-form" onSubmit={createTask}><label>Task title<input name="title" maxLength={200} required /></label><label>Description<textarea name="description" rows={2} maxLength={10000} /></label><button className="button primary" disabled={busy === 'task'}>{busy === 'task' ? 'Adding…' : 'Add task'}</button></form>
        {tasks.loading ? <LoadingState /> : tasks.error ? <ErrorState error={tasks.error} retry={tasks.reload} /> : tasks.data?.items.length ? <div className="rows">{tasks.data.items.map(task => <div className="row" key={task.id}><div><strong>{task.title}</strong><small>{task.description || 'No description'}</small></div><select aria-label={`Status for ${task.title}`} value={task.status} disabled={busy === task.id} onChange={event => void updateTask(task, event.target.value as Task['status'])}><option value="open">Open</option><option value="in_progress">In progress</option><option value="done">Done</option><option value="canceled">Canceled</option></select></div>)}</div> : <EmptyState title="No tasks" body="Add the first task above." />}
      </section>
      <section className="panel panel-stack">
        <div className="panel-head"><div><span className="eyebrow">Source</span><h2>Repositories</h2></div></div>
        <form className="compact-form" onSubmit={connectRepository}><label>Name<input name="name" maxLength={100} required /></label><label>Workspace path<input name="localPath" placeholder={project.data?.slug || 'my-project'} required /><small>Use a relative path to create a repository managed by ForgeFlow.</small></label><label className="check"><input name="initialize" type="checkbox" defaultChecked /> Initialize a managed Git repository</label><button className="button primary" disabled={busy === 'repository'}>{busy === 'repository' ? 'Connecting…' : 'Connect repository'}</button></form>
        {repositories.loading ? <LoadingState /> : repositories.error ? <ErrorState error={repositories.error} retry={repositories.reload} /> : repositories.data?.items.length ? <div className="rows">{repositories.data.items.map(repo => <div className="row" key={repo.id}><div><strong>{repo.name}</strong><small>{repo.localPath}</small></div><code>{repo.defaultBranch}</code></div>)}</div> : <EmptyState title="No repository connected" body="Create a managed repository above or connect an existing checkout in the workspace volume." />}
      </section>
    </div>
    <div className="two-column">
      <section className="panel panel-stack">
        <div className="panel-head"><div><span className="eyebrow">Automation</span><h2>Workflows</h2></div><Link to={`/workflows/new?project=${projectId}`}>Create →</Link></div>
        {workflows.loading ? <LoadingState /> : workflows.error ? <ErrorState error={workflows.error} retry={workflows.reload} /> : workflows.data?.items.length ? <div className="rows">{workflows.data.items.map(workflow => <div className="row" key={workflow.id}><div><strong>{workflow.name}</strong><small>{workflow.steps.length} step{workflow.steps.length === 1 ? '' : 's'} · version {workflow.version}</small></div><button className="button secondary small" disabled={busy === workflow.id} onClick={() => void runWorkflow(workflow.id)}>{busy === workflow.id ? 'Starting…' : 'Run'}</button></div>)}</div> : <EmptyState title="No workflows" body="Create a workflow to automate a repeatable check." action={<Link className="button secondary" to={`/workflows/new?project=${projectId}`}>Create workflow</Link>} />}
      </section>
      <section className="panel panel-stack">
        <div className="panel-head"><div><span className="eyebrow">Execution</span><h2>Recent runs</h2></div><Link to="/runs">View all →</Link></div>
        {runs.loading ? <LoadingState /> : runs.error ? <ErrorState error={runs.error} retry={runs.reload} /> : runs.data?.items.length ? <div className="rows">{runs.data.items.slice(0, 8).map(run => <Link className="row row-link" to={`/runs/${run.id}`} key={run.id}><div><strong>{run.id.slice(0, 8)}</strong><small>{new Date(run.createdAt).toLocaleString()}</small></div><StatusPill status={run.status} /></Link>)}</div> : <EmptyState title="No workflow runs" body="Start a workflow and its result will appear here." />}
      </section>
    </div>
  </Page>
}
