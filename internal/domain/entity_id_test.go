package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntityID_ReturnsNonEmpty(t *testing.T) {
	id := NewEntityID()

	assert.NotEmpty(t, id.String())
}

func TestNewEntityID_ReturnsUniqueValues(t *testing.T) {
	id1 := NewEntityID()
	id2 := NewEntityID()

	assert.NotEqual(t, id1, id2)
}

func TestEntityID_String_ReturnsValue(t *testing.T) {
	id := NewEntityID()

	str := id.String()

	assert.NotEmpty(t, str)
	assert.Len(t, str, 36) // UUID format: 8-4-4-4-12
}

func TestParseEntityID_ValidUUID_Succeeds(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	id, err := ParseEntityID(validUUID)

	require.NoError(t, err)
	assert.Equal(t, validUUID, id.String())
}

func TestParseEntityID_InvalidUUID_Fails(t *testing.T) {
	invalidUUID := "not-a-uuid"

	_, err := ParseEntityID(invalidUUID)

	require.Error(t, err)
}

func TestParseEntityID_EmptyString_Fails(t *testing.T) {
	_, err := ParseEntityID("")

	require.Error(t, err)
}

func TestEntityID_IsEmpty_WhenEmpty_ReturnsTrue(t *testing.T) {
	var id EntityID

	assert.True(t, id.IsEmpty())
}

func TestEntityID_IsEmpty_WhenSet_ReturnsFalse(t *testing.T) {
	id := NewEntityID()

	assert.False(t, id.IsEmpty())
}
