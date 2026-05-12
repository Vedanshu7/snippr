import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { isAuthenticated } from '@/lib/auth'
import LoginPage from '@/pages/LoginPage'
import RegisterPage from '@/pages/RegisterPage'
import DashboardPage from '@/pages/DashboardPage'
import NewSnippetPage from '@/pages/NewSnippetPage'
import SnippetDetailPage from '@/pages/SnippetDetailPage'
import EditSnippetPage from '@/pages/EditSnippetPage'
import PublicSnippetPage from '@/pages/PublicSnippetPage'

const qc = new QueryClient()

function RequireAuth({ children }: { children: React.ReactNode }) {
  if (!isAuthenticated()) return <Navigate to="/login" replace />
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
          <Route
            path="/dashboard"
            element={<RequireAuth><DashboardPage /></RequireAuth>}
          />
          <Route
            path="/snippets/new"
            element={<RequireAuth><NewSnippetPage /></RequireAuth>}
          />
          <Route
            path="/snippets/:id"
            element={<RequireAuth><SnippetDetailPage /></RequireAuth>}
          />
          <Route
            path="/snippets/:id/edit"
            element={<RequireAuth><EditSnippetPage /></RequireAuth>}
          />
          <Route path="*" element={<Navigate to="/dashboard" replace />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
