package remarkable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePageOrder(t *testing.T) {
	contentJSON := `{
		"cPages": {
			"pages": [
				{"id": "page-uuid-1", "idx": {"timestamp": "1:2", "value": "ba"}},
				{"id": "page-uuid-2", "idx": {"timestamp": "1:3", "value": "bb"}}
			]
		},
		"fileType": "notebook",
		"pageCount": 2
	}`

	pages, err := ParsePageOrder([]byte(contentJSON))
	require.NoError(t, err)
	assert.Equal(t, []string{"page-uuid-1", "page-uuid-2"}, pages)
}

func TestParsePageOrderEmpty(t *testing.T) {
	contentJSON := `{"cPages": {"pages": []}, "fileType": "notebook"}`

	pages, err := ParsePageOrder([]byte(contentJSON))
	require.NoError(t, err)
	assert.Empty(t, pages)
}

func TestParsePageOrderInvalidJSON(t *testing.T) {
	_, err := ParsePageOrder([]byte("not json"))
	assert.Error(t, err)
}
