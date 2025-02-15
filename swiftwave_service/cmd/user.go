package cmd

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
	"golang.org/x/term"
	"os"
)

func init() {
	userManagementCmd.AddCommand(createUserCmd)
	userManagementCmd.AddCommand(deleteUserCmd)
	userManagementCmd.AddCommand(disableTotpCmd)
	createUserCmd.Flags().StringP("username", "u", "", "userID")
	createUserCmd.Flags().StringP("password", "p", "", "Password [Optional]")
	deleteUserCmd.Flags().StringP("username", "u", "", "userID")
	disableTotpCmd.Flags().StringP("username", "u", "", "userID")
}

var userManagementCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  `Manage users`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

var createUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	Long:  "Create a new user",
	Run: func(cmd *cobra.Command, args []string) {
		username := cmd.Flag("username").Value.String()
		password := cmd.Flag("password").Value.String()
		if username == "" {
			printError("userID is required")
			err := cmd.Help()
			if err != nil {
				return
			}
			return
		}
		if password == "" {
			// Ask for password
			fmt.Print("Enter password: ")
			enteredPassword, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				printError("Failed to read password")
				return
			}
			fmt.Println()
			fmt.Print("Confirm password: ")
			confirmPassword, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				printError("Failed to read password")
				return
			}
			fmt.Println()
			if string(enteredPassword) != string(confirmPassword) {
				printError("Passwords do not match")
				return
			}
			password = string(enteredPassword)
			if password == "" {
				printError("Password is required")
				return
			}
		} else {
			color.Yellow("Passing password as a flag is not recommended. It will be visible in the terminal history.")
			color.Yellow("Use it only for automation purposes")
		}
		// Initiating database client
		dbClient, err := db.GetClient(config.LocalConfig, 10)
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
		printSuccess("Created user > " + createUser.Username)
	},
}

var deleteUserCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	Long:  "Delete a user",
	Run: func(cmd *cobra.Command, args []string) {
		username := cmd.Flag("username").Value.String()
		if username == "" {
			printError("userID is required")
			err := cmd.Help()
			if err != nil {
				return
			}
			return
		}
		// Initiating database client
		dbClient, err := db.GetClient(config.LocalConfig, 10)

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

var disableTotpCmd = &cobra.Command{
	Use:   "disable-totp",
	Short: "Disable Totp for a user",
	Long:  "Disable Totp for a user",
	Run: func(cmd *cobra.Command, args []string) {
		username := cmd.Flag("username").Value.String()
		if username == "" {
			printError("userID is required")
			err := cmd.Help()
			if err != nil {
				return
			}
			return
		}
		// Initiating database client
		dbClient, err := db.GetClient(config.LocalConfig, 10)
		if err != nil {
			printError("Failed to connect to database")
			return
		}
		// Disable Totp
		err = core.DisableTotp(context.Background(), *dbClient, username)
		if err != nil {
			printError("Failed to disable Totp")
			printError("Reason: " + err.Error())
			return
		}
		printSuccess("Disabled Totp for user > " + username)
	},
}
