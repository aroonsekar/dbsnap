package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/aroonsekar/dbsnap/internal/config"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore database from a historical backup",
	Run: func(cmd *cobra.Command, args []string) {
		runRestoreInteractive()
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func runRestoreInteractive() {
	cfg, uri, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}

	// 1. Scan the backup directory
	entries, err := os.ReadDir(cfg.BackupPath)
	if err != nil {
		fmt.Printf("❌ Failed to read backup directory: %v\n", err)
		return
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			backups = append(backups, entry.Name())
		}
	}

	if len(backups) == 0 {
		fmt.Println("No backups found to restore in the target directory.")
		return
	}

	// Sort descending so the newest backups are at the top of the menu
	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j]
	})

	// Dynamically generate the huh selection options
	options := make([]huh.Option[string], len(backups))
	for i, b := range backups {
		options[i] = huh.NewOption(b, b)
	}

	var selectedBackup string
	var wipeFirst bool

	// 2. Interactive Selection Menu
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a backup to restore (Use arrow keys):").
				Options(options...).
				Value(&selectedBackup),
			huh.NewConfirm().
				Title("Wipe existing database collections before restoring? (Destructive)").
				Value(&wipeFirst),
		),
	)

	if err := form.Run(); err != nil {
		fmt.Println("Restore aborted.")
		return
	}

	// mongodump creates a folder matching the target database name inside the timestamp folder
	targetRestorePath := filepath.Join(cfg.BackupPath, selectedBackup)

	// 3. Execute mongorestore
	fmt.Printf("\n[System] Restoring data from %s...\n", selectedBackup)

	// Build the command arguments explicitly
	argsStr := []string{
		"--uri", uri,
		"--gzip",
	}

	if wipeFirst {
		argsStr = append(argsStr, "--drop")
	}

	// Explicitly declare the target directory
	argsStr = append(argsStr, "--dir", targetRestorePath)

	restoreTask := exec.Command("mongorestore", argsStr...)

	// Stream standard output and errors directly to the terminal
	restoreTask.Stdout = os.Stdout
	restoreTask.Stderr = os.Stderr

	if err := restoreTask.Run(); err != nil {
		fmt.Printf("\n❌ Restore failed: %v\n", err)
		return
	}

	fmt.Println("\n✅ Database restore completed successfully.")
}
