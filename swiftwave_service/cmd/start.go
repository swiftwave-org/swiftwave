package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	swiftwaveservice "github.com/swiftwave-org/swiftwave/swiftwave_service"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start swiftwave service",
	Long:  `Start swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		if config == nil {
			return
		}
		if config.LocalConfig.IsDevelopmentMode {
			color.Yellow("Running in Development mode")
			color.Red("This can impose security risks. Turn off development mode (swiftwave config) in production environment.")
		}
		swiftwaveservice.Start(config)
	},
}
