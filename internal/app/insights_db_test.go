package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenInsightsDB_MissingFile(t *testing.T) {
	db, err := OpenInsightsDB("/nonexistent/path/insights.db")
	assert.NoError(t, err)
	assert.Nil(t, db)
}
