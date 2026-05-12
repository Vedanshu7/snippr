import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { workspaces as workspacesApi } from '@/lib/api'
import { useAuthStore } from '@/store/authStore'
import { useWorkspaceStore } from '@/store/workspaceStore'
import Navbar from '@/components/Navbar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Copy, Globe, Lock, LogOut, RefreshCw, Trash2, UserMinus } from 'lucide-react'

export default function WorkspaceSettingsPage() {
  const { id } = useParams<{ id: string }>()
  const workspaceId = Number(id)
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { user } = useAuthStore()
  const { setActiveWorkspace } = useWorkspaceStore()
  const [copied, setCopied] = useState(false)

  const { data: workspace, isLoading } = useQuery({
    queryKey: ['workspace', workspaceId],
    queryFn: () => workspacesApi.get(workspaceId).then((r) => r.data),
    enabled: !!workspaceId,
  })

  const { data: members } = useQuery({
    queryKey: ['workspace-members', workspaceId],
    queryFn: () => workspacesApi.members(workspaceId).then((r) => r.data),
    enabled: !!workspaceId,
  })

  // Edit form local state (controlled, not react-hook-form — simpler for a settings page).
  const [editName, setEditName] = useState('')
  const [editPublic, setEditPublic] = useState(false)
  const [editing, setEditing] = useState(false)

  function startEdit() {
    setEditName(workspace?.name ?? '')
    setEditPublic(workspace?.is_public ?? false)
    setEditing(true)
  }

  const updateMutation = useMutation({
    mutationFn: () => workspacesApi.update(workspaceId, { name: editName, is_public: editPublic }),
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ['workspace', workspaceId] })
      qc.invalidateQueries({ queryKey: ['workspaces'] })
      setActiveWorkspace(res.data)
      setEditing(false)
    },
  })

  const rotateMutation = useMutation({
    mutationFn: () => workspacesApi.rotateInvite(workspaceId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace', workspaceId] })
      qc.invalidateQueries({ queryKey: ['workspaces'] })
    },
  })

  const leaveMutation = useMutation({
    mutationFn: () => workspacesApi.leave(workspaceId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspaces'] })
      setActiveWorkspace(null)
      navigate('/workspaces')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => workspacesApi.delete(workspaceId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspaces'] })
      setActiveWorkspace(null)
      navigate('/workspaces')
    },
  })

  const removeMemberMutation = useMutation({
    mutationFn: (userId: number) => workspacesApi.removeMember(workspaceId, userId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['workspace-members', workspaceId] })
    },
  })

  function copyCode() {
    if (!workspace) return
    navigator.clipboard.writeText(workspace.invite_code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  if (isLoading || !workspace) return null

  // Derive isAdmin from the members list (role column is the source of truth).
  const isAdmin = members?.find((m) => m.user_id === user?.id)?.role === 'admin'

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="max-w-2xl mx-auto px-4 py-8 space-y-6">
        <h1 className="text-2xl font-semibold">{workspace.name} — Settings</h1>

        {/* Edit workspace (admin only) */}
        {isAdmin && (
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">Workspace Details</CardTitle>
                {!editing && (
                  <Button variant="ghost" size="sm" onClick={startEdit}>Edit</Button>
                )}
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {editing ? (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="edit-name">Name</Label>
                    <Input
                      id="edit-name"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                    />
                  </div>

                  <div className="flex items-center justify-between rounded-lg border p-3">
                    <div className="flex items-center gap-2">
                      {editPublic
                        ? <Globe className="h-4 w-4 text-muted-foreground" />
                        : <Lock className="h-4 w-4 text-muted-foreground" />}
                      <div>
                        <p className="text-sm font-medium">{editPublic ? 'Public' : 'Private'}</p>
                        <p className="text-xs text-muted-foreground">
                          {editPublic
                            ? 'Anyone can discover and join'
                            : 'Invite code required to join'}
                        </p>
                      </div>
                    </div>
                    <button
                      type="button"
                      role="switch"
                      aria-checked={editPublic}
                      onClick={() => setEditPublic(!editPublic)}
                      className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${
                        editPublic ? 'bg-primary' : 'bg-input'
                      }`}
                    >
                      <span className={`inline-block h-3.5 w-3.5 rounded-full bg-white transition-transform ${
                        editPublic ? 'translate-x-4' : 'translate-x-1'
                      }`} />
                    </button>
                  </div>

                  <div className="flex gap-2 justify-end">
                    <Button variant="outline" size="sm" onClick={() => setEditing(false)}>Cancel</Button>
                    <Button
                      size="sm"
                      onClick={() => updateMutation.mutate()}
                      disabled={updateMutation.isPending || !editName.trim()}
                    >
                      {updateMutation.isPending ? 'Saving…' : 'Save'}
                    </Button>
                  </div>
                </>
              ) : (
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  {workspace.is_public
                    ? <><Globe className="h-4 w-4" /> Public workspace</>
                    : <><Lock className="h-4 w-4" /> Private workspace</>}
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Invite code */}
        <Card>
          <CardHeader><CardTitle className="text-base">Invite Code</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center gap-2">
              <code className="flex-1 font-mono text-lg bg-muted px-3 py-2 rounded">
                {workspace.invite_code}
              </code>
              <Button variant="outline" size="icon" onClick={copyCode} aria-label="Copy">
                <Copy className="h-4 w-4" />
              </Button>
              {isAdmin && (
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => rotateMutation.mutate()}
                  disabled={rotateMutation.isPending}
                  aria-label="Rotate invite code"
                >
                  <RefreshCw className={`h-4 w-4 ${rotateMutation.isPending ? 'animate-spin' : ''}`} />
                </Button>
              )}
            </div>
            {copied && <p className="text-xs text-muted-foreground">Copied!</p>}
            <p className="text-sm text-muted-foreground">
              Share this code with teammates so they can join this workspace.
            </p>
          </CardContent>
        </Card>

        {/* Members list */}
        <Card>
          <CardHeader><CardTitle className="text-base">Members</CardTitle></CardHeader>
          <CardContent>
            <ul className="divide-y">
              {members?.map((m) => (
                <li key={m.user_id} className="py-2.5 flex items-center justify-between text-sm">
                  <div className="flex items-center gap-2">
                    <span>{m.email}</span>
                    <Badge variant={m.role === 'admin' ? 'default' : 'secondary'} className="text-xs">
                      {m.role}
                    </Badge>
                  </div>
                  {/* Admin can remove any non-admin member (can't remove themselves) */}
                  {isAdmin && m.role !== 'admin' && m.user_id !== user?.id && (
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-7 px-2 text-destructive hover:text-destructive"
                      onClick={() => removeMemberMutation.mutate(m.user_id)}
                      disabled={removeMemberMutation.isPending}
                    >
                      <UserMinus className="h-3.5 w-3.5" />
                    </Button>
                  )}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        {/* Member: leave workspace */}
        {!isAdmin && (
          <div className="pt-2 border-t">
            <Button
              variant="outline"
              size="sm"
              onClick={() => leaveMutation.mutate()}
              disabled={leaveMutation.isPending}
            >
              <LogOut className="h-4 w-4 mr-1" />
              {leaveMutation.isPending ? 'Leaving…' : 'Leave Workspace'}
            </Button>
          </div>
        )}

        {/* Admin: danger zone */}
        {isAdmin && (
          <div className="border border-destructive/30 rounded-lg p-4 space-y-3">
            <h3 className="text-sm font-semibold text-destructive">Danger Zone</h3>
            <p className="text-sm text-muted-foreground">
              Permanently deletes this workspace. All snippets will revert to personal scope.
            </p>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => deleteMutation.mutate()}
              disabled={deleteMutation.isPending}
            >
              <Trash2 className="h-4 w-4 mr-1" />
              {deleteMutation.isPending ? 'Deleting…' : 'Delete Workspace'}
            </Button>
          </div>
        )}
      </main>
    </div>
  )
}
