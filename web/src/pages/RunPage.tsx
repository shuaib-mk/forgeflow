import { type FormEvent, useEffect } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { EmptyState, ErrorState, LoadingState, StatusPill } from '../components/States'
import { useAsync } from '../hooks/useAsync'
import { api } from '../services/api'
import { Page } from './DashboardPage'
import { useAuth } from '../stores/AuthContext'

export function RunsPage() {
  const navigate = useNavigate()
  const { organizationId } = useAuth()
  const runs = useAsync(() => api.runs(organizationId), [organizationId])
  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    navigate(`/runs/${String(new FormData(event.currentTarget).get('run')).trim()}`)
  }
  return (
    <Page>
      <div className="page-title">
        <div>
          <span className="eyebrow">Execution history</span>
          <h1>Workflow runs</h1>
          <p>Review every recent workflow run in your organization.</p>
        </div>
      </div>
      <form className="lookup" onSubmit={submit}>
        <label>
          Open a run by ID
          <input name="run" placeholder="Run UUID" required />
        </label>
        <button className="button secondary">Open run</button>
      </form>
      <section className="panel">
        <div className="panel-head">
          <div>
            <span className="eyebrow">Recent activity</span>
            <h2>All runs</h2>
          </div>
          <button className="button secondary small" onClick={runs.reload}>
            Refresh
          </button>
        </div>
        {runs.loading ? (
          <LoadingState />
        ) : runs.error ? (
          <ErrorState error={runs.error} retry={runs.reload} />
        ) : runs.data?.items.length ? (
          <div className="rows">
            {runs.data.items.map((run) => (
              <Link className="row row-link" to={`/runs/${run.id}`} key={run.id}>
                <div>
                  <strong>{run.id.slice(0, 8)}</strong>
                  <small>
                    Project {run.projectId.slice(0, 8)} · {new Date(run.createdAt).toLocaleString()}
                  </small>
                </div>
                <StatusPill status={run.status} />
              </Link>
            ))}
          </div>
        ) : (
          <EmptyState title="No workflow runs" body="Create a workflow inside a project and select Run." />
        )}
      </section>
    </Page>
  )
}

export function RunDetailPage() {
  const { runId = '' } = useParams()
  const run = useAsync(() => api.run(runId), [runId])
  const logs = useAsync(() => api.logs(runId), [runId])
  const active = run.data?.status === 'queued' || run.data?.status === 'running'

  useEffect(() => {
    if (!active) return
    const timer = window.setInterval(() => {
      run.reload()
      logs.reload()
    }, 1500)
    return () => window.clearInterval(timer)
  }, [active, run, logs])

  if (run.loading && !run.data)
    return (
      <Page>
        <LoadingState label="Loading run" />
      </Page>
    )
  if (run.error)
    return (
      <Page>
        <ErrorState error={run.error} retry={run.reload} />
      </Page>
    )
  return (
    <Page>
      <Link className="back" to="/runs">
        ← Runs
      </Link>
      <div className="page-title">
        <div>
          <span className="eyebrow">Run / {runId.slice(0, 8)}</span>
          <h1>Workflow execution</h1>
          <p>Created {run.data && new Date(run.data.createdAt).toLocaleString()}</p>
        </div>
        {run.data && <StatusPill status={run.data.status} />}
      </div>
      {active && (
        <div className="success" role="status">
          This run is active. Status and logs refresh automatically.
        </div>
      )}
      {run.data?.error && <div className="inline-error">{run.data.error}</div>}
      <section className="panel">
        <div className="panel-head">
          <div>
            <span className="eyebrow">Captured output</span>
            <h2>Logs</h2>
          </div>
          <button
            className="button secondary small"
            onClick={() => {
              run.reload()
              logs.reload()
            }}
          >
            Refresh
          </button>
        </div>
        {logs.loading && !logs.data ? (
          <LoadingState />
        ) : logs.error ? (
          <ErrorState error={logs.error} retry={logs.reload} />
        ) : logs.data?.items.length ? (
          <pre className="logs">{logs.data.items.join('')}</pre>
        ) : (
          <EmptyState
            title="No logs yet"
            body={active ? 'The worker has not produced output yet.' : 'This run completed without captured output.'}
          />
        )}
      </section>
    </Page>
  )
}
