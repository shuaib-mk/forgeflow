import { NavLink, Outlet } from 'react-router-dom'
import { useAuth } from '../stores/AuthContext'

const navigation = [
  ['Overview', '/', '⌁'], ['Projects', '/projects', '▦'], ['Workflow editor', '/workflows/new', '⌘'], ['Runs', '/runs', '▶'], ['Analytics', '/analytics', '↗'], ['Audit log', '/audit', '≋'], ['Plugins', '/plugins', '⬡'], ['Settings', '/settings', '⚙'],
]

export function AppLayout() {
  const { user, organizations, organizationId, selectOrganization, logout } = useAuth()
  return <div className="shell">
    <aside className="sidebar">
      <a className="brand" href="/" aria-label="ForgeFlow home"><span className="brand-mark">F</span><span>Forge<span>Flow</span></span></a>
      {organizations.length > 0 && <label className="organization-picker"><span>Organization</span><select value={organizationId} onChange={event => selectOrganization(event.target.value)}>{organizations.map(item => <option key={item.organization.id} value={item.organization.id}>{item.organization.name}</option>)}</select></label>}
      <nav aria-label="Primary navigation">{navigation.map(([label, path, icon]) => <NavLink key={path} to={path} end={path === '/'}><span aria-hidden="true">{icon}</span>{label}</NavLink>)}</nav>
      <div className="sidebar-foot"><div className="avatar">{user?.displayName.slice(0, 2).toUpperCase()}</div><div className="identity"><strong>{user?.displayName}</strong><span>{user?.email}</span></div><button className="icon-button" title="Sign out" aria-label="Sign out" onClick={() => void logout()}>↪</button></div>
    </aside>
    <main className="main"><Outlet /></main>
  </div>
}
