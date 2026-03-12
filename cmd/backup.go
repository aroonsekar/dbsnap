package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aroonsekar/dbsnap/internal/config"
	"github.com/aroonsekar/dbsnap/internal/logger"
	"github.com/aroonsekar/dbsnap/internal/mongo"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:    "backup",
	Short:  "Execute the backup and purge sequence",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		runBackupPipeline()
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func runBackupPipeline() {
	cfg, uri, err := config.Load()
	if err != nil {
		fmt.Printf("Initialization error: %v\n", err)
		os.Exit(1)
	}

	// 1. Initialize Logging and Rotation
	logger.Init(cfg.BackupPath)
	defer logger.Close()
	logger.Rotate(cfg.BackupPath, 30)

	logger.Info("Starting dbSnap scheduled execution pipeline")

	// 2. Connect to Database for Purge Operations
	dbClient, err := mongo.Connect(uri)
	if err != nil {
		logger.Error("Database connection failed: %v", err)
		os.Exit(1)
	}
	defer dbClient.Disconnect(nil)

	// 3. Setup Backup Directory
	timestamp := time.Now().Format("20060102_150405")
	currentBackupDir := filepath.Join(cfg.BackupPath, timestamp)

	// 4. Execute Iterative Backup & Purge across all Targets
	for _, target := range cfg.Targets {
		logger.Info("Initiating backup for database: %s", target.DatabaseName)

		for _, coll := range target.Collections {
			logger.Info("Processing collection: %s.%s", target.DatabaseName, coll)

			// Execute mongodump natively with gzip compression
			dumpCmd := exec.Command("mongodump",
				"--uri", uri,
				"--db", target.DatabaseName,
				"--collection", coll,
				"--gzip",
				"--out", currentBackupDir,
			)

			if output, err := dumpCmd.CombinedOutput(); err != nil {
				logger.Error("Backup failed for '%s.%s'. Skipping purge. Error: %v\nOutput: %s", target.DatabaseName, coll, err, string(output))
				continue // Isolate and continue
			}

			logger.Info("Backup verified for '%s.%s'", target.DatabaseName, coll)

			// Execute conditional purge
			if target.PurgeAfter {
				deletedCount, err := mongo.PurgeCollection(dbClient, target.DatabaseName, coll)
				if err != nil {
					logger.Error("Failed to purge collection '%s.%s': %v", target.DatabaseName, coll, err)
				} else {
					logger.Info("Successfully purged %d records from '%s.%s'", deletedCount, target.DatabaseName, coll)
				}
			}
		}
	}

	// 5. Enforce Configured Backup Retention
	logger.Info("Enforcing %d-day retention policy", cfg.RetentionDays)
	cleanupOldBackups(cfg.BackupPath, cfg.RetentionDays)

	logger.Info("Pipeline execution completed successfully")
}

func cleanupOldBackups(backupPath string, retentionDays int) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	entries, err := os.ReadDir(backupPath)
	if err != nil {
		logger.Error("Failed to read backup directory for cleanup: %v", err)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			targetDir := filepath.Join(backupPath, entry.Name())
			if err := os.RemoveAll(targetDir); err != nil {
				logger.Error("Failed to remove expired backup '%s': %v", entry.Name(), err)
			} else {
				logger.Info("Removed expired backup: %s", entry.Name())
			}
		}
	}
}
