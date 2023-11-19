package cmd

import "github.com/spf13/cobra"

/*
Create the HAProxy services across all the manager nodes
---
Commands to run:
swiftwave haproxy start
swiftwave haproxy stop
swiftwave haproxy status
---
*/

func init() {
	haproxyCmd.AddCommand(haproxyStartCmd)
	haproxyCmd.AddCommand(haproxyStopCmd)
	haproxyCmd.AddCommand(haproxyStatusCmd)
	rootCmd.AddCommand(haproxyCmd)
}

var haproxyCmd = &cobra.Command{
	Use:   "haproxy",
	Short: "Manage HAProxy service",
	Long:  "Manage HAProxy service",
}


// Start command
var haproxyStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start HAProxy service",
	Long:  "Start HAProxy service",
	Run: func(cmd *cobra.Command, args []string) {
		// Start HAProxy service
	},
}

// Stop command
var haproxyStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop HAProxy service",
	Long:  "Stop HAProxy service",
	Run: func(cmd *cobra.Command, args []string) {
		// Stop HAProxy service
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