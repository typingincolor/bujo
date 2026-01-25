import { useState, useRef } from 'react'
import { Upload, X } from 'lucide-react'
import * as Popover from '@radix-ui/react-popover'
import { cn } from '@/lib/utils'

interface FileUploadButtonProps {
  files?: File[]
  onFilesChange?: (files: File[]) => void
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function FileUploadButton({
  files = [],
  onFilesChange,
}: FileUploadButtonProps) {
  const [isDragOver, setIsDragOver] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = event.target.files
    if (selectedFiles && selectedFiles.length > 0) {
      onFilesChange?.([...files, ...Array.from(selectedFiles)])
    }
  }

  const handleRemoveFile = (index: number) => {
    const newFiles = files.filter((_, i) => i !== index)
    onFilesChange?.(newFiles)
  }

  const handleDragOver = (event: React.DragEvent) => {
    event.preventDefault()
    setIsDragOver(true)
  }

  const handleDragLeave = (event: React.DragEvent) => {
    event.preventDefault()
    setIsDragOver(false)
  }

  const handleDrop = (event: React.DragEvent) => {
    event.preventDefault()
    setIsDragOver(false)
    const droppedFiles = event.dataTransfer.files
    if (droppedFiles && droppedFiles.length > 0) {
      onFilesChange?.([...files, ...Array.from(droppedFiles)])
    }
  }

  return (
    <Popover.Root>
      <Popover.Trigger asChild>
        <button
          className="relative flex items-center justify-center h-8 w-8 rounded-lg border border-border hover:bg-secondary/50 transition-colors"
          aria-label="Upload files"
        >
          <Upload className="h-4 w-4" />
          {files.length > 0 && (
            <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-primary text-[10px] text-primary-foreground">
              {files.length}
            </span>
          )}
        </button>
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          className="z-50 w-80 bg-card border border-border rounded-lg shadow-lg p-4"
          align="end"
          sideOffset={5}
        >
          <div className="space-y-4">
            <h4 className="font-medium">Upload Files</h4>

            <div
              data-testid="drop-zone"
              className={cn(
                'flex flex-col items-center justify-center rounded-lg border-2 border-dashed p-6 transition-colors',
                isDragOver
                  ? 'border-primary bg-primary/10'
                  : 'border-muted-foreground/25'
              )}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
            >
              <Upload className="mb-2 h-8 w-8 text-muted-foreground" />
              <p className="text-center text-sm text-muted-foreground">
                Drag and drop files here, or{' '}
                <label className="cursor-pointer text-primary underline">
                  choose files
                  <input
                    ref={inputRef}
                    type="file"
                    multiple
                    className="sr-only"
                    onChange={handleFileChange}
                    aria-label="Choose files"
                  />
                </label>
              </p>
            </div>

            {files.length > 0 && (
              <ul className="space-y-2">
                {files.map((file, index) => (
                  <li
                    key={`${file.name}-${index}`}
                    className="flex items-center justify-between rounded bg-muted px-3 py-2 text-sm"
                  >
                    <div className="flex-1 truncate">
                      <span className="font-medium">{file.name}</span>
                      <span className="ml-2 text-muted-foreground">
                        {formatFileSize(file.size)}
                      </span>
                    </div>
                    <button
                      className="h-6 w-6 flex items-center justify-center rounded hover:bg-secondary/50"
                      onClick={() => handleRemoveFile(index)}
                      aria-label={`Remove ${file.name}`}
                    >
                      <X className="h-4 w-4" />
                    </button>
                  </li>
                ))}
              </ul>
            )}

            <p className="text-xs text-muted-foreground">
              Storage not connected. Files will be stored locally.
            </p>
          </div>
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  )
}
