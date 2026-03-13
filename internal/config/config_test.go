package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

func TestSaveAndLoad(t *testing.T) {
	// Initialize memory mock for OS keyring to prevent UI permission dialogs
	keyring.MockInit()

	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	cfg := AppConfig{
		BackupPath:    "/tmp/backup",
		RetentionDays: 7,
		Targets: []DBTarget{
			{DatabaseName: "testdb"},
		},
	}
	mongoURI := "mongodb://localhost:27017"

	err := Save(cfg, mongoURI)
	assert.NoError(t, err)

	// Verify file was physically written
	configPath := filepath.Join(tempHome, ".dbsnap", configName)
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	loadedCfg, loadedURI, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, cfg, loadedCfg)
	assert.Equal(t, mongoURI, loadedURI)
}

func TestLoad_NoConfig(t *testing.T) {
	keyring.MockInit()

	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	_, _, err := Load()
	assert.ErrorContains(t, err, "config not found")
}
