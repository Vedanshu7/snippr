import { useEffect, useState } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useAuthStore } from '@/store/authStore'
import { auth } from '@/lib/api'
import LoginPage from '@/pages/LoginPage'
import RegisterPage from '@/pages/RegisterPage'
import DashboardPage from '@/pages/DashboardPage'
import NewSnippetPage from '@/pages/NewSnippetPage'
import SnippetDetailPage from '@/pages/SnippetDetailPage'
import EditSnippetPage from '@/pages/EditSnippetPage'
import PublicSnippetPage from '@/pages/PublicSnippetPage'
import WorkspacesPage from '@/pages/WorkspacesPage'
import WorkspacePage from '@/pages/WorkspacePage'
import WorkspaceSettingsPage from '@/pages/WorkspaceSettingsPage'

const qc = new QueryClient()

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { user, setUser } = useAuthStore()
  const [checking, setChecking] = useState(!user)

  useEffect(() => {
    if (user) { setChecking(false); return }
    // No user in store — check if a session cookie is still valid.
    auth.me()
      .then((res) => setUser(res.data))
      .catch(() => {})
      .finally(() => setChecking(false))
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  if (checking) return null // brief flash while we check the cookie
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <QueryClientProvider client={qc}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/s/:slug" element={<PublicSnippetPage />} />
          <Route path="/dashboard" element={<RequireAuth><DashboardPage /></RequireAuth>} />
          <Route path="/snippets/new" element={<RequireAuth><NewSnippetPage /></RequireAuth>} />
          <Route path="/snippets/:id" element={<RequireAuth><SnippetDetailPage /></RequireAuth>} />
          <Route path="/snippets/:id/edit" element={<RequireAuth><EditSnippetPage /></RequireAuth>} />
          <Route path="/workspaces" element={<RequireAuth><WorkspacesPage /></RequireAuth>} />
          <Route path="/workspaces/:id" element={<RequireAuth><WorkspacePage /></RequireAuth>} />
          <Route path="/workspaces/:id/settings" element={<RequireAuth><WorkspaceSettingsPage /></RequireAuth>} />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
