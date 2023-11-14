package cmd

import (
	"github.com/spf13/cobra"
	"strconv"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Print info of swiftwave",
	Long:  `Print info of swiftwave`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Version : ", systemConfig.Version)
		cmd.Println("Deployed in", systemConfig.Mode, "mode")
		printInfo("Domain pointed to current server > " + systemConfig.ServiceConfig.AddressOfCurrentNode)
		printInfo("Listening on " + systemConfig.ServiceConfig.BindAddress + ":" + strconv.Itoa(systemConfig.ServiceConfig.BindPort))
		printInfo("Service accessible at https://" + systemConfig.ServiceConfig.AddressOfCurrentNode + ":" + strconv.Itoa(systemConfig.ServiceConfig.BindPort))

	},
}
