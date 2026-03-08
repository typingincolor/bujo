import { useState, useMemo } from 'react'
import { ImportEntries } from '../../wailsjs/go/wails/App'
import { remarkable, wails } from '../../wailsjs/go/models'

interface OCRReviewPanelProps {
  pages: wails.ImportedPage[]
  documentName: string
  onDone: () => void
  onBack: () => void
}

function reconstructTextWithConfidence(results: remarkable.OCRResult[], threshold = 0.8): { text: string, lowConfidenceCount: number } {
  if (!results || results.length === 0) return { text: '', lowConfidenceCount: 0 }

  const sorted = [...results].sort((a, b) => a.y - b.y)
  const minX = Math.min(...sorted.map(r => r.x))
  const indentWidth = 50

  let lowConfidenceCount = 0
  const text = sorted.map((r) => {
    const depth = Math.round((r.x - minX) / indentWidth)
    const indent = '  '.repeat(depth)
    if (r.confidence < threshold) lowConfidenceCount++
    return indent + r.text
  }).join('\n')

  return { text, lowConfidenceCount }
}

export function OCRReviewPanel({ pages, documentName, onDone, onBack }: OCRReviewPanelProps) {
  const [currentPage, setCurrentPage] = useState(0)
  const [date, setDate] = useState(() => new Date().toISOString().split('T')[0])
  const [importing, setImporting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const pageData = useMemo(() => {
    return pages.map(p => reconstructTextWithConfidence(p.ocrResults))
  }, [pages])

  const [editedTexts, setEditedTexts] = useState<string[]>(() => pageData.map(d => d.text))

  const page = pages[currentPage]
  const hasError = page?.error
  const { lowConfidenceCount } = pageData[currentPage] ?? { lowConfidenceCount: 0 }

  function updateText(index: number, text: string) {
    setEditedTexts(prev => {
      const next = [...prev]
      next[index] = text
      return next
    })
  }

  async function handleImport() {
    setImporting(true)
    setError(null)

    const combined = editedTexts.filter(t => t.trim()).join('\n')
    if (!combined.trim()) {
      setError('No text to import')
      setImporting(false)
      return
    }

    try {
      await ImportEntries(combined, date)
      onDone()
    } catch (err) {
      setError(String(err))
      setImporting(false)
    }
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-border">
        <div className="flex items-center gap-4">
          <button onClick={onBack} className="text-sm text-muted-foreground hover:text-foreground">
            &larr; Back
          </button>
          <span className="font-medium">{documentName}</span>
          <span className="text-sm text-muted-foreground">
            Page {currentPage + 1} of {pages.length}
          </span>
        </div>
        <div className="flex items-center gap-3">
          <label className="text-sm text-muted-foreground">
            Date:
            <input
              type="date"
              value={date}
              onChange={e => setDate(e.target.value)}
              className="ml-2 px-2 py-1 bg-background border border-border rounded text-sm"
            />
          </label>
          <button
            onClick={handleImport}
            disabled={importing}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm disabled:opacity-50"
          >
            {importing ? 'Importing...' : 'Import to Journal'}
          </button>
        </div>
      </div>

      {error && (
        <div className="px-4 py-2 bg-destructive/10 text-destructive text-sm">{error}</div>
      )}

      {/* Page navigation */}
      {pages.length > 1 && (
        <div className="flex gap-1 p-2 border-b border-border overflow-x-auto">
          {pages.map((_, i) => (
            <button
              key={i}
              onClick={() => setCurrentPage(i)}
              className={`px-3 py-1 rounded text-sm ${
                i === currentPage
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-accent'
              }`}
            >
              {i + 1}
            </button>
          ))}
        </div>
      )}

      {/* Side-by-side content */}
      <div className="flex-1 flex min-h-0">
        {/* Left: PNG preview */}
        <div className="w-1/2 overflow-auto border-r border-border p-4">
          {hasError ? (
            <div className="text-destructive text-sm">{page.error}</div>
          ) : page?.png ? (
            <img
              src={`data:image/png;base64,${page.png}`}
              alt={`Page ${currentPage + 1}`}
              className="max-w-full"
            />
          ) : (
            <div className="text-muted-foreground text-sm">No image available</div>
          )}
        </div>

        {/* Right: Text editor */}
        <div className="w-1/2 overflow-auto p-4">
          <textarea
            value={editedTexts[currentPage] ?? ''}
            onChange={e => updateText(currentPage, e.target.value)}
            className="w-full h-full min-h-[400px] p-3 bg-background border border-border rounded font-mono text-sm resize-none focus:outline-none focus:ring-1 focus:ring-primary"
            placeholder="OCR text will appear here..."
          />
          {lowConfidenceCount > 0 && (
            <p className="text-xs text-amber-500 mt-2">
              {lowConfidenceCount} low-confidence region(s) detected — review text carefully
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
