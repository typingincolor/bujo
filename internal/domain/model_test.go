package domain

import (
	"testing"
)

func TestParseModelSpec(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantVariant string
		wantErr     bool
	}{
		{
			name:        "model with variant",
			input:       "llama3.2:3b",
			wantName:    "llama3.2",
			wantVariant: "3b",
			wantErr:     false,
		},
		{
			name:        "model without variant",
			input:       "tinyllama",
			wantName:    "tinyllama",
			wantVariant: "",
			wantErr:     false,
		},
		{
			name:        "model with multiple colons",
			input:       "mistral:7b:q4",
			wantName:    "mistral",
			wantVariant: "7b:q4",
			wantErr:     false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only colon",
			input:   ":",
			wantErr: true,
		},
		{
			name:    "colon at start",
			input:   ":3b",
			wantErr: true,
		},
		{
			name:    "whitespace",
			input:   "  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseModelSpec(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseModelSpec() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseModelSpec() unexpected error: %v", err)
				return
			}

			if got.Name != tt.wantName {
				t.Errorf("ParseModelSpec() name = %v, want %v", got.Name, tt.wantName)
			}

			if got.Variant != tt.wantVariant {
				t.Errorf("ParseModelSpec() variant = %v, want %v", got.Variant, tt.wantVariant)
			}
		})
	}
}

func TestModelSpec_String(t *testing.T) {
	tests := []struct {
		name     string
		spec     ModelSpec
		expected string
	}{
		{
			name:     "with variant",
			spec:     ModelSpec{Name: "llama3.2", Variant: "3b"},
			expected: "llama3.2:3b",
		},
		{
			name:     "without variant",
			spec:     ModelSpec{Name: "tinyllama", Variant: ""},
			expected: "tinyllama",
		},
		{
			name:     "with complex variant",
			spec:     ModelSpec{Name: "mistral", Variant: "7b:q4"},
			expected: "mistral:7b:q4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.String()
			if got != tt.expected {
				t.Errorf("ModelSpec.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestModelSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spec    ModelSpec
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid with variant",
			spec:    ModelSpec{Name: "llama3.2", Variant: "3b"},
			wantErr: false,
		},
		{
			name:    "valid without variant",
			spec:    ModelSpec{Name: "tinyllama", Variant: ""},
			wantErr: false,
		},
		{
			name:    "invalid - empty name",
			spec:    ModelSpec{Name: "", Variant: "3b"},
			wantErr: true,
			errMsg:  "model name cannot be empty",
		},
		{
			name:    "invalid - whitespace name",
			spec:    ModelSpec{Name: "  ", Variant: "3b"},
			wantErr: true,
			errMsg:  "model name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ModelSpec.Validate() expected error, got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("ModelSpec.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ModelSpec.Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestParseModelVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantMajor int
		wantMinor int
		wantPatch int
		wantErr   bool
	}{
		{
			name:      "full version",
			input:     "1.2.3",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 3,
			wantErr:   false,
		},
		{
			name:      "major.minor only",
			input:     "1.2",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:      "major only",
			input:     "1",
			wantMajor: 1,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "negative number",
			input:   "-1.0.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseModelVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseModelVersion() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseModelVersion() unexpected error: %v", err)
				return
			}

			if got.Major != tt.wantMajor {
				t.Errorf("ParseModelVersion() major = %v, want %v", got.Major, tt.wantMajor)
			}

			if got.Minor != tt.wantMinor {
				t.Errorf("ParseModelVersion() minor = %v, want %v", got.Minor, tt.wantMinor)
			}

			if got.Patch != tt.wantPatch {
				t.Errorf("ParseModelVersion() patch = %v, want %v", got.Patch, tt.wantPatch)
			}
		})
	}
}

