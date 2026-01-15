package domain

import (
	"testing"
)

func TestPromptType_String(t *testing.T) {
	tests := []struct {
		name     string
		pt       PromptType
		expected string
	}{
		{
			name:     "summary daily",
			pt:       PromptTypeSummaryDaily,
			expected: "summary-daily",
		},
		{
			name:     "summary weekly",
			pt:       PromptTypeSummaryWeekly,
			expected: "summary-weekly",
		},
		{
			name:     "ask",
			pt:       PromptTypeAsk,
			expected: "ask",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pt.String()
			if got != tt.expected {
				t.Errorf("PromptType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPromptType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		pt       PromptType
		expected bool
	}{
		{
			name:     "valid - summary daily",
			pt:       PromptTypeSummaryDaily,
			expected: true,
		},
		{
			name:     "valid - ask",
			pt:       PromptTypeAsk,
			expected: true,
		},
		{
			name:     "invalid - empty",
			pt:       PromptType(""),
			expected: false,
		},
		{
			name:     "invalid - unknown",
			pt:       PromptType("unknown"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pt.IsValid()
			if got != tt.expected {
				t.Errorf("PromptType.IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPromptTemplate_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tmpl    PromptTemplate
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid template",
			tmpl: PromptTemplate{
				Type:     PromptTypeSummaryDaily,
				Content:  "You are analyzing {{.Horizon}} entries.",
				Filename: "summary-daily.txt",
			},
			wantErr: false,
		},
		{
			name: "invalid - empty type",
			tmpl: PromptTemplate{
				Type:    PromptType(""),
				Content: "Some content",
			},
			wantErr: true,
			errMsg:  "invalid prompt type",
		},
		{
			name: "invalid - unknown type",
			tmpl: PromptTemplate{
				Type:    PromptType("unknown"),
				Content: "Some content",
			},
			wantErr: true,
			errMsg:  "invalid prompt type",
		},
		{
			name: "invalid - empty content",
			tmpl: PromptTemplate{
				Type:    PromptTypeSummaryDaily,
				Content: "",
			},
			wantErr: true,
			errMsg:  "prompt content cannot be empty",
		},
		{
			name: "invalid - whitespace only content",
			tmpl: PromptTemplate{
				Type:    PromptTypeSummaryDaily,
				Content: "   \n\t  ",
			},
			wantErr: true,
			errMsg:  "prompt content cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tmpl.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("PromptTemplate.Validate() expected error, got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("PromptTemplate.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("PromptTemplate.Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPromptTypeFromHorizon(t *testing.T) {
	tests := []struct {
		name     string
		horizon  SummaryHorizon
		expected PromptType
	}{
		{
			name:     "daily",
			horizon:  SummaryHorizonDaily,
			expected: PromptTypeSummaryDaily,
		},
		{
			name:     "weekly",
			horizon:  SummaryHorizonWeekly,
			expected: PromptTypeSummaryWeekly,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PromptTypeFromHorizon(tt.horizon)
			if got != tt.expected {
				t.Errorf("PromptTypeFromHorizon() = %v, want %v", got, tt.expected)
			}
		})
	}
}
