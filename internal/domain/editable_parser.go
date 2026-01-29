package domain

import (
	"strings"
	"time"
)

var editableSymbolToType = map[rune]EntryType{
	'.': EntryTypeTask,
	'-': EntryTypeNote,
	'o': EntryTypeEvent,
	'x': EntryTypeDone,
	'~': EntryTypeCancelled,
	'?': EntryTypeQuestion,
	'>': EntryTypeMigrated,
}

type EditableDocumentParser struct {
	dateParser func(string) (time.Time, error)
}

func NewEditableDocumentParser(dateParser func(string) (time.Time, error)) *EditableDocumentParser {
	return &EditableDocumentParser{dateParser: dateParser}
}

func (p *EditableDocumentParser) Parse(input string, existing []Entry) (*EditableDocument, error) {
	lines := strings.Split(input, "\n")
	doc := &EditableDocument{
		Lines:           make([]ParsedLine, 0, len(lines)),
		OriginalMapping: make(map[EntityID]int),
	}

	existingEntityIDs := make(map[EntityID]bool)
	for _, entry := range existing {
		existingEntityIDs[entry.EntityID] = true
	}

	seenEntityIDs := make(map[EntityID]bool)

	contentToEntityID := make(map[string]EntityID)
	for _, entry := range existing {
		contentToEntityID[entry.Content] = entry.EntityID
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

		if parsedLine.IsValid && !parsedLine.IsHeader {
			content, migrateDate, err := ParseMigrationSyntax(line, p.dateParser)
			if err == nil && migrateDate != nil {
				parsedLine.MigrateTarget = migrateDate
				innerLine := p.ParseLine(content, lineNum)
				parsedLine.Symbol = innerLine.Symbol
				parsedLine.Content = innerLine.Content
				parsedLine.Priority = innerLine.Priority
			}

			if parsedLine.EntityID != nil {
				eid := *parsedLine.EntityID
				if seenEntityIDs[eid] {
					parsedLine.EntityID = nil
				} else {
					seenEntityIDs[eid] = true
					doc.OriginalMapping[eid] = lineNum
				}
			} else if entityID, ok := contentToEntityID[parsedLine.Content]; ok {
				parsedLine.EntityID = &entityID
				doc.OriginalMapping[entityID] = lineNum
			}
		}

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

	if !entry.EntityID.IsEmpty() {
		result.WriteString("[")
		result.WriteString(entry.EntityID.String())
		result.WriteString("] ")
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

	if strings.HasPrefix(rest, "[") {
		closeBracket := strings.Index(rest, "] ")
		if closeBracket > 0 {
			idStr := rest[1:closeBracket]
			entityID := EntityID(idStr)
			result.EntityID = &entityID
			rest = rest[closeBracket+2:]
		}
	}

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
