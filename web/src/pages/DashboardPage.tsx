import { Link } from 'react-router-dom'
import { api } from '../services/api'
import { useAsync } from '../hooks/useAsync'
import { EmptyState, ErrorState, LoadingState } from '../components/States'

const organizationId = () => localStorage.getItem('forgeflow.organization') ?? ''
export function DashboardPage() {
  const analytics = useAsync(() => api.analytics(organizationId()), [])
  const projects = useAsync(() => api.projects(organizationId()), [])
  if (!organizationId()) return <Page><EmptyState title="Choose your organization" body="Add the organization ID in Settings to load your workspace." action={<Link className="button primary" to="/settings">Open settings</Link>} /></Page>
  return <Page><div className="page-title"><div><span className="eyebrow">Engineering command center</span><h1>Good afternoon.</h1><p>Here’s what is moving across your workspace.</p></div><Link className="button primary" to="/projects/new">＋ New project</Link></div>
    {analytics.loading ? <LoadingState label="Loading workspace" /> : analytics.error ? <ErrorState error={analytics.error} retry={analytics.reload} /> : <div className="metric-grid">{Object.entries(analytics.data ?? {}).map(([key, value]) => <article className="metric" key={key}><span>{key.replace(/([A-Z])/g, ' $1')}</span><strong>{value}</strong><small>{key === 'failedRuns' && value > 0 ? 'Needs attention' : 'Current total'}</small></article>)}</div>}
    <section className="panel"><div className="panel-head"><div><span className="eyebrow">Recently updated</span><h2>Projects</h2></div><Link to="/projects">View all →</Link></div>{projects.loading ? <LoadingState /> : projects.error ? <ErrorState error={projects.error} retry={projects.reload} /> : projects.data?.items.length ? <div className="project-list">{projects.data.items.slice(0, 5).map(project => <Link to={`/projects/${project.id}`} key={project.id}><span className="project-icon">{project.name.slice(0, 1)}</span><div><strong>{project.name}</strong><small>{project.description || 'No description yet'}</small></div><time>{new Date(project.updatedAt).toLocaleDateString()}</time><span>→</span></Link>)}</div> : <EmptyState title="No projects yet" body="Create a project to connect a repository and run your first workflow." />}</section>
  </Page>
}

export function Page({ children }: { children: React.ReactNode }) { return <div className="page">{children}</div> }

