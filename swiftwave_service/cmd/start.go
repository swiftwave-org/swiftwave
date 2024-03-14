package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	swiftwave "github.com/swiftwave-org/swiftwave/swiftwave_service"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config/bootstrap"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Swiftwave",
	Long:  `Start Swiftwave`,
	Run: func(cmd *cobra.Command, args []string) {
		if config == nil {
			return
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
			swiftwave.StartSwiftwave(config)
		}
	},
}
