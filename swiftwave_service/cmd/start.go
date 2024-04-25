package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	swiftwave "github.com/swiftwave-org/swiftwave/swiftwave_service"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config/bootstrap"
	"os"
	"strings"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Swiftwave",
	Long:  `Start Swiftwave`,
	Run: func(cmd *cobra.Command, args []string) {
		if config == nil {
			return
		}
		err := os.Setenv("SSH_AUTH_SOCK", config.LocalConfig.EnvironmentVariables.SshAuthSock)
		if err != nil {
			printError("Failed to set SSH_AUTH_SOCK in environment variables\n" + err.Error())
			os.Exit(1)
		}
		err = os.Setenv("SSH_KNOWN_HOSTS", config.LocalConfig.EnvironmentVariables.SshKnownHosts)
		if err != nil {
			printError("Failed to set SSH_KNOWN_HOSTS in environment variables\n" + err.Error())
			os.Exit(1)
		}

		if strings.Compare(config.LocalConfig.EnvironmentVariables.SshAuthSock, "") == 0 {
			printError("SSH_AUTH_SOCK is not available in environment variables\n")
			printInfo("Enable SSH Agent Forwarding in your SSH config")
			printInfo("Run `swiftwave config` to edit the config file and set SSH_AUTH_SOCK")
		}
		if strings.Compare(config.LocalConfig.EnvironmentVariables.SshKnownHosts, "") == 0 {
			printError("SSH_KNOWN_HOSTS is not available in environment variables\n")
			printInfo("Enable SSH Agent Forwarding in your SSH config")
			printInfo("Run `swiftwave config` to edit the config file and set SSH_KNOWN_HOSTS")
		}

		if config.LocalConfig.IsDevelopmentMode {
			color.Yellow("Running in Development mode")
			color.Red("This can impose security risks. Turn off development mode (swiftwave config) in production environment.")
		}
		// check if system setup is required
		setupRequired, err := bootstrap.IsSystemSetupRequired()
		if err != nil {
			printError("Failed to check if system setup is required")
			printError(err.Error())
			os.Exit(1)
			return
		}
		if setupRequired {
			printSuccess(fmt.Sprintf("System Setup Server started successfully.\nVisit http://%s:%d to setup the system.", config.LocalConfig.ServiceConfig.ManagementNodeAddress, config.LocalConfig.ServiceConfig.BindPort))
			err := bootstrap.StartBootstrapServer()
			if err != nil {
				printError("Failed to start bootstrap server")
				printError(err.Error())
			}
		} else {
			isRequired, err := isLocalRegistryRequired()
			if err != nil {
				printError("Failed to check if local registry is required")
				printError(err.Error())
				os.Exit(1)
				return
			}
			if isRequired {
				color.Yellow("Local registry will be used for image storage")
				isRunning, err := isLocalRegistryRunning(cmd.Context())
				if err != nil {
					printError("Failed to check if local registry is running")
					printError(err.Error())
					os.Exit(1)
					return
				}
				if !isRunning {
					color.Yellow("Starting local registry")
					err := startLocalRegistry(cmd.Context())
					if err != nil {
						printError("Failed to start local registry")
						printError(err.Error())
						os.Exit(1)
						return
					}
				}
			}
			swiftwave.StartSwiftwave(config)
		}
	},
}
