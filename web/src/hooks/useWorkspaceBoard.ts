import { useEffect, useState } from 'react'
import type { Snippet } from '@/lib/api'

interface BoardPayload {
  type: string
  payload: Snippet | { id: number }
}

export function useWorkspaceBoard(workspaceId: number, initial: Snippet[] = []) {
  const [snippets, setSnippets] = useState<Snippet[]>(initial)

  // Keep snippets in sync when the initial list changes (e.g. after a query refetch).
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

    return () => es.close()
  }, [workspaceId])

  return snippets
}
