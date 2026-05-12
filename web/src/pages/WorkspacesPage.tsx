import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { workspaces as workspacesApi } from '@/lib/api'
import { useWorkspaceStore } from '@/store/workspaceStore'
import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { workspaceSchema, joinWorkspaceSchema, type WorkspaceInput, type JoinWorkspaceInput } from '@/lib/schemas'
import Navbar from '@/components/Navbar'
import WorkspaceCard from '@/components/WorkspaceCard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Globe, Lock, LogIn, Plus } from 'lucide-react'

function CreateWorkspaceDialog({ onCreated }: { onCreated: () => void }) {
  const [open, setOpen] = useState(false)
  const { register, handleSubmit, reset, watch, setValue, formState: { errors, isSubmitting } } =
    useForm<WorkspaceInput>({
      resolver: zodResolver(workspaceSchema),
      defaultValues: { name: '', is_public: false },
    })

  const isPublic = watch('is_public')
  const qc = useQueryClient()
  const { setActiveWorkspace } = useWorkspaceStore()
  const navigate = useNavigate()

  async function onSubmit(data: WorkspaceInput) {
    const res = await workspacesApi.create(data.name, data.is_public)
    qc.invalidateQueries({ queryKey: ['workspaces'] })
    setOpen(false)
    reset()
    setActiveWorkspace(res.data)
    navigate(`/workspaces/${res.data.id}`)
    onCreated()
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm"><Plus className="h-4 w-4 mr-1" />New Workspace</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader><DialogTitle>Create Workspace</DialogTitle></DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 mt-2">
          <div className="space-y-2">
            <Label htmlFor="ws-name">Name</Label>
            <Input id="ws-name" placeholder="My Team" {...register('name')} />
            {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
          </div>

          <div className="flex items-center justify-between rounded-lg border p-3">
            <div className="flex items-center gap-2">
              {isPublic
                ? <Globe className="h-4 w-4 text-muted-foreground" />
                : <Lock className="h-4 w-4 text-muted-foreground" />}
              <div>
                <p className="text-sm font-medium">{isPublic ? 'Public' : 'Private'}</p>
                <p className="text-xs text-muted-foreground">
                  {isPublic
                    ? 'Anyone can discover and join this workspace'
                    : 'Invite code required to join'}
                </p>
              </div>
            </div>
            <button
              type="button"
              role="switch"
              aria-checked={isPublic}
              onClick={() => setValue('is_public', !isPublic)}
              className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${
                isPublic ? 'bg-primary' : 'bg-input'
              }`}
            >
              <span className={`inline-block h-3.5 w-3.5 rounded-full bg-white transition-transform ${
                isPublic ? 'translate-x-4' : 'translate-x-1'
              }`} />
            </button>
          </div>

          <div className="flex justify-end">
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Creating…' : 'Create'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function JoinWorkspaceDialog({ onJoined }: { onJoined: () => void }) {
  const [open, setOpen] = useState(false)
  const { register, handleSubmit, reset, formState: { errors, isSubmitting }, setError } =
    useForm<JoinWorkspaceInput>({ resolver: zodResolver(joinWorkspaceSchema) })

  const qc = useQueryClient()
  const { setActiveWorkspace } = useWorkspaceStore()
  const navigate = useNavigate()

  async function onSubmit(data: JoinWorkspaceInput) {
    try {
      const res = await workspacesApi.join(data.invite_code)
      qc.invalidateQueries({ queryKey: ['workspaces'] })
      setOpen(false)
      reset()
      setActiveWorkspace(res.data)
      navigate(`/workspaces/${res.data.id}`)
      onJoined()
    } catch {
      setError('invite_code', { message: 'Invalid invite code' })
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm"><LogIn className="h-4 w-4 mr-1" />Join with Code</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader><DialogTitle>Join Private Workspace</DialogTitle></DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 mt-2">
          <div className="space-y-2">
            <Label htmlFor="invite">Invite Code</Label>
            <Input id="invite" placeholder="Ab3Xy9Zq" {...register('invite_code')} />
            {errors.invite_code && <p className="text-xs text-destructive">{errors.invite_code.message}</p>}
          </div>
          <div className="flex justify-end">
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Joining…' : 'Join'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default function WorkspacesPage() {
  const qc = useQueryClient()
  const { setActiveWorkspace } = useWorkspaceStore()
  const navigate = useNavigate()

  const { data: myWorkspaces, isLoading } = useQuery({
    queryKey: ['workspaces'],
    queryFn: () => workspacesApi.list().then((r) => r.data),
  })

  const { data: publicWorkspaces } = useQuery({
    queryKey: ['workspaces-public'],
    queryFn: () => workspacesApi.listPublic().then((r) => r.data),
  })

  // Public workspaces the user hasn't joined yet.
  const myIds = new Set((myWorkspaces ?? []).map((w) => w.id))
  const discoverable = (publicWorkspaces ?? []).filter((w) => !myIds.has(w.id))

  function refresh() {
    qc.invalidateQueries({ queryKey: ['workspaces'] })
    qc.invalidateQueries({ queryKey: ['workspaces-public'] })
  }

  async function joinPublic(workspaceId: number) {
    const res = await workspacesApi.joinPublic(workspaceId)
    qc.invalidateQueries({ queryKey: ['workspaces'] })
    setActiveWorkspace(res.data)
    navigate(`/workspaces/${res.data.id}`)
  }

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="max-w-5xl mx-auto px-4 py-8 space-y-10">

        {/* My Workspaces */}
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-semibold">My Workspaces</h1>
            <div className="flex gap-2">
              <JoinWorkspaceDialog onJoined={refresh} />
              <CreateWorkspaceDialog onCreated={refresh} />
            </div>
          </div>

          {isLoading ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {[...Array(3)].map((_, i) => (
                <div key={i} className="h-24 rounded-lg bg-muted animate-pulse" />
              ))}
            </div>
          ) : myWorkspaces && myWorkspaces.length > 0 ? (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {myWorkspaces.map((w) => <WorkspaceCard key={w.id} workspace={w} />)}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <p className="text-muted-foreground">No workspaces yet. Create one or join with an invite code.</p>
            </div>
          )}
        </section>

        {/* Discover Public Workspaces */}
        {discoverable.length > 0 && (
          <section className="space-y-4">
            <h2 className="text-lg font-semibold">Discover Public Workspaces</h2>
            <div className="divide-y border rounded-lg">
              {discoverable.map((w) => (
                <div key={w.id} className="flex items-center justify-between px-4 py-3">
                  <div className="flex items-center gap-2">
                    <Globe className="h-4 w-4 text-muted-foreground" />
                    <span className="font-medium text-sm">{w.name}</span>
                    <Badge variant="secondary" className="text-xs">public</Badge>
                  </div>
                  <Button size="sm" onClick={() => joinPublic(w.id)}>Join</Button>
                </div>
              ))}
            </div>
          </section>
        )}
      </main>
    </div>
  )
}
