import { Link, useNavigate } from 'react-router-dom'
import { auth } from '@/lib/api'
import { useAuthStore } from '@/store/authStore'
import { useThemeStore } from '@/store/themeStore'
import { Button } from '@/components/ui/button'
import { Code2, LogOut, Moon, Plus, Sun } from 'lucide-react'

export default function Navbar() {
  const navigate = useNavigate()
  const { logout } = useAuthStore()
  const { theme, toggle } = useThemeStore()

  async function handleLogout() {
    try { await auth.logout() } catch { /* ignore */ }
    logout()
    navigate('/login')
  }

  return (
    <header className="border-b bg-background sticky top-0 z-10">
      <div className="max-w-5xl mx-auto px-4 h-14 flex items-center justify-between">
        <Link to="/dashboard" className="flex items-center gap-2 font-semibold">
          <Code2 className="h-5 w-5" />
          Snippr
        </Link>
        <div className="flex items-center gap-2">
          <Button variant="default" size="sm" asChild>
            <Link to="/snippets/new"><Plus className="h-4 w-4 mr-1" />New</Link>
          </Button>
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
