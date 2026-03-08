import { useState, useEffect } from 'react'
import { ListRemarkableDocuments, IsRemarkableRegistered, ImportRemarkablePages } from '../../wailsjs/go/wails/App'
import { remarkable, wails } from '../../wailsjs/go/models'
import { OCRReviewPanel } from './OCRReviewPanel'

type Step = 'loading' | 'not-registered' | 'document-list' | 'importing' | 'review'

interface RemarkableViewProps {
  onNavigateToSettings: () => void
}

export function RemarkableView({ onNavigateToSettings }: RemarkableViewProps) {
  const [step, setStep] = useState<Step>('loading')
  const [documents, setDocuments] = useState<remarkable.Document[]>([])
  const [error, setError] = useState<string | null>(null)
  const [importResult, setImportResult] = useState<wails.ImportRemarkableResult | null>(null)
  const [selectedDocName, setSelectedDocName] = useState('')

  useEffect(() => {
    IsRemarkableRegistered().then(registered => {
      if (!registered) {
        setStep('not-registered')
        return
      }
      loadDocuments()
    })
  }, [])

  async function loadDocuments() {
    try {
      setError(null)
      const docs = await ListRemarkableDocuments()
      setDocuments(docs)
      setStep('document-list')
    } catch (err) {
      setError(String(err))
      setStep('document-list')
    }
  }

  async function handleSelectDocument(docID: string) {
    const doc = documents.find(d => d.ID === docID)
    setSelectedDocName(doc?.VisibleName ?? docID)
    setStep('importing')
    setError(null)

    try {
      const result = await ImportRemarkablePages(docID)
      setImportResult(result)
      setStep('review')
    } catch (err) {
      setError(String(err))
      setStep('document-list')
    }
  }

  if (step === 'loading') {
    return <div className="p-6 text-muted-foreground">Loading...</div>
  }

  if (step === 'not-registered') {
    return (
      <div className="p-6 space-y-4">
        <p className="text-muted-foreground">
          reMarkable tablet not connected. Register your device in Settings to get started.
        </p>
        <button
          onClick={onNavigateToSettings}
          className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm"
        >
          Open Settings
        </button>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-6 space-y-4">
        <p className="text-destructive">{error}</p>
        <button
          onClick={loadDocuments}
          className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm"
        >
          Retry
        </button>
      </div>
    )
  }

  if (step === 'importing') {
    return (
      <div className="p-6 space-y-2">
        <p className="text-muted-foreground">
          Downloading and processing pages from &quot;{selectedDocName}&quot;...
        </p>
        <p className="text-xs text-muted-foreground">This may take a moment for notebooks with many pages.</p>
      </div>
    )
  }

  if (step === 'review' && importResult) {
    return (
      <OCRReviewPanel
        pages={importResult.Pages}
        documentName={selectedDocName}
        onDone={() => {
          setStep('document-list')
          setImportResult(null)
          loadDocuments()
        }}
        onBack={() => {
          setStep('document-list')
          setImportResult(null)
        }}
      />
    )
  }

  return (
    <div className="p-6 space-y-2">
      {documents.length === 0 ? (
        <p className="text-muted-foreground">No notebooks found on your reMarkable.</p>
      ) : (
        documents.map(doc => (
          <button
            key={doc.ID}
            onClick={() => handleSelectDocument(doc.ID)}
            className="w-full text-left px-4 py-3 rounded-lg border border-border hover:bg-accent transition-colors"
          >
            <div className="font-medium">{doc.VisibleName}</div>
            <div className="text-xs text-muted-foreground">{doc.FileType} · {doc.LastModified}</div>
          </button>
        ))
      )}
    </div>
  )
}
