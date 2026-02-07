package domain

import (
	"testing"
	"time"
)

func TestNewSearchOptions(t *testing.T) {
	query := "test query"
	opts := NewSearchOptions(query)

	if opts.Query != query {
		t.Errorf("expected Query to be %q, got %q", query, opts.Query)
	}
	if opts.Limit != 50 {
		t.Errorf("expected default Limit to be 50, got %d", opts.Limit)
	}
	if opts.Type != nil {
		t.Errorf("expected Type to be nil, got %v", opts.Type)
	}
	if opts.DateFrom != nil {
		t.Errorf("expected DateFrom to be nil, got %v", opts.DateFrom)
	}
	if opts.DateTo != nil {
		t.Errorf("expected DateTo to be nil, got %v", opts.DateTo)
	}
}

func TestSearchOptions_WithType(t *testing.T) {
	opts := NewSearchOptions("test").WithType(EntryTypeTask)

	if opts.Type == nil {
		t.Fatal("expected Type to be set")
	}
	if *opts.Type != EntryTypeTask {
		t.Errorf("expected Type to be %v, got %v", EntryTypeTask, *opts.Type)
	}
}

func TestSearchOptions_WithDateRange(t *testing.T) {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	opts := NewSearchOptions("test").WithDateRange(from, to)

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

func TestSearchOptions_WithLimit(t *testing.T) {
	opts := NewSearchOptions("test").WithLimit(100)

	if opts.Limit != 100 {
		t.Errorf("expected Limit to be 100, got %d", opts.Limit)
	}
}

func TestSearchOptions_WithTags(t *testing.T) {
	opts := NewSearchOptions("test").WithTags([]string{"shopping", "errands"})

	if opts.Tags == nil {
		t.Fatal("expected Tags to be set")
	}
	if len(opts.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(opts.Tags))
	}
	if opts.Tags[0] != "shopping" || opts.Tags[1] != "errands" {
		t.Errorf("expected tags [shopping, errands], got %v", opts.Tags)
	}
}

func TestNewSearchOptions_TagsNil(t *testing.T) {
	opts := NewSearchOptions("test")
	if opts.Tags != nil {
		t.Errorf("expected Tags to be nil, got %v", opts.Tags)
	}
}

func TestSearchOptions_Chaining(t *testing.T) {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	opts := NewSearchOptions("test query").
		WithType(EntryTypeNote).
		WithDateRange(from, to).
		WithLimit(25)

	if opts.Query != "test query" {
		t.Errorf("expected Query to be %q, got %q", "test query", opts.Query)
	}
	if opts.Type == nil || *opts.Type != EntryTypeNote {
		t.Errorf("expected Type to be EntryTypeNote")
	}
	if opts.DateFrom == nil || !opts.DateFrom.Equal(from) {
		t.Errorf("expected DateFrom to be %v", from)
	}
	if opts.DateTo == nil || !opts.DateTo.Equal(to) {
		t.Errorf("expected DateTo to be %v", to)
	}
	if opts.Limit != 25 {
		t.Errorf("expected Limit to be 25, got %d", opts.Limit)
	}
}
