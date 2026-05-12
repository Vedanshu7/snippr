import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { snippets as api } from '@/lib/api'
import CodeBlock from '@/components/CodeBlock'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Code2 } from 'lucide-react'
import { formatDistanceToNow } from '@/lib/time'

export default function PublicSnippetPage() {
  const { slug } = useParams<{ slug: string }>()

  const { data: snippet, isLoading } = useQuery({
    queryKey: ['public', slug],
    queryFn: () => api.getPublic(slug!).then((r) => r.data),
    enabled: Boolean(slug),
  })

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="max-w-3xl mx-auto px-4 h-14 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 font-semibold">
            <Code2 className="h-5 w-5" />
            Snippr
          </Link>
          <Button variant="outline" size="sm" asChild>
            <Link to="/login">Sign in</Link>
          </Button>
        </div>
      </header>

      <main className="max-w-3xl mx-auto px-4 py-8">
        {isLoading ? (
          <div className="h-96 rounded-lg bg-muted animate-pulse" />
        ) : !snippet ? (
          <div className="text-center py-20">
            <p className="text-muted-foreground">Snippet not found or no longer public.</p>
          </div>
        ) : (
          <div className="space-y-6">
            <div className="space-y-2">
              <h1 className="text-2xl font-semibold">{snippet.title}</h1>
              <div className="flex items-center gap-2 flex-wrap">
                <Badge variant="secondary">{snippet.language}</Badge>
                {snippet.tags.map((t) => (
                  <Badge key={t} variant="outline">{t}</Badge>
                ))}
                <span className="text-xs text-muted-foreground ml-auto">
                  Updated {formatDistanceToNow(snippet.updated_at)}
                </span>
              </div>
            </div>

            <CodeBlock content={snippet.content} language={snippet.language} />

            <p className="text-sm text-muted-foreground text-center">
              Shared via <span className="font-medium">Snippr</span> ·{' '}
              <Link to="/register" className="text-primary underline-offset-4 hover:underline">
                Create your own
              </Link>
            </p>
          </div>
        )}
      </main>
    </div>
  )
}
