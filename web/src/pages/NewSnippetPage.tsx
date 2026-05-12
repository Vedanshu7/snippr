import Navbar from '@/components/Navbar'
import SnippetForm from '@/components/SnippetForm'

export default function NewSnippetPage() {
  return (
    <div className="min-h-screen bg-background">
      <Navbar />
      <main className="max-w-2xl mx-auto px-4 py-8">
        <h1 className="text-2xl font-semibold mb-6">New snippet</h1>
        <SnippetForm />
      </main>
    </div>
  )
}
