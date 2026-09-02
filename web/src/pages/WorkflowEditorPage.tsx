import { type FormEvent, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { api } from '../services/api'
import { Page } from './DashboardPage'
import { useAuth } from '../stores/AuthContext'
import { useAsync } from '../hooks/useAsync'
import { ErrorState, LoadingState } from '../components/States'

const starter = `{
  "name": "forgeflow-smoke-check",
  "steps": [
    {
      "id": "hello",
      "name": "Confirm worker execution",
      "command": "sh",
      "args": ["-c", "printf 'ForgeFlow workflow is running successfully.\\n'"],
      "retries": 0
    },
    {
      "id": "git-version",
      "name": "Confirm Git is available",
      "command": "git",
      "args": ["--version"],
      "dependsOn": ["hello"],
      "retries": 0
    }
  ]
}`

export function WorkflowEditorPage() {
  const location = useLocation()
  const navigate = useNavigate()
  const { organizationId } = useAuth()
  const initialProjectId = new URLSearchParams(location.search).get('project') ?? ''
  const [projectId, setProjectId] = useState(initialProjectId)
  const [source, setSource] = useState(starter)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const projects = useAsync(() => api.projects(organizationId), [organizationId])

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    setError('')
    setSaving(true)
    try {
      const definition = JSON.parse(source)
      await api.createWorkflow(projectId, definition)
      navigate(`/projects/${projectId}`)
    } catch (value) {
      setError(value instanceof Error ? value.message : 'Invalid workflow')
    } finally {
      setSaving(false)
    }
  }

  return (
    <Page>
      <div className="page-title">
        <div>
          <span className="eyebrow">Workflow editor</span>
          <h1>Turn checks into a dependable flow.</h1>
          <p>Choose a project, adjust the starter workflow, save it, then run it from the project page.</p>
        </div>
      </div>
      {projects.loading ? (
        <LoadingState label="Loading projects" />
      ) : projects.error ? (
        <ErrorState error={projects.error} retry={projects.reload} />
      ) : (
        <form className="editor" onSubmit={submit}>
          <div className="editor-bar">
            <label>
              Project
              <select value={projectId} onChange={(event) => setProjectId(event.target.value)} required>
                <option value="">Choose a project</option>
                {projects.data?.items.map((project) => (
                  <option key={project.id} value={project.id}>
                    {project.name}
                  </option>
                ))}
              </select>
            </label>
            <span className="pill running">Draft</span>
          </div>
          {error && (
            <div className="inline-error" role="alert">
              {error}
            </div>
          )}
          <label className="sr-only" htmlFor="workflow-source">
            Workflow JSON
          </label>
          <textarea id="workflow-source" value={source} onChange={(event) => setSource(event.target.value)} spellCheck={false} />
          <footer>
            <span>Commands execute in the project's first connected repository.</span>
            <button className="button primary" disabled={!projectId || saving}>
              {saving ? 'Saving…' : 'Save workflow'}
            </button>
          </footer>
        </form>
      )}
    </Page>
  )
}
