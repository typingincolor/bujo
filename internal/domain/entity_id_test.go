package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntityID_ReturnsUniqueUUIDs(t *testing.T) {
	id1 := NewEntityID()
	id2 := NewEntityID()

	assert.NotEmpty(t, id1.String())
	assert.Len(t, id1.String(), 36) // UUID format: 8-4-4-4-12
	assert.NotEqual(t, id1, id2)
}

func TestParseEntityID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"invalid UUID", "not-a-uuid", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseEntityID(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.input, id.String())
			}
		})
	}
}

func TestEntityID_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		id       EntityID
		expected bool
	}{
		{"empty entity ID", EntityID(""), true},
		{"non-empty", NewEntityID(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.id.IsEmpty())
		})
	}
}
