package cmd

import (
	_ "embed"
	"github.com/spf13/cobra"
)

func init() {
	versionCmd.Flags().BoolP("short", "s", false, "Show only the version number")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Swiftwave",
	Long:  "Print the version number of Swiftwave",
	Run: func(cmd *cobra.Command, args []string) {
		if short, _ := cmd.Flags().GetBool("short"); short {
			cmd.Println(swiftwaveVersion)
			return
		}
		cmd.Println("Swiftwave is running in version " + swiftwaveVersion)
	},
}
