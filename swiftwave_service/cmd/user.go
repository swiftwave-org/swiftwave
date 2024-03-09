package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
)

func init() {
	createUserCmd.Flags().StringP("username", "u", "", "Username")
	createUserCmd.Flags().StringP("password", "p", "", "Password")
	deleteUserCmd.Flags().StringP("username", "u", "", "Username")
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
		dbClient, err := db.GetClient(config.LocalConfig, 1)
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

var deleteUserCmd = &cobra.Command{
	Use:   "delete-user",
	Short: "Delete a user",
	Long:  "Delete a user",
	Run: func(cmd *cobra.Command, args []string) {
		username := cmd.Flag("username").Value.String()
		if username == "" {
			printError("Username is required")
			err := cmd.Help()
			if err != nil {
				return
			}
			return
		}
		// Initiating database client
		dbClient, err := db.GetClient(config.LocalConfig, 1)

		if err != nil {
			printError("Failed to connect to database")
			return
		}
		// Fetch user
		user, err := core.FindUserByUsername(context.Background(), *dbClient, username)
		if err != nil {
			printError(fmt.Sprintf("User %s not found !", username))
			return
		}

		// Delete user
		err = core.DeleteUser(context.Background(), *dbClient, user.ID)
		if err != nil {
			printError("Failed to delete user")
			return
		}
		printSuccess("Deleted user > " + username)
	},
}
