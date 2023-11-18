package cmd

import "github.com/spf13/cobra"

/*
Create the HAProxy services across all the manager nodes
---
Commands to run:
swiftwave haproxy up
swiftwave haproxy down
swiftwave haproxy status
---
*/

func init() {
	haproxyCmd.AddCommand(haproxyUpCmd)
	haproxyCmd.AddCommand(haproxyDownCmd)
	haproxyCmd.AddCommand(haproxyStatusCmd)
	rootCmd.AddCommand(haproxyCmd)
}

var haproxyCmd = &cobra.Command{
	Use:   "haproxy",
	Short: "Manage HAProxy service",
	Long:  "Manage HAProxy service",
}


// Up command
var haproxyUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Create HAProxy service",
	Long:  "Create HAProxy service",
	Run: func(cmd *cobra.Command, args []string) {
		// Create HAProxy service
	},
}

// Down command
var haproxyDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Remove HAProxy service",
	Long:  "Remove HAProxy service",
	Run: func(cmd *cobra.Command, args []string) {
		// Remove HAProxy service
	},
}

// Status command
var haproxyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show HAProxy service status",
	Long:  "Show HAProxy service status",
	Run: func(cmd *cobra.Command, args []string) {
		// Show HAProxy service status
	},
}

// Private function to check if haproxy service is created
// TODO: