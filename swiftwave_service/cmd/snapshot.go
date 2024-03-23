package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"time"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Take a snapshot of the current state of the system",
	Run: func(cmd *cobra.Command, args []string) {
		currentTime := time.Now()
		formattedTime := currentTime.Format("02-01-2006_15_04")
		filename := fmt.Sprintf("swiftwave_snapshot_%s.tar.gz", formattedTime)
		c := exec.Command("tar", "-czvf", filename, "-C", "/var/lib/swiftwave", ".")
		err := c.Run()
		if err != nil {
			printError("Error taking snapshot: " + err.Error())
			os.Exit(1)
		} else {
			// mark as 777
			err := os.Chmod(filename, 0777)
			if err != nil {
				printError("Error setting permission on snapshot: " + err.Error())
				os.Exit(1)
			}
			printSuccess("Snapshot saved as " + filename)
			printWarning("Please keep this file safe, it contains sensitive information")
			os.Exit(0)
		}
	},
}
