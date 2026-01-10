package domain

import "time"

const ExportVersion = "1.0"

type ExportData struct {
	Version     string       `json:"version"`
	ExportedAt  time.Time    `json:"exported_at"`
	Entries     []Entry      `json:"entries"`
	Habits      []Habit      `json:"habits"`
	HabitLogs   []HabitLog   `json:"habit_logs"`
	DayContexts []DayContext `json:"day_contexts"`
	Summaries   []Summary    `json:"summaries"`
	Lists       []List       `json:"lists"`
	ListItems   []ListItem   `json:"list_items"`
	Goals       []Goal       `json:"goals"`
}

type ExportOptions struct {
	DateFrom *time.Time
	DateTo   *time.Time
}

func NewExportOptions() ExportOptions {
	return ExportOptions{}
}

func (o ExportOptions) WithDateRange(from, to time.Time) ExportOptions {
	o.DateFrom = &from
	o.DateTo = &to
	return o
}

type ImportMode string

const (
	ImportModeMerge   ImportMode = "merge"
	ImportModeReplace ImportMode = "replace"
)

type ImportOptions struct {
	Mode ImportMode
}

func NewImportOptions(mode ImportMode) ImportOptions {
	return ImportOptions{Mode: mode}
}
