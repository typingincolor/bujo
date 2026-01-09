package tui

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type TUIConfig struct {
	DefaultView string `yaml:"default_view"`
	Theme       string `yaml:"theme"`
	DateFormat  string `yaml:"date_format"`
	ShowHelp    bool   `yaml:"show_help"`
}

func DefaultTUIConfig() TUIConfig {
	return TUIConfig{
		DefaultView: "journal",
		Theme:       "default",
		DateFormat:  "Mon, Jan 2 2006",
		ShowHelp:    true,
	}
}

func LoadTUIConfig() TUIConfig {
	config := DefaultTUIConfig()

	// Try loading from ~/.config/bujo/config.yaml first
	configDir, err := os.UserConfigDir()
	if err == nil {
		configPath := filepath.Join(configDir, "bujo", "config.yaml")
		if loadedConfig, ok := loadConfigFile(configPath); ok {
			return mergeConfig(config, loadedConfig)
		}
	}

	// Fall back to ~/.bujo/config.yaml
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".bujo", "config.yaml")
		if loadedConfig, ok := loadConfigFile(configPath); ok {
			return mergeConfig(config, loadedConfig)
		}
	}

	return config
}

func LoadTUIConfigFromPath(path string) TUIConfig {
	config := DefaultTUIConfig()
	if loadedConfig, ok := loadConfigFile(path); ok {
		return mergeConfig(config, loadedConfig)
	}
	return config
}

func loadConfigFile(path string) (TUIConfig, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TUIConfig{}, false
	}

	var config TUIConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return TUIConfig{}, false
	}

	return config, true
}

func mergeConfig(base, override TUIConfig) TUIConfig {
	if override.DefaultView != "" {
		base.DefaultView = override.DefaultView
	}
	if override.Theme != "" {
		base.Theme = override.Theme
	}
	if override.DateFormat != "" {
		base.DateFormat = override.DateFormat
	}
	// ShowHelp is a boolean, so we always take the override value
	// if it was explicitly set in the config file
	// For now, we simply use the override if the file was loaded
	base.ShowHelp = override.ShowHelp

	return base
}

func (c TUIConfig) GetViewType() ViewType {
	switch c.DefaultView {
	case "habits":
		return ViewTypeHabits
	case "lists":
		return ViewTypeLists
	default:
		return ViewTypeJournal
	}
}

func ConfigPaths() []string {
	var paths []string

	configDir, err := os.UserConfigDir()
	if err == nil {
		paths = append(paths, filepath.Join(configDir, "bujo", "config.yaml"))
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(homeDir, ".bujo", "config.yaml"))
	}

	return paths
}
