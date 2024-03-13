package cmd

import (
	"github.com/spf13/cobra"
	swiftwave "github.com/swiftwave-org/swiftwave/swiftwave_service"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
)

var dbMigrateCmd = &cobra.Command{
	Use:   "db-migrate",
	Short: "Migrate the database",
	Long:  `Migrate the database`,
	Run: func(cmd *cobra.Command, args []string) {
		autorunDBIfRequired()
		// Initiating database client
		client, err := db.GetClient(config.LocalConfig, 1)
		if err != nil {
			printError("Failed to create database client")
		}
		// Migrate the database
		err = swiftwave.MigrateDatabase(client)
		if err != nil {
			printError("Failed to migrate the database")
		} else {
			printSuccess("Successfully migrated the database")
		}
	},
}
