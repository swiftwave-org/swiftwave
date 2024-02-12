package cmd

import (
	_ "embed"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

//go:embed swiftwave-service-tls-renew.service
var swiftwaveServiceTLSRenewService string

//go:embed swiftwave-service-tls-renew.timer
var swiftwaveServiceTLSRenewTimer string

func init() {
	autoServiceTLSRenewCmd.AddCommand(enableServiceTLSRenewServiceCmd)
	autoServiceTLSRenewCmd.AddCommand(disableServiceTLSRenewServiceCmd)
}

var autoServiceTLSRenewCmd = &cobra.Command{
	Use:   "auto-renew",
	Short: "Auto renew swiftwave service TLS certificates going to expire in 30 days",
	Long:  `Auto update swiftwave service TLS certificates going to expire in 30 days`,
	Run: func(cmd *cobra.Command, args []string) {
		// print help
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

var enableServiceTLSRenewServiceCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable auto renew service",
	Long:  `Enable auto renew service`,
	Run: func(cmd *cobra.Command, args []string) {
		// Move swiftwave-service-tls-renew.service to /etc/systemd/system/
		err := os.WriteFile("/etc/systemd/system/swiftwave-service-tls-renew.service", []byte(swiftwaveServiceTLSRenewService), 0644)
		if err != nil {
			printError("Failed to write swiftwave-service-tls-renew.service file")
			return
		}
		// Move swiftwave-service-tls-renew.timer to /etc/systemd/system/
		err = os.WriteFile("/etc/systemd/system/swiftwave-service-tls-renew.timer", []byte(swiftwaveServiceTLSRenewTimer), 0644)
		if err != nil {
			printError("Failed to write swiftwave-service-tls-renew.timer file")
			return
		}
		// Reload systemd daemon
		runCommand := exec.Command("systemctl", "daemon-reload")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to reload systemd daemon")
		} else {
			printSuccess("Reloaded systemd daemon")
		}
		// Enable swiftwave service tls renew timer
		runCommand = exec.Command("systemctl", "enable", "swiftwave-service-tls-renew.timer")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to enable swiftwave service tls renew service")
		} else {
			printSuccess("Enabled swiftwave service tls renew service")
		}
		// Start swiftwave service
		runCommand = exec.Command("systemctl", "start", "swiftwave-service-tls-renew.timer")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to start swiftwave service tls renew service")
		} else {
			printSuccess("Started swiftwave service tls renew service")
		}
	},
}

var disableServiceTLSRenewServiceCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable auto renew service",
	Long:  `Disable auto renew service`,
	Run: func(cmd *cobra.Command, args []string) {
		// Stop swiftwave service
		runCommand := exec.Command("systemctl", "stop", "swiftwave-service-tls-renew.timer")
		err := runCommand.Run()
		if err != nil {
			printError("Failed to stop swiftwave auto service tls renew service")
		} else {
			printSuccess("Stopped swiftwave auto service tls renew service")
		}
		// Disable swiftwave service
		runCommand = exec.Command("systemctl", "disable", "swiftwave-service-tls-renew.timer")
		err = runCommand.Run()
		if err != nil {
			printError("Failed to disable swiftwave auto service tls renew service")
		} else {
			printSuccess("Disabled swiftwave auto service tls renew service")
		}
		// Remove swiftwave-service-tls-renew.service from /etc/systemd/system/
		err = os.Remove("/etc/systemd/system/swiftwave-service-tls-renew.service")
		if err != nil {
			printError("Failed to remove swiftwave-service-tls-renew.service file")
		}
		// Remove swiftwave-service-tls-renew.timer from /etc/systemd/system/
		err = os.Remove("/etc/systemd/system/swiftwave-service-tls-renew.timer")
		if err != nil {
			printError("Failed to remove swiftwave-service-tls-renew.timer file")
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
