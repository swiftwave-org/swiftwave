package cmd

import (
	_ "embed"
	"fmt"
	swiftwave_config "github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/system_config/bootstrap"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
	"os"

	"github.com/spf13/cobra"
)

var config *swiftwave_config.Config

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(dbMigrateCmd)
	rootCmd.AddCommand(tlsCmd)
	rootCmd.AddCommand(userManagementCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(serviceCmd)
	rootCmd.AddCommand(postgresCmd)
	rootCmd.AddCommand(localRegistryCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(autoUpdaterCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(taskQueueCmd)
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
	if (len(os.Args) > 1 && (os.Args[1] == "init" || os.Args[1] == "completion" || os.Args[1] == "--help" || os.Args[1] == "help" || os.Args[1] == "snapshot")) || len(os.Args) == 1 {
		// if first argument is "init" or no arguments, do not load config
	} else if len(os.Args) >= 1 && (os.Args[1] == "postgres" || os.Args[1] == "db-migrate" || os.Args[1] == "config" || os.Args[1] == "auto-updater" || os.Args[1] == "update" || os.Args[1] == "service") {
		// load only local config
		c, err := local_config.Fetch()
		if err != nil {
			printError("Failed to load config: " + err.Error())
			os.Exit(1)
		}
		config = &swiftwave_config.Config{
			LocalConfig:  c,
			SystemConfig: nil,
		}
		if os.Args[1] == "db-migrate" {
			autorunDBIfRequired()
		}
	} else {
		// load only local config
		lc, err := local_config.Fetch()
		if err != nil {
			printError("Failed to load local config: " + err.Error())
			os.Exit(1)
		}
		config = &swiftwave_config.Config{
			LocalConfig:  lc,
			SystemConfig: nil,
		}
		autorunDBIfRequired()
		// Migrate Database
		dbClient, err := db.GetClient(lc, 5)
		if err != nil {
			printError("Failed to connect to database: " + err.Error())
			os.Exit(1)
		}
		err = core.MigrateDatabase(dbClient)
		if err != nil {
			printError("Failed to migrate database: " + err.Error())
			os.Exit(1)
		} else {
			printSuccess("Database migrated successfully")
		}

		loadSystemConfig := false

		// if it's start command, and system setup is required, don't load complete config
		if len(os.Args) > 1 && (os.Args[1] == "start" || os.Args[1] == "localregistry" || os.Args[1] == "tq") {
			setupRequired, err := bootstrap.IsSystemSetupRequired()
			if err != nil {
				printError("Failed to check if system setup is required: " + err.Error())
				os.Exit(1)
			}
			if !setupRequired {
				loadSystemConfig = true
			} else {
				if os.Args[1] == "tq" || os.Args[1] == "localregistry" {
					printError("System setup is required. Run 'swiftwave start' to setup system")
					os.Exit(1)
				}
			}
		}

		if loadSystemConfig {
			// load complete config
			c, err := swiftwave_config.Fetch()
			if err != nil {
				printError("Failed to load config: " + err.Error())
				printInfo("Run 'swiftwave init' to initialize")
				os.Exit(1)
			}
			config = c
		}
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
