import { autocompletion, CompletionContext, CompletionResult } from '@codemirror/autocomplete'
import { Extension } from '@codemirror/state'

export function tagCompletionSource(tags: string[]) {
  return (context: CompletionContext): CompletionResult | null => {
    const word = context.matchBefore(/#[a-zA-Z][a-zA-Z0-9-]*/)
    if (!word) return null

    const partial = word.text.slice(1)
    const filtered = tags.filter(t => t.toLowerCase().startsWith(partial.toLowerCase()))

    if (filtered.length === 0) return null

    return {
      from: word.from,
      options: filtered.map(tag => ({
        label: `#${tag}`,
        type: 'keyword',
      })),
    }
  }
}

export function tagAutocomplete(tags: string[]): Extension {
  return autocompletion({
    override: [tagCompletionSource(tags)],
  })
}
