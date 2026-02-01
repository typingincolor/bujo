package domain

import (
	"strings"
)

var editableSymbolToType = map[rune]EntryType{
	'.': EntryTypeTask,
	'-': EntryTypeNote,
	'o': EntryTypeEvent,
	'x': EntryTypeDone,
	'~': EntryTypeCancelled,
	'?': EntryTypeQuestion,
	'>': EntryTypeMigrated,
	'*': EntryTypeAnswered,
	'^': EntryTypeMovedToList,
}

type EditableDocumentParser struct{}

func NewEditableDocumentParser() *EditableDocumentParser {
	return &EditableDocumentParser{}
}

func (p *EditableDocumentParser) Parse(input string) (*EditableDocument, error) {
	lines := strings.Split(input, "\n")
	doc := &EditableDocument{
		Lines: make([]ParsedLine, 0, len(lines)),
	}

	for i, line := range lines {
		lineNum := i + 1

		if strings.TrimSpace(line) == "" {
			doc.Lines = append(doc.Lines, ParsedLine{
				LineNumber: lineNum,
				Raw:        line,
				IsValid:    false,
			})
			continue
		}

		parsedLine := p.ParseLine(line, lineNum)
		doc.Lines = append(doc.Lines, parsedLine)
	}

	return doc, nil
}

var typeToEditableSymbol = map[EntryType]rune{
	EntryTypeTask:      '.',
	EntryTypeNote:      '-',
	EntryTypeEvent:     'o',
	EntryTypeDone:      'x',
	EntryTypeCancelled: '~',
	EntryTypeQuestion:  '?',
	EntryTypeMigrated:  '>',
	EntryTypeAnswered:  '*',
	EntryTypeAnswer:      '-',
	EntryTypeMovedToList: '^',
}

func Serialize(entries []Entry) string {
	if len(entries) == 0 {
		return ""
	}

	hasParentIDs := false
	for _, e := range entries {
		if e.ParentID != nil || e.ID != 0 {
			hasParentIDs = true
			break
		}
	}

	if !hasParentIDs {
		return serializeFlat(entries)
	}

	childrenOf := make(map[int64][]int)
	for i, e := range entries {
		if e.ParentID != nil {
			childrenOf[*e.ParentID] = append(childrenOf[*e.ParentID], i)
		}
	}

	var result strings.Builder
	written := 0

	var writeEntry func(idx, depth int)
	writeEntry = func(idx, depth int) {
		entry := entries[idx]
		if written > 0 {
			result.WriteString("\n")
		}
		serializeEntryLine(&result, entry, depth)
		written++

		for _, childIdx := range childrenOf[entry.ID] {
			writeEntry(childIdx, depth+1)
		}
	}

	for i, entry := range entries {
		if entry.ParentID == nil {
			writeEntry(i, 0)
		}
	}

	return result.String()
}

func serializeFlat(entries []Entry) string {
	var result strings.Builder
	for i, entry := range entries {
		if i > 0 {
			result.WriteString("\n")
		}
		serializeEntryLine(&result, entry, entry.Depth)
	}

	return result.String()
}

func serializeEntryLine(result *strings.Builder, entry Entry, depth int) {
	for j := 0; j < depth; j++ {
		result.WriteString("  ")
	}

	symbol, ok := typeToEditableSymbol[entry.Type]
	if !ok {
		symbol = '.'
	}
	result.WriteRune(symbol)
	result.WriteString(" ")

	if entry.Priority != PriorityNone && entry.Priority != "" {
		result.WriteString(entry.Priority.Symbol())
		result.WriteString(" ")
	}

	result.WriteString(entry.Content)
}

func (p *EditableDocumentParser) ParseLine(line string, lineNum int) ParsedLine {
	result := ParsedLine{
		LineNumber: lineNum,
		Raw:        line,
		IsValid:    true,
	}

	if strings.HasPrefix(line, "──") || strings.HasPrefix(line, "--") {
		result.IsHeader = true
		return result
	}

	depth, rest := ParseIndentation(line)
	result.Depth = depth

	if len(rest) == 0 {
		result.IsValid = false
		result.ErrorMessage = "Entry content required"
		return result
	}

	firstRune := []rune(rest)[0]
	entryType, ok := editableSymbolToType[firstRune]
	if !ok {
		result.IsValid = false
		result.ErrorMessage = "Unknown entry type"
		return result
	}
	result.Symbol = entryType

	rawContent := ""
	if len([]rune(rest)) > 1 {
		rawContent = strings.TrimSpace(string([]rune(rest)[1:]))
	}

	if rawContent == "" {
		result.IsValid = false
		result.ErrorMessage = "Entry content required"
		return result
	}

	content, priority := ParsePriorityAndContent(rawContent)
	result.Content = content
	result.Priority = priority

	return result
}