func TestModelVersion_String(t *testing.T) {
	tests := []struct {
		name     string
		version  ModelVersion
		expected string
	}{
		{
			name:     "full version",
			version:  ModelVersion{Major: 1, Minor: 2, Patch: 3},
			expected: "1.2.3",
		},
		{
			name:     "zero patch",
			version:  ModelVersion{Major: 1, Minor: 2, Patch: 0},
			expected: "1.2.0",
		},
		{
			name:     "all zeros",
			version:  ModelVersion{Major: 0, Minor: 0, Patch: 0},
			expected: "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.version.String()
			if got != tt.expected {
				t.Errorf("ModelVersion.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestModelVersion_NewerThan(t *testing.T) {
	tests := []struct {
		name     string
		v1       ModelVersion
		v2       ModelVersion
		expected bool
	}{
		{
			name:     "major version newer",
			v1:       ModelVersion{Major: 2, Minor: 0, Patch: 0},
			v2:       ModelVersion{Major: 1, Minor: 9, Patch: 9},
			expected: true,
		},
		{
			name:     "minor version newer",
			v1:       ModelVersion{Major: 1, Minor: 2, Patch: 0},
			v2:       ModelVersion{Major: 1, Minor: 1, Patch: 9},
			expected: true,
		},
		{
			name:     "patch version newer",
			v1:       ModelVersion{Major: 1, Minor: 1, Patch: 2},
			v2:       ModelVersion{Major: 1, Minor: 1, Patch: 1},
			expected: true,
		},
		{
			name:     "same version",
			v1:       ModelVersion{Major: 1, Minor: 2, Patch: 3},
			v2:       ModelVersion{Major: 1, Minor: 2, Patch: 3},
			expected: false,
		},
		{
			name:     "older version",
			v1:       ModelVersion{Major: 1, Minor: 0, Patch: 0},
			v2:       ModelVersion{Major: 1, Minor: 2, Patch: 3},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.NewerThan(tt.v2)
			if got != tt.expected {
				t.Errorf("ModelVersion.NewerThan() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestModelInfo_Validate(t *testing.T) {
	validSpec := ModelSpec{Name: "llama3.2", Variant: "3b"}
	validVersion := ModelVersion{Major: 1, Minor: 0, Patch: 0}

	tests := []struct {
		name    string
		info    ModelInfo
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid model info",
			info: ModelInfo{
				Spec:        validSpec,
				Version:     validVersion,
				Size:        2147483648,
				Description: "A great model",
				HFRepo:      "TheBloke/Llama-3.2-3B-GGUF",
				HFFile:      "llama3.2-3b-q4.gguf",
			},
			wantErr: false,
		},
		{
			name: "invalid spec",
			info: ModelInfo{
				Spec:    ModelSpec{Name: "", Variant: "3b"},
				Version: validVersion,
				Size:    1000,
			},
			wantErr: true,
			errMsg:  "invalid model spec: model name cannot be empty",
		},
		{
			name: "zero size",
			info: ModelInfo{
				Spec:    validSpec,
				Version: validVersion,
				Size:    0,
			},
			wantErr: true,
			errMsg:  "model size must be positive",
		},
		{
			name: "negative size",
			info: ModelInfo{
				Spec:    validSpec,
				Version: validVersion,
				Size:    -1000,
			},
			wantErr: true,
			errMsg:  "model size must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.info.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("ModelInfo.Validate() expected error, got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("ModelInfo.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ModelInfo.Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestModelInfo_HasUpdate(t *testing.T) {
	tests := []struct {
		name     string
		info     ModelInfo
		expected bool
	}{
		{
			name: "has update",
			info: ModelInfo{
				Version:      ModelVersion{Major: 1, Minor: 1, Patch: 0},
				LocalVersion: &ModelVersion{Major: 1, Minor: 0, Patch: 0},
			},
			expected: true,
		},
		{
			name: "no update - same version",
			info: ModelInfo{
				Version:      ModelVersion{Major: 1, Minor: 0, Patch: 0},
				LocalVersion: &ModelVersion{Major: 1, Minor: 0, Patch: 0},
			},
			expected: false,
		},
		{
			name: "no local version",
			info: ModelInfo{
				Version:      ModelVersion{Major: 1, Minor: 0, Patch: 0},
				LocalVersion: nil,
			},
			expected: false,
		},
		{
			name: "local version newer (edge case)",
			info: ModelInfo{
				Version:      ModelVersion{Major: 1, Minor: 0, Patch: 0},
				LocalVersion: &ModelVersion{Major: 1, Minor: 1, Patch: 0},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.HasUpdate()
			if got != tt.expected {
				t.Errorf("ModelInfo.HasUpdate() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestModelInfo_IsDownloaded(t *testing.T) {
	tests := []struct {
		name     string
		info     ModelInfo
		expected bool
	}{
		{
			name: "downloaded - has local path",
			info: ModelInfo{
				LocalPath: "/home/user/.bujo/models/llama3.2-3b.gguf",
			},
			expected: true,
		},
		{
			name: "not downloaded - empty local path",
			info: ModelInfo{
				LocalPath: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.IsDownloaded()
			if got != tt.expected {
				t.Errorf("ModelInfo.IsDownloaded() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAvailableModels(t *testing.T) {
	models := AvailableModels()

	if len(models) == 0 {
		t.Fatal("AvailableModels() returned empty list")
	}

	expectedModels := map[string]bool{
		"tinyllama":   true,
		"llama3.2:1b": true,
		"llama3.2:3b": true,
		"phi-3-mini":  true,
		"mistral:7b":  true,
	}

	for _, model := range models {
		modelName := model.Spec.String()
		if !expectedModels[modelName] {
			t.Errorf("AvailableModels() contains unexpected model: %s", modelName)
		}

		if err := model.Validate(); err != nil {
			t.Errorf("AvailableModels() contains invalid model %s: %v", modelName, err)
		}

		delete(expectedModels, modelName)
	}

	for missing := range expectedModels {
		t.Errorf("AvailableModels() missing expected model: %s", missing)
	}
}
