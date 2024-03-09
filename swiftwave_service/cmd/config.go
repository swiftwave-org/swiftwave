package cmd

import (
	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config/local_config"
	"os"
)

func init() {
	configCmd.Flags().StringP("editor", "e", "", "Editor to use (vi, vim, nano, gedit, etc.)")
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Open SwiftWave configuration file in editor",
	Run: func(cmd *cobra.Command, args []string) {
		p := local_config.LocalConfigPath
		if checkIfFileExists(p) {
			if editor, _ := cmd.Flags().GetString("editor"); editor != "" {
				// set env variable
				err := os.Setenv("EDITOR", editor)
				if err != nil {
					printError("Failed to set EDITOR environment variable")
				}
			}
			openFileInEditor(p)
		} else {
			printError("Config file not found at " + p)
			printInfo("Run `swiftwave init` to create a new config file")
		}
	},
}
