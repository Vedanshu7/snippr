import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { snippets as api } from '@/lib/api'
import Navbar from '@/components/Navbar'
import SnippetForm from '@/components/SnippetForm'

export default function EditSnippetPage() {
  const { id } = useParams<{ id: string }>()

  const { data: snippet, isLoading } = useQuery({
    queryKey: ['snippet', id],
    queryFn: () => api.get(Number(id)).then((r) => r.data),
    enabled: Boolean(id),
  })

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="max-w-2xl mx-auto px-4 py-8">
        <h1 className="text-2xl font-semibold mb-6">Edit snippet</h1>
        {isLoading ? (
          <div className="h-64 rounded-lg bg-muted animate-pulse" />
        ) : snippet ? (
          <SnippetForm initial={snippet} />
        ) : (
          <p className="text-muted-foreground">Snippet not found.</p>
        )}
      </main>
    </div>
  )
}
