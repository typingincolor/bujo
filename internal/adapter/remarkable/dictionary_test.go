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
