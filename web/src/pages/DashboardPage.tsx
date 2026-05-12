import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { snippets } from '@/lib/api'
import Navbar from '@/components/Navbar'
import SnippetCard from '@/components/SnippetCard'
import { Input } from '@/components/ui/input'
import { Search } from 'lucide-react'

export default function DashboardPage() {
  const [query, setQuery] = useState('')
  const [debouncedQ, setDebouncedQ] = useState('')

  function handleSearch(e: React.ChangeEvent<HTMLInputElement>) {
    const val = e.target.value
    setQuery(val)
    clearTimeout((window as any)._searchTimer)
    ;(window as any)._searchTimer = setTimeout(() => setDebouncedQ(val), 300)
  }

  const { data, isLoading } = useQuery({
    queryKey: ['snippets', debouncedQ],
    queryFn: () => snippets.list(debouncedQ || undefined).then((r) => r.data),
  })

  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="max-w-5xl mx-auto px-4 py-8 space-y-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search snippets…"
            className="pl-9"
            value={query}
            onChange={handleSearch}
          />
        </div>

        {isLoading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="h-48 rounded-lg bg-muted animate-pulse" />
            ))}
          </div>
        ) : data && data.length > 0 ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {data.map((s) => (
              <SnippetCard key={s.id} snippet={s} />
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <p className="text-muted-foreground">
              {debouncedQ ? 'No snippets matched your search.' : 'No snippets yet. Create your first one!'}
            </p>
          </div>
        )}
      </main>
    </div>
  )
}
