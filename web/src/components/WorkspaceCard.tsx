import { Link } from 'react-router-dom'
import type { Workspace } from '@/lib/api'
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Globe, Lock, Users } from 'lucide-react'

interface Props {
  workspace: Workspace
}

export default function WorkspaceCard({ workspace }: Props) {
  const created = new Date(workspace.created_at).toLocaleDateString()

  return (
    <Link to={`/workspaces/${workspace.id}`}>
      <Card className="hover:border-primary/50 transition-colors cursor-pointer">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Users className="h-4 w-4 text-muted-foreground" />
              <CardTitle className="text-base">{workspace.name}</CardTitle>
            </div>
            <Badge variant="secondary" className="gap-1 text-xs">
              {workspace.is_public
                ? <><Globe className="h-3 w-3" />public</>
                : <><Lock className="h-3 w-3" />private</>}
            </Badge>
          </div>
          <CardDescription>Created {created}</CardDescription>
        </CardHeader>
      </Card>
    </Link>
  )
}
