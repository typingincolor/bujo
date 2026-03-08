import { remarkable } from '../../wailsjs/go/models'

export function isFolder(doc: remarkable.Document): boolean {
  return doc.FileType === ''
}

export function isImportable(doc: remarkable.Document): boolean {
  return doc.FileType === 'notebook' || doc.FileType === 'pdf'
}

export function formatDate(timestamp: string): string {
  const ms = parseInt(timestamp, 10)
  if (isNaN(ms)) return timestamp
  return new Date(ms).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

export function sortDocuments(documents: remarkable.Document[], folderId: string): remarkable.Document[] {
  const filtered = documents.filter(doc => doc.Parent === folderId)
  const folders = filtered.filter(isFolder).sort((a, b) => a.VisibleName.localeCompare(b.VisibleName))
  const files = filtered.filter(d => !isFolder(d)).sort((a, b) => a.VisibleName.localeCompare(b.VisibleName))
  return [...folders, ...files]
}
