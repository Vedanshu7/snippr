import { useEffect, useRef, useState } from 'react'
import hljs from 'highlight.js'
import { Button } from '@/components/ui/button'
import { Copy, Check } from 'lucide-react'

interface Props {
  content: string
  language: string
}

export default function CodeBlock({ content, language }: Props) {
  const ref = useRef<HTMLElement>(null)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!ref.current) return
    ref.current.removeAttribute('data-highlighted')
    ref.current.className = `language-${language}`
    hljs.highlightElement(ref.current)
  }, [content, language])

  function copy() {
    navigator.clipboard.writeText(content)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="relative group">
      <pre className="rounded-lg overflow-x-auto text-sm leading-relaxed bg-muted p-4">
        <code ref={ref} className={`language-${language}`}>
          {content}
        </code>
      </pre>
      <Button
        variant="ghost"
        size="sm"
        className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
        onClick={copy}
      >
        {copied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
      </Button>
    </div>
  )
}
