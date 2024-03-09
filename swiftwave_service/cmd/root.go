package cmd

import (
	_ "embed"
	"fmt"
	swiftwave_config "github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"os"

	"github.com/spf13/cobra"
)

var config *swiftwave_config.Config

//go:embed .version
var swiftwaveVersion string

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(createUserCmd)
	rootCmd.AddCommand(deleteUserCmd)
	rootCmd.AddCommand(tlsCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(postgresCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(autoUpdaterCmd)
}

var rootCmd = &cobra.Command{
	Use:   "swiftwave",
	Short: "SwiftWave is a self-hosted lightweight PaaS solution",
	Long:  `SwiftWave is a self-hosted lightweight PaaS solution to deploy and manage your applications on any VPS without any hassle of managing servers.`,
	Run: func(cmd *cobra.Command, args []string) {
		// print help
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

func Execute() {
	// set config and manager
	cobra.EnableCommandSorting = false
	// Check whether first argument is "install" or no arguments
	if (len(os.Args) > 1 && (os.Args[1] == "init" || os.Args[1] == "completion" || os.Args[1] == "--help")) || len(os.Args) == 1 {
		// if first argument is "init" or no arguments, do not load config
	} else {
		// load config
		c, err := swiftwave_config.Fetch()
		if err != nil {
			printError("Failed to load config: " + err.Error())
			os.Exit(1)
		}
		config = c
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}