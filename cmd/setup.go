package cmd

import (
	"fmt"
	"os/exec"

	"github.com/aroonsekar/dbsnap/internal/config"
	"github.com/aroonsekar/dbsnap/internal/mongo"
	"github.com/aroonsekar/dbsnap/internal/scheduler"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure dbSnap and schedule background backups",
	Run: func(cmd *cobra.Command, args []string) {
		runSetupInteractive()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetupInteractive() {
	var backupPath string

	if !checkMongoTools() {
		fmt.Println("❌ MongoDB Database Tools (mongodump) not found in PATH. Please install them.")
		return
	}

	// 1. Path Selection
	err := huh.NewInput().Title("Enter the absolute path where backups should be saved:").Value(&backupPath).Run()
	if err != nil {
		return
	}
	fmt.Printf("Path locked in: %s\n\n", backupPath)

	// 2. Database Connection
	var mongoURI string
	err = huh.NewInput().Title("Enter your MongoDB Connection URI:").Value(&mongoURI).Run()
	if err != nil {
		return
	}

	fmt.Println("[System] Connecting to cluster to fetch databases...")
	dbClient, err := mongo.Connect(mongoURI)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}
	defer dbClient.Disconnect(nil)

	availableDBs, err := mongo.ListDatabases(dbClient)
	if err != nil || len(availableDBs) == 0 {
		fmt.Println("❌ Could not fetch databases or none exist.")
		return
	}

	// 3. Dynamic Selection Loop
	var targets []config.DBTarget
	addAnother := true

	for addAnother {
		var selectedDB string
		dbOptions := make([]huh.Option[string], len(availableDBs))
		for i, db := range availableDBs {
			dbOptions[i] = huh.NewOption(db, db)
		}

		err = huh.NewSelect[string]().Title("Select a Database to backup:").Options(dbOptions...).Value(&selectedDB).Run()
		if err != nil {
			return
		}

		fmt.Printf("[System] Fetching collections for %s...\n", selectedDB)
		availableColls, err := mongo.ListCollections(dbClient, selectedDB)
		if err != nil || len(availableColls) == 0 {
			fmt.Println("❌ Could not fetch collections or none exist.")
			continue
		}

		var selectedColls []string
		collOptions := make([]huh.Option[string], len(availableColls))
		for i, coll := range availableColls {
			collOptions[i] = huh.NewOption(coll, coll)
		}

		err = huh.NewMultiSelect[string]().Title("Select Collections (Space to select, Enter to confirm):").Options(collOptions...).Value(&selectedColls).Run()
		if err != nil {
			return
		}

		var purgeAfter bool
		err = huh.NewConfirm().Title(fmt.Sprintf("Purge records in %s after successful backup? (Destructive)", selectedDB)).Value(&purgeAfter).Run()
		if err != nil {
			return
		}

		targets = append(targets, config.DBTarget{
			DatabaseName: selectedDB,
			Collections:  selectedColls,
			PurgeAfter:   purgeAfter,
		})

		err = huh.NewConfirm().Title("Add another database to this backup job?").Value(&addAnother).Run()
		if err != nil {
			return
		}
	}

	// 4. Secure & Schedule
	cfg := config.AppConfig{
		BackupPath:    backupPath,
		RetentionDays: 2190,
		Targets:       targets,
	}

	fmt.Println("\n[System] Securing credentials in OS Vault...")
	if err := config.Save(cfg, mongoURI); err != nil {
		fmt.Printf("❌ Failed to save config: %v\n", err)
		return
	}

	var executionTime string
	err = huh.NewInput().Title("What time should the nightly backup run? (Format HH:MM)").Value(&executionTime).Validate(func(str string) error {
		if len(str) != 5 || str[2] != ':' {
			return fmt.Errorf("use HH:MM format")
		}
		return nil
	}).Run()
	if err != nil {
		return
	}

	if err = scheduler.SetupWindowsTask(executionTime); err != nil {
		fmt.Printf("\n❌ %v\n", err)
		return
	}
}

func checkMongoTools() bool {
	_, err := exec.LookPath("mongodump")
	return err == nil
}
