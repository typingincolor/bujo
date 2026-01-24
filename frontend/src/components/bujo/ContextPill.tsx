interface ContextPillProps {
  count: number
  onClick?: () => void
  isLoading?: boolean
}

export function ContextPill({ count, onClick, isLoading = false }: ContextPillProps) {
  const title = isLoading
    ? 'Loading parent context'
    : `Show ${count} parent${count > 1 ? 's' : ''}`

  const displayText = isLoading ? '⋯' : `${count} above`

  return (
    <button
      data-testid="context-pill"
      onClick={onClick ? (e) => {
        e.stopPropagation()
        onClick()
      } : undefined}
      title={title}
      aria-label={title}
      className="px-1.5 py-0.5 text-xs font-medium bg-muted text-muted-foreground rounded-full hover:bg-secondary transition-colors flex-shrink-0"
    >
      {displayText}
    </button>
  )
}
