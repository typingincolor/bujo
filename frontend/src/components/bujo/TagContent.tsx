const TAG_REGEX = /#([a-zA-Z][a-zA-Z0-9-]*)/g
const MENTION_REGEX = /@([a-zA-Z][a-zA-Z0-9.]*)/g
const URL_REGEX = /https?:\/\/[^\s)\]]+/g

type Part = string | { tag: string } | { mention: string } | { url: string }

interface TagContentProps {
  content: string
  onTagClick?: (tag: string) => void
  onMentionClick?: (mention: string) => void
}

export function TagContent({ content, onTagClick, onMentionClick }: TagContentProps) {
  const parts: Part[] = []
  const matches: { index: number; length: number; part: { tag: string } | { mention: string } | { url: string } }[] = []

  for (const match of content.matchAll(TAG_REGEX)) {
    matches.push({ index: match.index, length: match[0].length, part: { tag: match[1] } })
  }
  for (const match of content.matchAll(MENTION_REGEX)) {
    matches.push({ index: match.index, length: match[0].length, part: { mention: match[1] } })
  }
  for (const match of content.matchAll(URL_REGEX)) {
    matches.push({ index: match.index, length: match[0].length, part: { url: match[0] } })
  }

  matches.sort((a, b) => a.index - b.index)

  let lastIndex = 0
  for (const m of matches) {
    if (m.index < lastIndex) continue
    const before = content.slice(lastIndex, m.index)
    if (before) parts.push(before)
    parts.push(m.part)
    lastIndex = m.index + m.length
  }

  const after = content.slice(lastIndex)
  if (after) parts.push(after)

  if (parts.length === 0) {
    return <span />
  }

  return (
    <span>
      {parts.map((part, i) => {
        if (typeof part === 'string') {
          return <span key={i}>{part}</span>
        }
        if ('tag' in part) {
          return (
            <span
              key={i}
              className={`tag${onTagClick ? ' cursor-pointer' : ''}`}
              onClick={onTagClick ? () => onTagClick(part.tag) : undefined}
            >
              #{part.tag}
            </span>
          )
        }
        if ('url' in part) {
          return (
            <a
              key={i}
              href={part.url}
              target="_blank"
              rel="noopener noreferrer"
              className="link"
            >
              {part.url}
            </a>
          )
        }
        return (
          <span
            key={i}
            className={`mention${onMentionClick ? ' cursor-pointer' : ''}`}
            onClick={onMentionClick ? () => onMentionClick(part.mention) : undefined}
          >
            @{part.mention}
          </span>
        )
      })}
    </span>
  )
}
