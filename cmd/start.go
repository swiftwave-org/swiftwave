package cmd

import (
	"github.com/spf13/cobra"
	swiftwave "github.com/swiftwave-org/swiftwave/swiftwave_service"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start swiftwave service",
	Long:  `Start swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		swiftwave.Start(systemConfig)
	},
}
