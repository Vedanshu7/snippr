import { Link, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { auth, workspaces as workspacesApi } from '@/lib/api'
import { useAuthStore } from '@/store/authStore'
import { useThemeStore } from '@/store/themeStore'
import { useWorkspaceStore } from '@/store/workspaceStore'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Code2, ChevronDown, LogOut, Moon, Plus, Settings, Sun, Users } from 'lucide-react'

export default function Navbar() {
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()
  const { theme, toggle } = useThemeStore()
  const { activeWorkspace, setActiveWorkspace } = useWorkspaceStore()

  const { data: workspaceList } = useQuery({
    queryKey: ['workspaces'],
    queryFn: () => workspacesApi.list().then((r) => r.data),
  })

  async function handleLogout() {
    try { await auth.logout() } catch { /* ignore */ }
    logout()
    navigate('/login')
  }

  function selectWorkspace(id: number | null) {
    if (id === null) {
      setActiveWorkspace(null)
      navigate('/dashboard')
    } else {
      const ws = workspaceList?.find((w) => w.id === id) ?? null
      setActiveWorkspace(ws)
      navigate(`/workspaces/${id}`)
    }
  }

  const newPath = activeWorkspace
    ? `/snippets/new`
    : '/snippets/new'

  return (
    <header className="border-b bg-background sticky top-0 z-10">
      <div className="max-w-5xl mx-auto px-4 h-14 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Link to="/dashboard" className="flex items-center gap-2 font-semibold">
            <Code2 className="h-5 w-5" />
            Snippr
          </Link>

          {/* Workspace switcher */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" className="gap-1 text-muted-foreground">
                {activeWorkspace ? (
                  <><Users className="h-3.5 w-3.5" />{activeWorkspace.name}</>
                ) : (
                  'Personal'
                )}
                <ChevronDown className="h-3.5 w-3.5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-52">
              <DropdownMenuItem onClick={() => selectWorkspace(null)}>
                Personal
              </DropdownMenuItem>
              {workspaceList && workspaceList.length > 0 && (
                <>
                  <DropdownMenuSeparator />
                  {workspaceList.map((w) => (
                    <DropdownMenuItem key={w.id} onClick={() => selectWorkspace(w.id)}>
                      <Users className="h-3.5 w-3.5 mr-2" />
                      {w.name}
                    </DropdownMenuItem>
                  ))}
                </>
              )}
              <DropdownMenuSeparator />
              <DropdownMenuItem asChild>
                <Link to="/workspaces">Manage Workspaces →</Link>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="default" size="sm" asChild>
            <Link to={newPath}><Plus className="h-4 w-4 mr-1" />New</Link>
          </Button>
          {activeWorkspace && activeWorkspace.owner_id === user?.id && (
            <Button variant="ghost" size="icon" asChild aria-label="Workspace settings">
              <Link to={`/workspaces/${activeWorkspace.id}/settings`}>
                <Settings className="h-4 w-4" />
              </Link>
            </Button>
          )}
          <Button variant="ghost" size="icon" onClick={toggle} aria-label="Toggle theme">
            {theme === 'light' ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
          </Button>
          <Button variant="ghost" size="sm" onClick={handleLogout}>
            <LogOut className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </header>
  )
}
