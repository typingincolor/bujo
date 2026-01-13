package domain

import (
	"testing"
	"time"
)

func TestNewExportOptions(t *testing.T) {
	opts := NewExportOptions()

	if opts.DateFrom != nil {
		t.Errorf("expected DateFrom to be nil, got %v", opts.DateFrom)
	}
	if opts.DateTo != nil {
		t.Errorf("expected DateTo to be nil, got %v", opts.DateTo)
	}
}

func TestExportOptions_WithDateRange(t *testing.T) {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	opts := NewExportOptions().WithDateRange(from, to)

	if opts.DateFrom == nil {
		t.Fatal("expected DateFrom to be set")
	}
	if !opts.DateFrom.Equal(from) {
		t.Errorf("expected DateFrom to be %v, got %v", from, *opts.DateFrom)
	}

	if opts.DateTo == nil {
		t.Fatal("expected DateTo to be set")
	}
	if !opts.DateTo.Equal(to) {
		t.Errorf("expected DateTo to be %v, got %v", to, *opts.DateTo)
	}
}

func TestNewImportOptions(t *testing.T) {
	tests := []struct {
		name string
		mode ImportMode
		want ImportMode
	}{
		{
			name: "merge mode",
			mode: ImportModeMerge,
			want: ImportModeMerge,
		},
		{
			name: "replace mode",
			mode: ImportModeReplace,
			want: ImportModeReplace,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewImportOptions(tt.mode)
			if opts.Mode != tt.want {
				t.Errorf("expected Mode to be %v, got %v", tt.want, opts.Mode)
			}
		})
	}
}
