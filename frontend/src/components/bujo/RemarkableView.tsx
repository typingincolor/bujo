import { useState } from 'react'
import { Folder, NotebookPen, FileText, BookOpen, File, ChevronRight, RefreshCw } from 'lucide-react'
import { ImportRemarkablePages } from '../../wailsjs/go/wails/App'
import { remarkable, wails } from '../../wailsjs/go/models'
import { OCRReviewPanel } from './OCRReviewPanel'

type Step = 'document-list' | 'importing' | 'review'

interface RemarkableViewProps {
  onNavigateToSettings: () => void
  documents: remarkable.Document[]
  isRegistered: boolean | null
  isLoading: boolean
  onRefresh: () => void
}

function isFolder(doc: remarkable.Document): boolean {
  return doc.FileType === ''
}

function isImportable(doc: remarkable.Document): boolean {
  return doc.FileType === 'notebook' || doc.FileType === 'pdf'
}

function fileIcon(doc: remarkable.Document) {
  if (isFolder(doc)) return <Folder className="w-4 h-4 text-blue-400" />
  switch (doc.FileType) {
    case 'notebook': return <NotebookPen className="w-4 h-4 text-amber-400" />
    case 'pdf': return <FileText className="w-4 h-4 text-red-400" />
    case 'epub': return <BookOpen className="w-4 h-4 text-green-400" />
    default: return <File className="w-4 h-4 text-muted-foreground" />
  }
}

function formatDate(timestamp: string): string {
  const ms = parseInt(timestamp, 10)
  if (isNaN(ms)) return timestamp
  return new Date(ms).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

export function RemarkableView({ onNavigateToSettings, documents, isRegistered, isLoading, onRefresh }: RemarkableViewProps) {
  const [step, setStep] = useState<Step>('document-list')
  const [error, setError] = useState<string | null>(null)
  const [importResult, setImportResult] = useState<wails.ImportRemarkableResult | null>(null)
  const [selectedDocName, setSelectedDocName] = useState('')
  const [currentFolderId, setCurrentFolderId] = useState('')
  const [breadcrumbs, setBreadcrumbs] = useState<{ id: string; name: string }[]>([])

  function handleNavigateToFolder(folderId: string, folderName: string) {
    setCurrentFolderId(folderId)
    setBreadcrumbs(prev => [...prev, { id: folderId, name: folderName }])
  }

  function handleBreadcrumbClick(index: number) {
    if (index === -1) {
      setCurrentFolderId('')
      setBreadcrumbs([])
    } else {
      const target = breadcrumbs[index]
      setCurrentFolderId(target.id)
      setBreadcrumbs(prev => prev.slice(0, index + 1))
    }
  }

  function currentDocuments(): remarkable.Document[] {
    const filtered = documents.filter(doc => doc.Parent === currentFolderId)
    const folders = filtered.filter(isFolder).sort((a, b) => a.VisibleName.localeCompare(b.VisibleName))
    const files = filtered.filter(d => !isFolder(d)).sort((a, b) => a.VisibleName.localeCompare(b.VisibleName))
    return [...folders, ...files]
  }

  async function handleSelectDocument(doc: remarkable.Document) {
    if (isFolder(doc)) {
      handleNavigateToFolder(doc.ID, doc.VisibleName)
      return
    }
    if (!isImportable(doc)) {
      return
    }
    setSelectedDocName(doc.VisibleName)
    setStep('importing')
    setError(null)

    try {
      const result = await ImportRemarkablePages(doc.ID)
      setImportResult(result)
      setStep('review')
    } catch (err) {
      setError(String(err))
      setStep('document-list')
    }
  }

  if (isRegistered === null) {
    return <div className="p-6 text-muted-foreground">Loading...</div>
  }

  if (!isRegistered) {
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
          onClick={() => { setError(null); onRefresh() }}
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
        pages={importResult.pages}
        documentName={selectedDocName}
        onDone={() => {
          setStep('document-list')
          setImportResult(null)
          onRefresh()
        }}
        onBack={() => {
          setStep('document-list')
          setImportResult(null)
        }}
      />
    )
  }

  const docs = currentDocuments()

  return (
    <div className="flex flex-col h-full min-h-0">
      <div className="flex items-center gap-1 px-4 py-2 text-sm text-muted-foreground border-b border-border flex-shrink-0">
        <button
          onClick={() => handleBreadcrumbClick(-1)}
          className="hover:text-foreground transition-colors"
        >
          My files
        </button>
        {breadcrumbs.map((crumb, i) => (
          <span key={crumb.id} className="flex items-center gap-1">
            <ChevronRight className="w-3 h-3" />
            <button
              onClick={() => handleBreadcrumbClick(i)}
              className="hover:text-foreground transition-colors"
            >
              {crumb.name}
            </button>
          </span>
        ))}
        <div className="flex-1" />
        <button
          onClick={onRefresh}
          disabled={isLoading}
          className="p-1 hover:text-foreground transition-colors disabled:opacity-50"
          title="Refresh"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
        </button>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto">
        {docs.length === 0 ? (
          <p className="p-6 text-muted-foreground">
            {isLoading ? 'Loading...' : 'This folder is empty.'}
          </p>
        ) : (
          <div className="divide-y divide-border">
            {docs.map(doc => {
              const clickable = isFolder(doc) || isImportable(doc)
              return (
                <button
                  key={doc.ID}
                  onClick={() => handleSelectDocument(doc)}
                  disabled={!clickable}
                  className={`w-full text-left px-4 py-2 flex items-center gap-3 transition-colors ${
                    clickable
                      ? 'hover:bg-accent cursor-pointer'
                      : 'opacity-50 cursor-default'
                  }`}
                >
                  {fileIcon(doc)}
                  <span className={`flex-1 truncate text-sm ${clickable ? '' : 'text-muted-foreground'}`}>
                    {doc.VisibleName}
                  </span>
                  <span className="text-xs text-muted-foreground flex-shrink-0">
                    {formatDate(doc.LastModified)}
                  </span>
                </button>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
