import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { Snippet, CreateSnippetReq } from '@/lib/api'
import { snippets as api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { X } from 'lucide-react'

const LANGUAGES = [
  'text', 'go', 'javascript', 'typescript', 'python', 'rust',
  'java', 'c', 'cpp', 'bash', 'sql', 'html', 'css', 'json', 'yaml', 'markdown',
]

interface Props {
  initial?: Snippet
}

export default function SnippetForm({ initial }: Props) {
  const navigate = useNavigate()
  const [title, setTitle] = useState(initial?.title ?? '')
  const [content, setContent] = useState(initial?.content ?? '')
  const [language, setLanguage] = useState(initial?.language ?? 'text')
  const [isPublic, setIsPublic] = useState(initial?.is_public ?? false)
  const [tags, setTags] = useState<string[]>(initial?.tags ?? [])
  const [tagInput, setTagInput] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  function addTag(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      const tag = tagInput.trim().toLowerCase()
      if (tag && !tags.includes(tag)) setTags([...tags, tag])
      setTagInput('')
    }
  }

  function removeTag(tag: string) {
    setTags(tags.filter((t) => t !== tag))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!title.trim()) { setError('Title is required'); return }
    setError('')
    setLoading(true)
    const data: CreateSnippetReq = { title, content, language, is_public: isPublic, tags }
    try {
      if (initial) {
        await api.update(initial.id, data)
        navigate(`/snippets/${initial.id}`)
      } else {
        const res = await api.create(data)
        navigate(`/snippets/${res.data.id}`)
      }
    } catch {
      setError('Failed to save snippet')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      <div className="space-y-2">
        <Label htmlFor="title">Title</Label>
        <Input
          id="title"
          placeholder="My awesome snippet"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="language">Language</Label>
        <Select value={language} onValueChange={setLanguage}>
          <SelectTrigger id="language">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {LANGUAGES.map((l) => (
              <SelectItem key={l} value={l}>{l}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="content">Content</Label>
        <Textarea
          id="content"
          placeholder="Paste your snippet here…"
          className="font-mono text-sm min-h-64 resize-y"
          value={content}
          onChange={(e) => setContent(e.target.value)}
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="tags">Tags</Label>
        <div className="flex flex-wrap gap-1 mb-1">
          {tags.map((t) => (
            <Badge key={t} variant="secondary" className="gap-1">
              {t}
              <button type="button" onClick={() => removeTag(t)}>
                <X className="h-3 w-3" />
              </button>
            </Badge>
          ))}
        </div>
        <Input
          id="tags"
          placeholder="Add tag, press Enter…"
          value={tagInput}
          onChange={(e) => setTagInput(e.target.value)}
          onKeyDown={addTag}
        />
      </div>

      <div className="flex items-center gap-2">
        <input
          id="public"
          type="checkbox"
          className="h-4 w-4 rounded border-input"
          checked={isPublic}
          onChange={(e) => setIsPublic(e.target.checked)}
        />
        <Label htmlFor="public" className="font-normal">Make public (sharable link)</Label>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="flex gap-2 justify-end">
        <Button type="button" variant="outline" onClick={() => navigate(-1)}>
          Cancel
        </Button>
        <Button type="submit" disabled={loading}>
          {loading ? 'Saving…' : initial ? 'Update' : 'Create'}
        </Button>
      </div>
    </form>
  )
}
