package domain

import (
	"errors"
	"strings"
)

var symbolToType = map[rune]EntryType{
	// ASCII input symbols
	'.': EntryTypeTask,
	'-': EntryTypeNote,
	'o': EntryTypeEvent,
	'x': EntryTypeDone,
	'>': EntryTypeMigrated,
	// Unicode display symbols (also accepted as input)
	'•': EntryTypeTask,
	'–': EntryTypeNote,
	'○': EntryTypeEvent,
	'✓': EntryTypeDone,
	'→': EntryTypeMigrated,
}

func ParseEntryType(line string) EntryType {
	if len(line) == 0 {
		return ""
	}
	for _, r := range line {
		return symbolToType[r]
	}
	return ""
}

func ParseIndentation(line string) (depth int, rest string) {
	for i, ch := range line {
		switch ch {
		case ' ':
			depth++
		case '\t':
			depth += 2
		default:
			return depth / 2, line[i:]
		}
	}
	return depth / 2, ""
}

func ParseContent(line string) string {
	runes := []rune(line)
	if len(runes) < 2 {
		return ""
	}
	return strings.TrimSpace(string(runes[1:]))
}

type TreeParser struct{}

func NewTreeParser() *TreeParser {
	return &TreeParser{}
}

func (p *TreeParser) Parse(input string) ([]Entry, error) {
	if input == "" {
		return []Entry{}, nil
	}

	lines := strings.Split(input, "\n")
	entries := make([]Entry, 0, len(lines))
	parentStack := make([]int, 0)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		depth, rest := ParseIndentation(line)
		entryType := ParseEntryType(rest)

		if !entryType.IsValid() {
			return nil, errors.New("unknown entry type symbol")
		}

		content := ParseContent(rest)

		entry := Entry{
			Type:    entryType,
			Content: content,
			Depth:   depth,
		}

		if depth == 0 {
			parentStack = []int{len(entries)}
		} else {
			if depth > len(parentStack) {
				return nil, errors.New("invalid indentation: child without parent at correct depth")
			}

			parentStack = parentStack[:depth]
			parentIdx := int64(parentStack[len(parentStack)-1])
			entry.ParentID = &parentIdx
			parentStack = append(parentStack, len(entries))
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
