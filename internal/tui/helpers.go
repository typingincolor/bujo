package tui

import (
	"strings"

	"github.com/typingincolor/bujo/internal/domain"
)

func detectLineIndentation(line string) string {
	indent := ""
	for _, ch := range line {
		if ch == ' ' {
			indent += " "
		} else {
			break
		}
	}
	return indent
}

func searchForward(content, query string, startPos int) int {
	if query == "" || len(content) == 0 {
		return -1
	}

	searchStart := startPos
	if searchStart < 0 {
		searchStart = 0
	}
	if searchStart >= len(content) {
		searchStart = 0
	}

	idx := strings.Index(content[searchStart:], query)
	if idx >= 0 {
		return searchStart + idx
	}

	if startPos > 0 {
		idx = strings.Index(content[:searchStart], query)
		if idx >= 0 {
			return idx
		}
	}

	return -1
}

func searchBackward(content, query string, endPos int) int {
	if query == "" || len(content) == 0 {
		return -1
	}

	searchEnd := endPos
	if searchEnd < 0 {
		searchEnd = len(content)
	}
	if searchEnd > len(content) {
		searchEnd = len(content)
	}

	idx := strings.LastIndex(content[:searchEnd], query)
	if idx >= 0 {
		return idx
	}

	if endPos < len(content) {
		idx = strings.LastIndex(content[searchEnd:], query)
		if idx >= 0 {
			return searchEnd + idx
		}
	}

	return -1
}

func countEntryDescendants(entryID int64, parentMap map[int64][]domain.Entry) int {
	children := parentMap[entryID]
	count := len(children)
	for _, child := range children {
		count += countEntryDescendants(child.ID, parentMap)
	}
	return count
}

func insertCursorInLine(line string, cursorCol int) string {
	if cursorCol < 0 {
		return line
	}
	if cursorCol < len(line) {
		return line[:cursorCol] + "█" + line[cursorCol+1:]
	}
	return line + "█"
}

func highlightSearchMatches(line, query string, cursorCol int) string {
	if query == "" || !strings.Contains(line, query) {
		return insertCursorInLine(line, cursorCol)
	}

	var result strings.Builder
	pos := 0
	remaining := line
	cursorInserted := cursorCol < 0

	for {
		idx := strings.Index(remaining, query)
		if idx < 0 {
			if !cursorInserted && cursorCol >= pos {
				relCol := cursorCol - pos
				if relCol < len(remaining) {
					result.WriteString(remaining[:relCol])
					result.WriteString("█")
					result.WriteString(remaining[relCol+1:])
				} else {
					result.WriteString(remaining)
					result.WriteString("█")
				}
			} else {
				result.WriteString(remaining)
			}
			break
		}

		matchStart := pos + idx
		matchEnd := matchStart + len(query)

		if !cursorInserted && cursorCol >= pos && cursorCol < matchStart {
			relCol := cursorCol - pos
			result.WriteString(remaining[:relCol])
			result.WriteString("█")
			result.WriteString(remaining[relCol+1 : idx])
			cursorInserted = true
		} else {
			result.WriteString(remaining[:idx])
		}

		if !cursorInserted && cursorCol >= matchStart && cursorCol < matchEnd {
			relCol := cursorCol - matchStart
			matchText := query[:relCol] + "█" + query[relCol+1:]
			result.WriteString(SearchHighlightStyle.Render(matchText))
			cursorInserted = true
		} else {
			result.WriteString(SearchHighlightStyle.Render(query))
		}

		pos = matchEnd
		remaining = remaining[idx+len(query):]
	}

	return result.String()
}
