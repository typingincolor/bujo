package cmd

import (
	"testing"
	"time"
)

func TestValidateDateRange(t *testing.T) {
	tests := []struct {
		name    string
		from    time.Time
		to      time.Time
		wantErr bool
	}{
		{
			name:    "valid range",
			from:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			to:      time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "same day is valid",
			from:    time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			to:      time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "from after to is invalid",
			from:    time.Date(2026, 1, 7, 0, 0, 0, 0, time.UTC),
			to:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateRange(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDateRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
