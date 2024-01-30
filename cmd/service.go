package cmd

import (
	_ "embed"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

//go:embed swiftwave.service
var swiftwaveService string

func init() {
	serviceCmd.AddCommand(enableServiceCmd)
	serviceCmd.AddCommand(disableServiceCmd)
	serviceCmd.AddCommand(restartServiceCmd)
	serviceCmd.AddCommand(statusServiceCmd)
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage Swiftwave Daemon Service",
	Long:  `Manage Swiftwave Daemon Service`,
	Run: func(cmd *cobra.Command, args []string) {
		// print help
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

var enableServiceCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a service",
	Long:  `Enable a service`,
	Run: func(cmd *cobra.Command, args []string) {
		// Move swiftwave.service to /etc/systemd/system/
		err := os.WriteFile("/etc/systemd/system/swiftwave.service", []byte(swiftwaveService), 0644)
		if err != nil {
			printError("Failed to write swiftwave.service file")
		}
		// Reload systemd daemon
		runCommand := exec.Command("systemctl", "daemon-reload")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to reload systemd daemon")
		} else {
			printSuccess("Reloaded systemd daemon")
		}
		// Enable swiftwave service
		runCommand = exec.Command("systemctl", "enable", "swiftwave.service")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to enable swiftwave service")
		} else {
			printSuccess("Enabled swiftwave service")
		}
		// Start swiftwave service
		runCommand = exec.Command("systemctl", "start", "swiftwave.service")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to start swiftwave service")
		} else {
			printSuccess("Started swiftwave service")
		}
	},
}

var disableServiceCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable a service",
	Long:  `Disable a service`,
	Run: func(cmd *cobra.Command, args []string) {
		// Stop swiftwave service
		runCommand := exec.Command("systemctl", "stop", "swiftwave.service")
		err := runCommand.Run()
		if err != nil {
			printError("Failed to stop swiftwave service")
		} else {
			printSuccess("Stopped swiftwave service")
		}
		// Disable swiftwave service
		runCommand = exec.Command("systemctl", "disable", "swiftwave.service")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to disable swiftwave service")
		} else {
			printSuccess("Disabled swiftwave service")
		}
		// Remove swiftwave.service from /etc/systemd/system/
		err = os.Remove("/etc/systemd/system/swiftwave.service")
		if err != nil {
			printError("Failed to remove swiftwave.service file")
		}
		// Reload systemd daemon
		runCommand = exec.Command("systemctl", "daemon-reload")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to reload systemd daemon")
		} else {
			printSuccess("Reloaded systemd daemon")
		}
	},
}

var restartServiceCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart a service",
	Long:  `Restart a service`,
	Run: func(cmd *cobra.Command, args []string) {
		// Restart swiftwave service
		runCommand := exec.Command("systemctl", "restart", "swiftwave.service")
		err := runCommand.Run()
		if err != nil {
			printError("Failed to restart swiftwave service")
		} else {
			printSuccess("Restarted swiftwave service")
		}
	},
}

var statusServiceCmd = &cobra.Command{
	Use:   "status",
	Short: "Get status of swiftwave service",
	Long:  `Get status of swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get status of swiftwave service
		runCommand := exec.Command("systemctl", "status", "swiftwave.service")
		runCommand.Stdout = os.Stdout
		runCommand.Stderr = os.Stderr
		err := runCommand.Run()
		if err != nil {
			printError("Failed to get status of swiftwave service")
		} else {
			printSuccess("Got status of swiftwave service")
		}
	},
}
