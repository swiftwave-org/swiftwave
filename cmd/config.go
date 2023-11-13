package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().StringP("editor", "e", "", "Editor to use (vi, vim, nano, gedit, etc.)")
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Open SwiftWave configuration file in editor",
	Run: func(cmd *cobra.Command, args []string) {
		if checkIfFileExists(configFilePath) {
			if editor, _ := cmd.Flags().GetString("editor"); editor != "" {
				// set env variable
				err := os.Setenv("EDITOR", editor)
				if err != nil {
					printError("Failed to set EDITOR environment variable")
				}
			}
			openFileInEditor(configFilePath)
		} else {
			printError("Config file not found at " + configFilePath)
			printInfo("Run `swiftwave init` to create a new config file")
		}
	},
}
