package cmd

import (
	_ "embed"
	"fmt"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"os"

	"github.com/spf13/cobra"
)

var localConfig *local_config.Config

//go:embed .version
var swiftwaveVersion string

func init() {
	rootCmd.PersistentFlags().Bool("dev", false, "Run in development mode")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(createUserCmd)
	rootCmd.AddCommand(deleteUserCmd)
	rootCmd.AddCommand(tlsCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(postgresCmd)
	rootCmd.AddCommand(dbMigrateCmd)
	rootCmd.AddCommand(applyPatchesCmd)
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

func Execute(config *local_config.Config) {
	localConfig = config
	// set config and manager
	cobra.EnableCommandSorting = false
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
