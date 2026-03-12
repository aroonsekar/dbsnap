package scheduler

import (
	"fmt"
	"os"
	"os/exec"
)

// CheckAdmin verifies if the CLI is running with elevated privileges on Windows.
// Using 'net session' is a lightweight, standard hack to check admin rights without heavy CGO libraries.
func CheckAdmin() bool {
	cmd := exec.Command("net", "session")
	err := cmd.Run()
	return err == nil
}

// SetupWindowsTask creates or overwrites the daily scheduled task.
func SetupWindowsTask(time string) error {
	if !CheckAdmin() {
		return fmt.Errorf("setup blocked: Administrator privileges required.\nPlease close this terminal, right-click your terminal app, select 'Run as Administrator', and try again")
	}

	// Get the absolute path to the currently running dbSnap binary
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %v", err)
	}

	taskName := "dbSnap_NightlyBackup"

	// schtasks arguments:
	// /tn = Task Name
	// /tr = Task Run (the command to execute)
	// /sc = Schedule type (daily)
	// /st = Start time (HH:mm)
	// /f  = Force overwrite if the task already exists
	// /rl highest = Run with maximum privileges (required for background file writes)
	cmd := exec.Command("schtasks", "/create",
		"/tn", taskName,
		"/tr", fmt.Sprintf("\"%s\" backup", executablePath),
		"/sc", "daily",
		"/st", time,
		"/f",
		"/rl", "highest",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create scheduled task: %v\n%s", err, string(output))
	}

	fmt.Printf("[System] Successfully scheduled Windows task '%s' to run daily at %s.\n", taskName, time)
	return nil
}
