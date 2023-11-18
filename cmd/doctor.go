package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Analyze and generate reports about the current system",
	Long:  "Analyze and generate reports about the current system",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if config file is valid
		// Check if Docker Unix socket is working
		// Check if Postgres Database is reachable
		// Check if haproxy-service is running
		// Check if haproxy socket is reachable
	},
}
