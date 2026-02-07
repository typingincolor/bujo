const TAG_REGEX = /#([a-zA-Z][a-zA-Z0-9-]*)/g

interface TagContentProps {
  content: string
  onTagClick?: (tag: string) => void
}

export function TagContent({ content, onTagClick }: TagContentProps) {
  const parts: (string | { tag: string })[] = []
  let lastIndex = 0

  for (const match of content.matchAll(TAG_REGEX)) {
    const before = content.slice(lastIndex, match.index)
    if (before) parts.push(before)
    parts.push({ tag: match[1] })
    lastIndex = match.index + match[0].length
  }

  const after = content.slice(lastIndex)
  if (after) parts.push(after)

  if (parts.length === 0) {
    return <span />
  }

  return (
    <span>
      {parts.map((part, i) =>
        typeof part === 'string' ? (
          <span key={i}>{part}</span>
        ) : (
          <span
            key={i}
            className={`tag${onTagClick ? ' cursor-pointer' : ''}`}
            onClick={onTagClick ? () => onTagClick(part.tag) : undefined}
          >
            #{part.tag}
          </span>
        ),
      )}
    </span>
  )
}
