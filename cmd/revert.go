package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aroonsekar/dbsnap/internal/scheduler"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var revertCmd = &cobra.Command{
	Use:   "revert",
	Short: "Remove scheduled tasks and wipe configuration",
	Run: func(cmd *cobra.Command, args []string) {
		runRevertInteractive()
	},
}

func init() {
	rootCmd.AddCommand(revertCmd)
}

func runRevertInteractive() {
	var confirm bool

	// 1. Safety Check
	err := huh.NewConfirm().
		Title("Are you sure you want to remove dbSnap's scheduled task and configuration? \n(Your existing backup files will NOT be deleted)").
		Value(&confirm).
		Run()

	if err != nil || !confirm {
		fmt.Println("Revert aborted.")
		return
	}

	fmt.Println("\n[System] Initiating teardown sequence...")

	// 2. Remove Windows Scheduled Task
	if !scheduler.CheckAdmin() {
		fmt.Println("❌ Teardown blocked: Administrator privileges required to remove scheduled tasks.")
		fmt.Println("Please right-click your terminal, select 'Run as Administrator', and try again.")
		return
	}

	taskName := "dbSnap_NightlyBackup"
	deleteTaskCmd := exec.Command("schtasks", "/delete", "/tn", taskName, "/f")

	if output, err := deleteTaskCmd.CombinedOutput(); err != nil {
		// It might fail if the task doesn't exist, which is fine
		fmt.Printf("⚠️  Could not remove scheduled task (it may not exist): %v\n", string(output))
	} else {
		fmt.Println("✅ Windows scheduled task removed.")
	}

	// 3. Wipe Secure Credentials from OS Vault
	err = keyring.Delete("dbSnap", "mongo_uri")
	if err != nil {
		if err == keyring.ErrNotFound {
			fmt.Println("✅ OS Vault already clean.")
		} else {
			fmt.Printf("⚠️  Failed to remove credentials from OS Vault: %v\n", err)
		}
	} else {
		fmt.Println("✅ Credentials wiped from OS Vault.")
	}

	// 4. Delete the config.json file
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".dbsnap", "config.json")

	if err := os.Remove(configPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("✅ Configuration file already clean.")
		} else {
			fmt.Printf("⚠️  Failed to delete config file: %v\n", err)
		}
	} else {
		fmt.Println("✅ Configuration file removed.")
	}

	fmt.Println("\n✅ dbSnap has been successfully reverted to a clean state.")
}
