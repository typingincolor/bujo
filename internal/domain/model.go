package domain

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ModelSpec struct {
	Name    string
	Variant string
}

func ParseModelSpec(s string) (ModelSpec, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ModelSpec{}, errors.New("model name cannot be empty")
	}

	parts := strings.SplitN(s, ":", 2)
	if parts[0] == "" {
		return ModelSpec{}, errors.New("model name cannot be empty")
	}

	spec := ModelSpec{
		Name: parts[0],
	}

	if len(parts) > 1 {
		spec.Variant = parts[1]
	}

	return spec, nil
}

func (m ModelSpec) String() string {
	if m.Variant == "" {
		return m.Name
	}
	return m.Name + ":" + m.Variant
}

func (m ModelSpec) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("model name cannot be empty")
	}
	return nil
}

type ModelVersion struct {
	Major int
	Minor int
	Patch int
}

func ParseModelVersion(s string) (ModelVersion, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ModelVersion{}, errors.New("version string cannot be empty")
	}

	parts := strings.Split(s, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return ModelVersion{}, errors.New("invalid version format")
	}

	var version ModelVersion
	var err error

	if len(parts) >= 1 {
		version.Major, err = strconv.Atoi(parts[0])
		if err != nil || version.Major < 0 {
			return ModelVersion{}, errors.New("invalid major version")
		}
	}

	if len(parts) >= 2 {
		version.Minor, err = strconv.Atoi(parts[1])
		if err != nil || version.Minor < 0 {
			return ModelVersion{}, errors.New("invalid minor version")
		}
	}

	if len(parts) >= 3 {
		version.Patch, err = strconv.Atoi(parts[2])
		if err != nil || version.Patch < 0 {
			return ModelVersion{}, errors.New("invalid patch version")
		}
	}

	return version, nil
}

func (v ModelVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v ModelVersion) NewerThan(other ModelVersion) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

type ModelInfo struct {
	Spec         ModelSpec
	Version      ModelVersion
	Size         int64
	Description  string
	HFRepo       string
	HFFile       string
	LocalPath    string
	LocalVersion *ModelVersion
}

func (m ModelInfo) Validate() error {
	if err := m.Spec.Validate(); err != nil {
		return fmt.Errorf("invalid model spec: %w", err)
	}
	if m.Size <= 0 {
		return errors.New("model size must be positive")
	}
	return nil
}

func (m ModelInfo) HasUpdate() bool {
	if m.LocalVersion == nil {
		return false
	}
	return m.Version.NewerThan(*m.LocalVersion)
}

func (m ModelInfo) IsDownloaded() bool {
	return m.LocalPath != ""
}

func AvailableModels() []ModelInfo {
	return []ModelInfo{
		{
			Spec:        ModelSpec{Name: "tinyllama", Variant: ""},
			Version:     ModelVersion{Major: 1, Minor: 0, Patch: 0},
			Size:        637 * 1024 * 1024,
			Description: "Fast, good for testing",
			HFRepo:      "TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF",
			HFFile:      "tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf",
		},
		{
			Spec:        ModelSpec{Name: "llama3.2", Variant: "1b"},
			Version:     ModelVersion{Major: 1, Minor: 0, Patch: 0},
			Size:        1300 * 1024 * 1024,
			Description: "Good balance - recommended",
			HFRepo:      "TheBloke/Llama-3.2-1B-GGUF",
			HFFile:      "llama-3.2-1b.Q4_K_M.gguf",
		},
		{
			Spec:        ModelSpec{Name: "llama3.2", Variant: "3b"},
			Version:     ModelVersion{Major: 1, Minor: 0, Patch: 0},
			Size:        2000 * 1024 * 1024,
			Description: "Better quality",
			HFRepo:      "TheBloke/Llama-3.2-3B-GGUF",
			HFFile:      "llama-3.2-3b.Q4_K_M.gguf",
		},
		{
			Spec:        ModelSpec{Name: "phi-3-mini", Variant: ""},
			Version:     ModelVersion{Major: 1, Minor: 0, Patch: 0},
			Size:        2300 * 1024 * 1024,
			Description: "Microsoft, good reasoning",
			HFRepo:      "microsoft/Phi-3-mini-4k-instruct-gguf",
			HFFile:      "Phi-3-mini-4k-instruct-q4.gguf",
		},
		{
			Spec:        ModelSpec{Name: "mistral", Variant: "7b"},
			Version:     ModelVersion{Major: 1, Minor: 0, Patch: 0},
			Size:        4100 * 1024 * 1024,
			Description: "High quality, needs more RAM",
			HFRepo:      "TheBloke/Mistral-7B-Instruct-v0.2-GGUF",
			HFFile:      "mistral-7b-instruct-v0.2.Q4_K_M.gguf",
		},
	}
}
