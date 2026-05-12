import { Link, useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { snippets as api } from '@/lib/api'
import Navbar from '@/components/Navbar'
import CodeBlock from '@/components/CodeBlock'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Globe, Pencil, Trash2, Link2 } from 'lucide-react'
import { formatDistanceToNow } from '@/lib/time'

export default function SnippetDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const qc = useQueryClient()

  const { data: snippet, isLoading } = useQuery({
    queryKey: ['snippet', id],
    queryFn: () => api.get(Number(id)).then((r) => r.data),
    enabled: Boolean(id),
  })

  const deleteMutation = useMutation({
    mutationFn: () => api.delete(Number(id)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['snippets'] })
      navigate('/dashboard')
    },
  })

  function copyShareLink() {
    if (!snippet?.share_slug) return
    const url = `${window.location.origin}/s/${snippet.share_slug}`
    navigator.clipboard.writeText(url)
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar />
        <main className="max-w-3xl mx-auto px-4 py-8">
          <div className="h-96 rounded-lg bg-muted animate-pulse" />
        </main>
      </div>
    )
  }

  if (!snippet) {
    return (
      <div className="min-h-screen bg-background">
        <Navbar />
        <main className="max-w-3xl mx-auto px-4 py-8 text-center">
          <p className="text-muted-foreground">Snippet not found.</p>
          <Button asChild className="mt-4" variant="outline">
            <Link to="/dashboard">Back to dashboard</Link>
          </Button>
        </main>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="max-w-3xl mx-auto px-4 py-8 space-y-6">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1 min-w-0">
            <h1 className="text-2xl font-semibold truncate">{snippet.title}</h1>
            <div className="flex items-center gap-2 flex-wrap">
              <Badge variant="secondary">{snippet.language}</Badge>
              {snippet.tags.map((t) => (
                <Badge key={t} variant="outline">{t}</Badge>
              ))}
              {snippet.is_public && (
                <span className="flex items-center gap-1 text-xs text-muted-foreground">
                  <Globe className="h-3 w-3" /> Public
                </span>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2 shrink-0">
            {snippet.is_public && snippet.share_slug && (
              <Button variant="outline" size="sm" onClick={copyShareLink}>
                <Link2 className="h-4 w-4" />
              </Button>
            )}
            <Button variant="outline" size="sm" asChild>
              <Link to={`/snippets/${snippet.id}/edit`}>
                <Pencil className="h-4 w-4" />
              </Link>
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                if (confirm('Delete this snippet?')) deleteMutation.mutate()
              }}
              disabled={deleteMutation.isPending}
            >
              <Trash2 className="h-4 w-4 text-destructive" />
            </Button>
          </div>
        </div>

        <Separator />

        <CodeBlock content={snippet.content} language={snippet.language} />

        <p className="text-xs text-muted-foreground">
          Updated {formatDistanceToNow(snippet.updated_at)}
        </p>
      </main>
    </div>
  )
}
