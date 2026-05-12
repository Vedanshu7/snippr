import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { workspaces as workspacesApi, snippets as snippetsApi } from '@/lib/api'
import { useWorkspaceStore } from '@/store/workspaceStore'
import { useAuthStore } from '@/store/authStore'
import Navbar from '@/components/Navbar'
import LiveBoard from '@/components/LiveBoard'
import SnippetCard from '@/components/SnippetCard'
import { Button } from '@/components/ui/button'
import { Settings, Zap, List } from 'lucide-react'

type Tab = 'snippets' | 'board'

export default function WorkspacePage() {
  const { id } = useParams<{ id: string }>()
  const workspaceId = Number(id)
  const [tab, setTab] = useState<Tab>('snippets')
  const { setActiveWorkspace } = useWorkspaceStore()
  const { user } = useAuthStore()

  const { data: workspace } = useQuery({
    queryKey: ['workspace', workspaceId],
    queryFn: () => workspacesApi.get(workspaceId).then((r) => r.data),
    enabled: !!workspaceId,
  })

  const { data: snippetList, isLoading } = useQuery({
    queryKey: ['snippets', 'workspace', workspaceId],
    queryFn: () => snippetsApi.list(undefined, undefined, workspaceId).then((r) => r.data),
    enabled: !!workspaceId,
  })

  // Keep the workspace store in sync when navigating directly via URL.
  useEffect(() => {
    if (workspace) setActiveWorkspace(workspace)
    return () => setActiveWorkspace(null)
  }, [workspace]) // eslint-disable-line react-hooks/exhaustive-deps

  if (!workspace) return null

  const isOwner = workspace.owner_id === user?.id

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="max-w-5xl mx-auto px-4 py-8 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">{workspace.name}</h1>
          </div>
          <div className="flex gap-2">
            {isOwner && (
              <Button variant="outline" size="sm" asChild>
                <Link to={`/workspaces/${workspaceId}/settings`}>
                  <Settings className="h-4 w-4 mr-1" />Settings
                </Link>
              </Button>
            )}
          </div>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 border-b">
          <button
            onClick={() => setTab('snippets')}
            className={`px-4 py-2 text-sm font-medium flex items-center gap-1.5 border-b-2 transition-colors ${
              tab === 'snippets'
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <List className="h-3.5 w-3.5" />Snippets
          </button>
          <button
            onClick={() => setTab('board')}
            className={`px-4 py-2 text-sm font-medium flex items-center gap-1.5 border-b-2 transition-colors ${
              tab === 'board'
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <Zap className="h-3.5 w-3.5" />Live Board
          </button>
        </div>

        {tab === 'snippets' ? (
          <>
            {isLoading ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {[...Array(6)].map((_, i) => (
                  <div key={i} className="h-48 rounded-lg bg-muted animate-pulse" />
                ))}
              </div>
            ) : snippetList && snippetList.length > 0 ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {snippetList.map((s) => <SnippetCard key={s.id} snippet={s} />)}
              </div>
            ) : (
              <div className="flex flex-col items-center justify-center py-20 text-center">
                <p className="text-muted-foreground">No snippets yet. Create one and scope it to this workspace.</p>
              </div>
            )}
          </>
        ) : (
          <LiveBoard workspaceId={workspaceId} initialSnippets={snippetList ?? []} />
        )}
      </main>
    </div>
  )
}
