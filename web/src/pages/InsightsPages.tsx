import { api } from '../services/api'
import { useAsync } from '../hooks/useAsync'
import { EmptyState, ErrorState, LoadingState } from '../components/States'
import { Page } from './DashboardPage'
import { useAuth } from '../stores/AuthContext'

export function AnalyticsPage() {
  const { organizationId } = useAuth()
  const result = useAsync(() => api.analytics(organizationId), [organizationId])
  return (
    <Page>
      <div className="page-title">
        <div>
          <span className="eyebrow">Project intelligence</span>
          <h1>Analytics</h1>
          <p>Operational signals calculated from durable workflow and task state.</p>
        </div>
      </div>
      {result.loading ? (
        <LoadingState />
      ) : result.error ? (
        <ErrorState error={result.error} retry={result.reload} />
      ) : (
        <div className="metric-grid">
          {Object.entries(result.data ?? {}).map(([key, value]) => (
            <article className="metric" key={key}>
              <span>{key.replace(/([A-Z])/g, ' $1')}</span>
              <strong>{value}</strong>
              <small>Current total</small>
            </article>
          ))}
        </div>
      )}
    </Page>
  )
}
export function AuditPage() {
  const { organizationId } = useAuth()
  const result = useAsync(() => api.audit(organizationId), [organizationId])
  return (
    <Page>
      <div className="page-title">
        <div>
          <span className="eyebrow">Accountability</span>
          <h1>Audit log</h1>
          <p>A durable trail of material changes, actors, and request IDs.</p>
        </div>
      </div>
      <section className="panel">
        {result.loading ? (
          <LoadingState />
        ) : result.error ? (
          <ErrorState error={result.error} retry={result.reload} />
        ) : result.data?.items.length ? (
          <div className="audit-list">
            {result.data.items.map((event) => (
              <article key={event.id}>
                <span className="audit-dot" />
                <div>
                  <strong>{event.action}</strong>
                  <small>
                    {event.resourceType} · {event.resourceId.slice(0, 8)}
                  </small>
                </div>
                <code>{event.requestId.slice(0, 12)}</code>
                <time>{new Date(event.createdAt).toLocaleString()}</time>
              </article>
            ))}
          </div>
        ) : (
          <EmptyState title="No audit events" body="Events appear here as your team changes projects and workflows." />
        )}
      </section>
    </Page>
  )
}
