import { useEffect, useState } from 'react'
import type { Snippet } from '@/lib/api'

export type EditingPresence = Record<number, { userID: number; email: string }>

interface BoardPayload {
  type: string
  payload: Snippet | { id: number } | { snippet_id: number; user_id: number; email: string; editing: boolean }
}

export function useWorkspaceBoard(workspaceId: number, initial: Snippet[] = []) {
  const [snippets, setSnippets] = useState<Snippet[]>(initial)
  const [editingBy, setEditingBy] = useState<EditingPresence>({})

  useEffect(() => {
    setSnippets(initial)
  }, [initial.length]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!workspaceId) return

    const es = new EventSource(`/api/workspaces/${workspaceId}/board`, {
      withCredentials: true,
    })

    es.addEventListener('snippet_added', (e: MessageEvent) => {
      const { payload } = JSON.parse(e.data) as BoardPayload
      setSnippets((prev) => [payload as Snippet, ...prev])
    })

    es.addEventListener('snippet_updated', (e: MessageEvent) => {
      const { payload } = JSON.parse(e.data) as BoardPayload
      const updated = payload as Snippet
      setSnippets((prev) => prev.map((s) => (s.id === updated.id ? updated : s)))
    })

    es.addEventListener('snippet_deleted', (e: MessageEvent) => {
      const { payload } = JSON.parse(e.data) as BoardPayload
      const { id } = payload as { id: number }
      setSnippets((prev) => prev.filter((s) => s.id !== id))
    })

    es.addEventListener('editing_start', (e: MessageEvent) => {
      const { payload } = JSON.parse(e.data) as BoardPayload
      const p = payload as { snippet_id: number; user_id: number; email: string }
      setEditingBy((prev) => ({ ...prev, [p.snippet_id]: { userID: p.user_id, email: p.email } }))
    })

    es.addEventListener('editing_stop', (e: MessageEvent) => {
      const { payload } = JSON.parse(e.data) as BoardPayload
      const p = payload as { snippet_id: number }
      setEditingBy((prev) => {
        const next = { ...prev }
        delete next[p.snippet_id]
        return next
      })
    })

    return () => es.close()
  }, [workspaceId])

  return { snippets, editingBy }
}
