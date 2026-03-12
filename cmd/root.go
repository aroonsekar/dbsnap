package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const asciiBanner = `
     _  _    _____                 
  __| || |__/ ____|                
 / _` + "`" + ` || '_ \ (___  _ __   __ _ _ __ 
| (_| || |_) \___ \| '_ \ / _` + "`" + ` | '_ \
 \__,_||_.__/____) | | | | (_| | |_) |
                   |_| |_|\__,_| .__/ 
                               | |    
                               |_|    
`

var rootCmd = &cobra.Command{
	Use:   "dbsnap",
	Short: "Automated, secure MongoDB backup & restore CLI",
	Long: fmt.Sprintf(`%s
dbSnap v1.0 - Production-Grade MongoDB Automated Backups

DESCRIPTION:
dbSnap is a lightweight, secure CLI that runs in the background to automatically 
backup and optionally purge your MongoDB collections. It connects natively to 
your OS task scheduler and secures credentials in the OS Keyring.

GETTING STARTED:
  1. Setup: Run 'dbsnap setup' to configure your connection, select databases, 
     and schedule the background task.
  2. Restore: Run 'dbsnap restore' to recover data from a specific timestamp.
  3. Teardown: Run 'dbsnap revert' to cleanly remove the scheduled task.

DISCLAIMER:
Using the 'Purge' feature is destructive. It will run 'deleteMany({})' on your 
selected collections after a successful backup. Ensure your user roles and 
retention policies are properly configured.
`, asciiBanner),
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
