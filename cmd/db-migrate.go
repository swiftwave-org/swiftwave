package cmd

import (
	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbMigrateCmd = &cobra.Command{
	Use:   "db-migrate",
	Short: "Migrate the database",
	Long:  `Migrate the database`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initiating database client
		dbDialect := postgres.Open(systemConfig.PostgresqlConfig.DSN())
		client, err := gorm.Open(dbDialect, &gorm.Config{})
		if err != nil {
			printError("Failed to create database client")
		}
		// Migrate the database
		err = core.MigrateDatabase(client)
		if err != nil {
			printError("Failed to migrate the database")
		} else {
			printSuccess("Successfully migrated the database")
		}
	},
}
