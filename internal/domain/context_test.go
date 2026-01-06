package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDayContext_Validate(t *testing.T) {
	validDate := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		ctx     DayContext
		wantErr bool
	}{
		{
			name: "valid context with all fields",
			ctx: DayContext{
				Date:     validDate,
				Location: stringPtr("Manchester Office"),
				Mood:     stringPtr("productive"),
				Weather:  stringPtr("cloudy"),
			},
			wantErr: false,
		},
		{
			name: "valid context with only location",
			ctx: DayContext{
				Date:     validDate,
				Location: stringPtr("Home"),
			},
			wantErr: false,
		},
		{
			name: "valid context with no optional fields",
			ctx: DayContext{
				Date: validDate,
			},
			wantErr: false,
		},
		{
			name: "zero date is invalid",
			ctx: DayContext{
				Date:     time.Time{},
				Location: stringPtr("Home"),
			},
			wantErr: true,
		},
		{
			name: "empty location string is invalid",
			ctx: DayContext{
				Date:     validDate,
				Location: stringPtr(""),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ctx.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDayContext_HasLocation(t *testing.T) {
	tests := []struct {
		name     string
		ctx      DayContext
		expected bool
	}{
		{
			name:     "has location when set",
			ctx:      DayContext{Location: stringPtr("Office")},
			expected: true,
		},
		{
			name:     "no location when nil",
			ctx:      DayContext{Location: nil},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ctx.HasLocation())
		})
	}
}

func TestDayContext_GetLocation(t *testing.T) {
	tests := []struct {
		name     string
		ctx      DayContext
		expected string
	}{
		{
			name:     "returns location when set",
			ctx:      DayContext{Location: stringPtr("Office")},
			expected: "Office",
		},
		{
			name:     "returns empty when nil",
			ctx:      DayContext{Location: nil},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ctx.GetLocation())
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
