import { Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from './layouts/AppLayout'
import { useAuth } from './stores/AuthContext'
import { LoadingState } from './components/States'
import { LoginPage } from './pages/LoginPage'
import { DashboardPage } from './pages/DashboardPage'
import { NewProjectPage, ProjectsPage } from './pages/ProjectsPage'
import { ProjectDetailPage } from './pages/ProjectDetailPage'
import { WorkflowEditorPage } from './pages/WorkflowEditorPage'
import { RunDetailPage, RunsPage } from './pages/RunPage'
import { AnalyticsPage, AuditPage } from './pages/InsightsPages'
import { NotFoundPage, PluginsPage, SettingsPage } from './pages/UtilityPages'

export function App() {
  const { user, loading } = useAuth()
  if (loading)
    return (
      <div className="full-state">
        <LoadingState label="Restoring session" />
      </div>
    )
  if (!user)
    return (
      <Routes>
        <Route path="*" element={<LoginPage />} />
      </Routes>
    )
  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route index element={<DashboardPage />} />
        <Route path="projects" element={<ProjectsPage />} />
        <Route path="projects/new" element={<NewProjectPage />} />
        <Route path="projects/:projectId" element={<ProjectDetailPage />} />
        <Route path="workflows/new" element={<WorkflowEditorPage />} />
        <Route path="runs" element={<RunsPage />} />
        <Route path="runs/:runId" element={<RunDetailPage />} />
        <Route path="analytics" element={<AnalyticsPage />} />
        <Route path="audit" element={<AuditPage />} />
        <Route path="plugins" element={<PluginsPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="404" element={<NotFoundPage />} />
        <Route path="*" element={<Navigate to="/404" replace />} />
      </Route>
    </Routes>
  )
}
