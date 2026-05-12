import { useEffect, useRef, useState } from 'react'
import type { Snippet } from '@/lib/api'
import { snippets as snippetsApi, workspaces as workspacesApi } from '@/lib/api'
import type { EditingPresence } from '@/hooks/useWorkspaceBoard'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Pencil, CheckCircle2, AlertCircle, Loader2 } from 'lucide-react'

interface Props {
  snippet: Snippet
  workspaceId: number
  currentUserId: number
  editingBy: EditingPresence
}

type SaveStatus = 'idle' | 'saving' | 'saved' | 'error'

const DEBOUNCE_MS = 800

export default function CollaborativeSnippetCard({
  snippet,
  workspaceId,
  currentUserId,
  editingBy,
}: Props) {
  const [localTitle, setLocalTitle] = useState(snippet.title)
  const [localContent, setLocalContent] = useState(snippet.content)
  const [saveStatus, setSaveStatus] = useState<SaveStatus>('idle')

  const dirtyRef = useRef(false)
  const debounceTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const cardRef = useRef<HTMLDivElement>(null)
  const isEditingRef = useRef(false)

  // Apply external SSE updates only when the user has no pending local edits.
  useEffect(() => {
    if (dirtyRef.current) return
    setLocalTitle(snippet.title)
    setLocalContent(snippet.content)
  }, [snippet.title, snippet.content, snippet.updated_at])

  function triggerSave(title: string, content: string) {
    if (debounceTimer.current) clearTimeout(debounceTimer.current)
    debounceTimer.current = setTimeout(async () => {
      setSaveStatus('saving')
      try {
        await snippetsApi.update(snippet.id, {
          title,
          content,
          language: snippet.language,
          is_public: snippet.is_public,
          tags: snippet.tags,
          workspace_id: snippet.workspace_id,
        })
        dirtyRef.current = false
        setSaveStatus('saved')
        setTimeout(() => setSaveStatus('idle'), 2000)
      } catch {
        setSaveStatus('error')
      }
    }, DEBOUNCE_MS)
  }

  function handleTitleChange(e: React.ChangeEvent<HTMLInputElement>) {
    dirtyRef.current = true
    setSaveStatus('idle')
    setLocalTitle(e.target.value)
    triggerSave(e.target.value, localContent)
  }

  function handleContentChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
    dirtyRef.current = true
    setSaveStatus('idle')
    setLocalContent(e.target.value)
    triggerSave(localTitle, e.target.value)
  }

  function handleFocus() {
    if (isEditingRef.current) return
    isEditingRef.current = true
    workspacesApi.announceEditing(workspaceId, snippet.id, true).catch(() => {})
  }

  function handleBlur(e: React.FocusEvent) {
    // If focus moved to another element inside this card, stay in editing state.
    if (cardRef.current?.contains(e.relatedTarget as Node)) return
    isEditingRef.current = false
    workspacesApi.announceEditing(workspaceId, snippet.id, false).catch(() => {})
  }

  const presence = editingBy[snippet.id]
  const someoneElseEditing = presence && presence.userID !== currentUserId

  return (
    <Card ref={cardRef} className="flex flex-col h-full">
      <CardHeader className="pb-2 space-y-1">
        <div className="flex items-center gap-2">
          <Input
            value={localTitle}
            onChange={handleTitleChange}
            onFocus={handleFocus}
            onBlur={handleBlur}
            className="font-semibold text-sm border-transparent bg-transparent px-0 h-auto focus-visible:border-input focus-visible:bg-background focus-visible:px-3"
            placeholder="Snippet title"
          />
          <Badge variant="secondary" className="shrink-0 text-xs">{snippet.language}</Badge>
        </div>
        {someoneElseEditing && (
          <div className="flex items-center gap-1 text-xs text-amber-600 dark:text-amber-400">
            <Pencil className="h-3 w-3" />
            <span>{presence.email} is editing…</span>
          </div>
        )}
      </CardHeader>

      <CardContent className="flex flex-col flex-1 gap-2">
        <Textarea
          value={localContent}
          onChange={handleContentChange}
          onFocus={handleFocus}
          onBlur={handleBlur}
          className="flex-1 min-h-[120px] resize-none font-mono text-xs leading-relaxed"
          placeholder="Snippet content…"
        />

        <div className="flex items-center gap-1.5 text-xs text-muted-foreground h-4">
          {saveStatus === 'saving' && (
            <>
              <Loader2 className="h-3 w-3 animate-spin" />
              <span>Saving…</span>
            </>
          )}
          {saveStatus === 'saved' && (
            <>
              <CheckCircle2 className="h-3 w-3 text-green-500" />
              <span className="text-green-600 dark:text-green-400">Saved</span>
            </>
          )}
          {saveStatus === 'error' && (
            <>
              <AlertCircle className="h-3 w-3 text-destructive" />
              <span className="text-destructive">Save failed</span>
            </>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
