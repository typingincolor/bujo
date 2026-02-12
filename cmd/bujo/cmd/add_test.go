package cmd

import (
	"bytes"
	"testing"
)

func TestWriteEntryIDs(t *testing.T) {
	t.Run("writes each ID on its own line", func(t *testing.T) {
		var buf bytes.Buffer
		writeEntryIDs(&buf, []int64{42, 99, 7})
		got := buf.String()
		want := "42\n99\n7\n"
		if got != want {
			t.Errorf("writeEntryIDs() = %q, want %q", got, want)
		}
	})

	t.Run("empty IDs writes nothing", func(t *testing.T) {
		var buf bytes.Buffer
		writeEntryIDs(&buf, []int64{})
		got := buf.String()
		if got != "" {
			t.Errorf("writeEntryIDs() = %q, want empty", got)
		}
	})

	t.Run("single ID", func(t *testing.T) {
		var buf bytes.Buffer
		writeEntryIDs(&buf, []int64{123})
		got := buf.String()
		want := "123\n"
		if got != want {
			t.Errorf("writeEntryIDs() = %q, want %q", got, want)
		}
	})
}
