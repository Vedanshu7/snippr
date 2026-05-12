import { Link } from 'react-router-dom'
import type { Snippet } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Globe, Lock } from 'lucide-react'
import { formatDistanceToNow } from '@/lib/time'

interface Props {
  snippet: Snippet
}

export default function SnippetCard({ snippet }: Props) {
  return (
    <Link to={`/snippets/${snippet.id}`}>
      <Card className="hover:border-primary/50 transition-colors cursor-pointer h-full">
        <CardHeader className="pb-2">
          <div className="flex items-start justify-between gap-2">
            <CardTitle className="text-base leading-tight">{snippet.title}</CardTitle>
            {snippet.is_public ? (
              <Globe className="h-4 w-4 text-muted-foreground shrink-0 mt-0.5" />
            ) : (
              <Lock className="h-4 w-4 text-muted-foreground shrink-0 mt-0.5" />
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-3">
          <pre className="text-xs text-muted-foreground bg-muted rounded p-2 overflow-hidden max-h-20 font-mono leading-relaxed">
            {snippet.content.slice(0, 200)}
          </pre>
          <div className="flex items-center justify-between gap-2">
            <div className="flex flex-wrap gap-1">
              <Badge variant="secondary" className="text-xs">{snippet.language}</Badge>
              {snippet.tags.slice(0, 3).map((t) => (
                <Badge key={t} variant="outline" className="text-xs">{t}</Badge>
              ))}
            </div>
            <span className="text-xs text-muted-foreground whitespace-nowrap">
              {formatDistanceToNow(snippet.updated_at)}
            </span>
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}
