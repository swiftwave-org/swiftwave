package cmd

import (
	"github.com/spf13/cobra"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:  "setup",
	Long: "Setup the environment for the first time",
	Run: func(cmd *cobra.Command, args []string) {
		var dir string
		// Create service.tls_cache_dir if it doesn't exist
		dir = systemConfig.ServiceConfig.TLSCacheDir
		if checkIfFolderExists(dir) {
			printSuccess("TLS cache directory ["+dir+"] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create TLS cache directory ["+dir+"]")
			} else {
				printSuccess("Created TLS cache directory ["+dir+"]")
			}
		}
		// Create service.data_dir if it doesn't exist
		dir = systemConfig.ServiceConfig.DataDir
		if checkIfFolderExists(dir) {
			printSuccess("Data directory ["+dir+"] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create data directory ["+dir+"]")
			} else {
				printSuccess("Created data directory ["+dir+"]")
			}
		}
		// Create haproxy.data_dir if it doesn't exist
		dir = systemConfig.HAProxyConfig.DataDir
		if checkIfFolderExists(dir) {
			printSuccess("HAProxy data directory ["+dir+"] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create HAProxy data directory ["+dir+"]")
			} else {
				printSuccess("Created HAProxy data directory ["+dir+"]")
			}
		}
		// Create haproxy.data_dir/ssl if it doesn't exist
		dir = systemConfig.HAProxyConfig.DataDir + "/ssl"
		if checkIfFolderExists(dir) {
			printSuccess("HAProxy SSL directory ["+dir+"] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create HAProxy SSL directory ["+dir+"]")
			} else {
				printSuccess("Created HAProxy SSL directory ["+dir+"]")
			}
		}
		// Create the swarm network if it doesn't exist
		dockerManager, err := containermanger.NewDockerManager(systemConfig.ServiceConfig.DockerUnixSocketPath)
		if err != nil {
			printError("Failed to connect to docker daemon")
		} else {
			status := dockerManager.ExistsNetwork(systemConfig.ServiceConfig.NetworkName)
			if status {
				printSuccess("Docker network ["+systemConfig.ServiceConfig.NetworkName+"] already exists")
			} else {
				err := dockerManager.CreateNetwork(systemConfig.ServiceConfig.NetworkName)
				if err != nil {
					printError("Failed to create docker network ["+systemConfig.ServiceConfig.NetworkName+"]")
				} else {
					printSuccess("Created docker network ["+systemConfig.ServiceConfig.NetworkName+"]")
				}
			}
		}
	},
}

