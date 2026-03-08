package remarkable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeOCRIndentation_FlattensSkippedDepths(t *testing.T) {
	input := "heading\n    deeply indented"
	result := NormalizeOCRIndentation(input)
	assert.Equal(t, "- heading\n  - deeply indented", result)
}

func TestNormalizeOCRIndentation_PreservesBujoMarkers(t *testing.T) {
	input := ". buy milk\n  - organic"
	result := NormalizeOCRIndentation(input)
	assert.Equal(t, ". buy milk\n  - organic", result)
}

func TestNormalizeOCRIndentation_AddsNotePrefix(t *testing.T) {
	input := "plain text"
	result := NormalizeOCRIndentation(input)
	assert.Equal(t, "- plain text", result)
}

func TestNormalizeOCRIndentation_SkipsEmptyLines(t *testing.T) {
	input := ". first\n\n. second"
	result := NormalizeOCRIndentation(input)
	assert.Equal(t, ". first\n. second", result)
}

func TestNormalizeOCRIndentation_TabsConvertToSpaces(t *testing.T) {
	input := ". root\n\t. child"
	result := NormalizeOCRIndentation(input)
	assert.Equal(t, ". root\n  . child", result)
}
