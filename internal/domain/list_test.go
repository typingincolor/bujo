package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList_Validate(t *testing.T) {
	tests := []struct {
		name    string
		list    List
		wantErr bool
	}{
		{
			name: "valid list",
			list: List{
				ID:        1,
				Name:      "Shopping",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid list with spaces in name",
			list: List{
				ID:        1,
				Name:      "Shopping List",
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty name",
			list: List{
				ID:        1,
				Name:      "",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "whitespace only name",
			list: List{
				ID:        1,
				Name:      "   ",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.list.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewList(t *testing.T) {
	list := NewList("Shopping List")

	assert.Equal(t, "Shopping List", list.Name)
	assert.False(t, list.CreatedAt.IsZero())
}

func TestNewList_TrimsWhitespace(t *testing.T) {
	list := NewList("  Shopping List  ")

	assert.Equal(t, "Shopping List", list.Name)
}
