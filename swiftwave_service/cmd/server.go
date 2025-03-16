package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/db"
)

func init() {
	serverManagementCmd.Flags().SortFlags = false
	serverManagementCmd.AddCommand(createServerCmd)
	serverManagementCmd.AddCommand(deleteServerCmd)
	serverManagementCmd.AddCommand(listServerCmd)
	serverManagementCmd.AddCommand(getSetupAgentCmd)

	createServerCmd.Flags().String("name", "", "Server Name")
	createServerCmd.Flags().String("ip", "", "Server IP")
	createServerCmd.MarkFlagRequired("name")
	createServerCmd.MarkFlagRequired("ip")

	deleteServerCmd.Flags().String("name", "", "Server Name")
	deleteServerCmd.MarkFlagRequired("name")
}

var serverManagementCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage servers",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

var createServerCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new server",
	Long:  "Create a new server",
	Run: func(cmd *cobra.Command, args []string) {
		server_name := cmd.Flag("name").Value.String()
		server_ip := cmd.Flag("ip").Value.String()
		if server_name == "" || server_ip == "" {
			printError("server name and ip are required")
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
		// Create server
		server := core.Server{
			Name:     server_name,
			PublicIP: server_ip,
		}
		err = core.CreateServer(dbClient, &server)
		if err != nil {
			printError("Failed to create server")
			return
		}
		printSuccess("Created server > " + server_name)
	},
}

var deleteServerCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a server",
	Long:  "Delete a server",
	Run: func(cmd *cobra.Command, args []string) {
		server_name := cmd.Flag("name").Value.String()
		if server_name == "" {
			printError("server name is required")
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
		server, err := core.FetchServerByName(dbClient, server_name)
		if err != nil {
			printError("Failed to fetch server")
			return
		}
		// Delete server
		err = core.DeleteServer(dbClient, server.ID)
		if err != nil {
			printError("Failed to delete server")
			return
		}
	},
}

var listServerCmd = &cobra.Command{
	Use:   "list",
	Short: "List servers",
	Long:  "List servers",
	Run: func(cmd *cobra.Command, args []string) {
		dbClient, err := db.GetClient(config.LocalConfig, 10)
		if err != nil {
			printError("Failed to connect to database")
			return
		}
		servers, err := core.FetchAllServers(dbClient)
		if err != nil {
			printError("Failed to fetch servers")
			return
		}
		// Create a tab writer
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', tabwriter.AlignRight)
		w.Write([]byte("ID\tName\tIP\tWireguard IP\tDeployment\tProxy\tStatus\tLast Ping\n"))
		for _, server := range servers {
			w.Write([]byte(fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%v\t%s\t%s\n", server.ID, server.Name, server.PublicIP, server.WireguardConfig.IP, server.Status, server.ProxyEnabled, server.Status, server.LastPing.Format("2006-01-02 15:04:05"))))
		}
		w.Flush()
	},
}

var getSetupAgentCmd = &cobra.Command{
	Use:   "setup-agent",
	Short: "Setup agent on server",
	Long:  "Setup agent on server",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}
