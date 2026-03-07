package remarkable

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "remarkable.json")

	cfg := Config{
		DeviceToken: "test-device-token",
		DeviceID:    "test-device-id",
	}

	err := SaveConfig(path, cfg)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "test-device-token")
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "remarkable.json")

	original := Config{
		DeviceToken: "saved-token",
		DeviceID:    "saved-id",
	}
	err := SaveConfig(path, original)
	require.NoError(t, err)

	loaded, err := LoadConfig(path)
	require.NoError(t, err)
	assert.Equal(t, original, loaded)
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/remarkable.json")
	assert.Error(t, err)
}
