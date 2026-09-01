import { FormEvent, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { api } from '../services/api'
import { Page } from './DashboardPage'

const starter = `{
  "name": "test-and-build",
  "steps": [
    { "id": "test", "name": "Run tests", "command": "go", "args": ["test", "./..."], "timeout": 600000000000 },
    { "id": "build", "name": "Build", "command": "go", "args": ["build", "./..."], "dependsOn": ["test"], "timeout": 600000000000 }
  ]
}`

export function WorkflowEditorPage() { const location = useLocation(); const navigate = useNavigate(); const projectId = new URLSearchParams(location.search).get('project') ?? ''; const [source, setSource] = useState(starter); const [error, setError] = useState(''); const submit = async (event: FormEvent) => { event.preventDefault(); setError(''); try { const definition = JSON.parse(source); await api.createWorkflow(projectId, definition); navigate(`/projects/${projectId}`) } catch (value) { setError(value instanceof Error ? value.message : 'Invalid workflow') } }; return <Page><div className="page-title"><div><span className="eyebrow">Workflow editor</span><h1>Turn checks into a dependable flow.</h1><p>Steps use direct commands, explicit arguments, timeouts, dependencies, and bounded retries.</p></div></div><form className="editor" onSubmit={submit}><div className="editor-bar"><span>workflow.json</span><span className="pill running">Draft</span></div>{error && <div className="inline-error" role="alert">{error}</div>}<label className="sr-only" htmlFor="workflow-source">Workflow JSON</label><textarea id="workflow-source" value={source} onChange={event => setSource(event.target.value)} spellCheck={false} /><footer><span>Validate offline with <code>forgeflow workflow validate</code></span><button className="button primary" disabled={!projectId}>Save workflow</button></footer></form></Page> }

