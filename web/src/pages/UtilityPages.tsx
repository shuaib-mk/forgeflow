import { FormEvent, useState } from 'react'
import { EmptyState } from '../components/States'
import { Page } from './DashboardPage'

export function SettingsPage() { const [saved, setSaved] = useState(false); const submit = (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); localStorage.setItem('forgeflow.organization', String(new FormData(event.currentTarget).get('organization'))); setSaved(true) }; return <Page><div className="page-title"><div><span className="eyebrow">Local preferences</span><h1>Settings</h1><p>Browser-only preferences never leave this device.</p></div></div><form className="form-card narrow-card" onSubmit={submit}>{saved && <div className="success" role="status">Settings saved.</div>}<label>Default organization ID<input name="organization" defaultValue={localStorage.getItem('forgeflow.organization') ?? ''} required /><small>Created with your account during registration.</small></label><button className="button primary">Save settings</button></form></Page> }
export function PluginsPage() { return <Page><div className="page-title"><div><span className="eyebrow">Trusted extensions</span><h1>Plugins</h1><p>Compiled integrations for workflow steps, notifications, analytics, and repositories.</p></div></div><EmptyState title="No plugins enabled" body="The built-in run-summary example is available to developers. ForgeFlow intentionally does not load untrusted binaries at runtime." /></Page> }
export function NotFoundPage() { return <Page><EmptyState title="Page not found" body="The route you requested does not exist." /></Page> }

