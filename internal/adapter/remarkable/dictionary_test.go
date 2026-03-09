package remarkable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCommonWord_RecognizesCommonWords(t *testing.T) {
	assert.True(t, isCommonWord("test"))
	assert.True(t, isCommonWord("milk"))
	assert.True(t, isCommonWord("task"))
}

func TestIsCommonWord_RejectsGarbledWords(t *testing.T) {
	assert.False(t, isCommonWord("fest"))
	assert.False(t, isCommonWord("nota"))
	assert.False(t, isCommonWord("fask"))
	assert.False(t, isCommonWord("xyzzy"))
}

func TestIsCommonWord_CaseInsensitive(t *testing.T) {
	assert.True(t, isCommonWord("Test"))
	assert.True(t, isCommonWord("MILK"))
}

func TestHasUnknownWords_AllKnown(t *testing.T) {
	assert.False(t, hasUnknownWords("- can we hook into existing channels?"))
}

func TestHasUnknownWords_GarbledWord(t *testing.T) {
	assert.True(t, hasUnknownWords("- did Stephen I cecil prensentation change anything?"))
}

func TestHasUnknownWords_SkipsProperNouns(t *testing.T) {
	assert.False(t, hasUnknownWords("- @Emma is going to help"))
}

func TestHasUnknownWords_SkipsShortWords(t *testing.T) {
	assert.False(t, hasUnknownWords("- Al is ok"))
}

func TestHasUnknownWords_FlagsMisspelledCommonWords(t *testing.T) {
	assert.True(t, hasUnknownWords("- Benck is going to get us a slot"))
	assert.True(t, hasUnknownWords("- Getty survey going out next week"))
}
