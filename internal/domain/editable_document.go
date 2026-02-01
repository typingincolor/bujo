package domain

type ParsedLine struct {
	LineNumber   int
	Raw          string
	Depth        int
	Symbol       EntryType
	Priority     Priority
	Content      string
	IsValid      bool
	IsHeader     bool
	ErrorMessage string
}

type EditableDocument struct {
	Lines []ParsedLine
}

type ParseError struct {
	LineNumber int
	Message    string
}
