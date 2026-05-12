import type { Snippet } from '@/lib/api'
import { useWorkspaceBoard } from '@/hooks/useWorkspaceBoard'
import CollaborativeSnippetCard from '@/components/CollaborativeSnippetCard'
import { useAuthStore } from '@/store/authStore'
import { Radio } from 'lucide-react'

interface Props {
  workspaceId: number
  initialSnippets: Snippet[]
}

export default function LiveBoard({ workspaceId, initialSnippets }: Props) {
  const { snippets, editingBy } = useWorkspaceBoard(workspaceId, initialSnippets)
  const user = useAuthStore((s) => s.user)

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Radio className="h-3.5 w-3.5 text-green-500 animate-pulse" />
        Live — edits save automatically and sync to all members
      </div>

      {snippets.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <p className="text-muted-foreground">No snippets yet. Add one to see it here!</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {snippets.map((s) => (
            <CollaborativeSnippetCard
              key={s.id}
              snippet={s}
              workspaceId={workspaceId}
              currentUserId={user?.id ?? 0}
              editingBy={editingBy}
            />
          ))}
        </div>
      )}
    </div>
  )
}
