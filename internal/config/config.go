package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "dbSnap"
	accountName = "mongo_uri"
	configName  = "config.json"
)

// DBTarget holds the specific backup and purge rules for a single database
type DBTarget struct {
	DatabaseName string   `json:"database_name"`
	Collections  []string `json:"collections"`
	PurgeAfter   bool     `json:"purge_after"`
}

type AppConfig struct {
	BackupPath    string     `json:"backup_path"`
	RetentionDays int        `json:"retention_days"`
	Targets       []DBTarget `json:"targets"`
}

// Save securely stores the URI in the OS vault and writes the rest to a local JSON file.
func Save(cfg AppConfig, mongoURI string) error {
	if err := keyring.Set(serviceName, accountName, mongoURI); err != nil {
		return fmt.Errorf("failed to save credentials to OS vault: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".dbsnap")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(configDir, configName))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}

// Load retrieves the config and pulls the secure URI back out of the vault.
func Load() (AppConfig, string, error) {
	var cfg AppConfig
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".dbsnap", configName)

	file, err := os.Open(configPath)
	if err != nil {
		return cfg, "", fmt.Errorf("config not found, please run 'dbsnap setup'")
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return cfg, "", err
	}

	uri, err := keyring.Get(serviceName, accountName)
	if err != nil {
		return cfg, "", fmt.Errorf("could not retrieve MongoDB credentials: %v", err)
	}

	return cfg, uri, nil
}
