package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	createUserCmd.Flags().StringP("username", "u", "", "Username")
	createUserCmd.Flags().StringP("password", "p", "", "Password")
}

var createUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Create a new user",
	Long:  "Create a new user",
	Run: func(cmd *cobra.Command, args []string) {
		username := cmd.Flag("username").Value.String()
		password := cmd.Flag("password").Value.String()
		if username == "" {
			printError("Username is required")
			err := cmd.Help()
			if err != nil {
				return
			}
			return
		}
		if password == "" {
			printError("Password is required")
			err := cmd.Help()
			if err != nil {
				return
			}
			return
		}
		// Initiating database client
		dbDialect := postgres.Open(systemConfig.PostgresqlConfig.DSN())
		dbClient, err := gorm.Open(dbDialect, &gorm.Config{})
		if err != nil {
			printError("Failed to connect to database")
			return
		}
		// Create user
		user := core.User{
			Username: username,
		}
		err = user.SetPassword(password)
		if err != nil {
			printError("Failed to set password")
			return
		}
		createUser, err := core.CreateUser(context.Background(), *dbClient, user)
		if err != nil {
			printError("Failed to create user")
			return
		}
		printSuccess("Created user > username: " + createUser.Username)
	},
}
