package cmd

import (
	"path/filepath"

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
			printSuccess("TLS cache directory [" + dir + "] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create TLS cache directory [" + dir + "]")
			} else {
				printSuccess("Created TLS cache directory [" + dir + "]")
			}
		}
		// Create lets_encrypt.account_private_key_path base directory if it doesn't exist
		dir = systemConfig.LetsEncryptConfig.AccountPrivateKeyPath
		dir = filepath.Dir(dir)
		if checkIfFolderExists(dir) {
			printSuccess("LetsEncrypt account private key directory [" + dir + "] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create LetsEncrypt account private key directory [" + dir + "]")
			} else {
				printSuccess("Created LetsEncrypt account private key directory [" + dir + "]")
			}
		}
		// Create haproxy.unix_socket_path base directory if it doesn't exist
		dir = systemConfig.HAProxyConfig.UnixSocketPath
		dir = filepath.Dir(dir)
		if checkIfFolderExists(dir) {
			printSuccess("HAProxy unix socket directory [" + dir + "] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create HAProxy unix socket directory [" + dir + "]")
			} else {
				printSuccess("Created HAProxy unix socket directory [" + dir + "]")
			}
		}
		// Create a blank file for haproxy.unix_socket_path if it doesn't exist
		file := systemConfig.HAProxyConfig.UnixSocketPath
		if checkIfFileExists(file) {
			printSuccess("HAProxy unix socket file [" + file + "] already exists")
		} else {
			err := createFile(file)
			if err != nil {
				printError("Failed to create HAProxy unix socket file [" + file + "]")
			} else {
				printSuccess("Created HAProxy unix socket file [" + file + "]")
			}
		}
		// Create service.data_dir if it doesn't exist
		dir = systemConfig.ServiceConfig.DataDir
		if checkIfFolderExists(dir) {
			printSuccess("Data directory [" + dir + "] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create data directory [" + dir + "]")
			} else {
				printSuccess("Created data directory [" + dir + "]")
			}
		}
		// Create haproxy.data_dir if it doesn't exist
		dir = systemConfig.HAProxyConfig.DataDir
		if checkIfFolderExists(dir) {
			printSuccess("HAProxy data directory [" + dir + "] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create HAProxy data directory [" + dir + "]")
			} else {
				printSuccess("Created HAProxy data directory [" + dir + "]")
			}
		}
		// Create haproxy.data_dir/ssl if it doesn't exist
		dir = systemConfig.HAProxyConfig.DataDir + "/ssl"
		if checkIfFolderExists(dir) {
			printSuccess("HAProxy SSL directory [" + dir + "] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create HAProxy SSL directory [" + dir + "]")
			} else {
				printSuccess("Created HAProxy SSL directory [" + dir + "]")
			}
		}
		// Create the swarm network if it doesn't exist
		dockerManager, err := containermanger.NewDockerManager(systemConfig.ServiceConfig.DockerUnixSocketPath)
		if err != nil {
			printError("Failed to connect to docker daemon")
		} else {
			status := dockerManager.ExistsNetwork(systemConfig.ServiceConfig.NetworkName)
			if status {
				printSuccess("Docker network [" + systemConfig.ServiceConfig.NetworkName + "] already exists")
			} else {
				err := dockerManager.CreateNetwork(systemConfig.ServiceConfig.NetworkName)
				if err != nil {
					printError("Failed to create docker network [" + systemConfig.ServiceConfig.NetworkName + "]")
				} else {
					printSuccess("Created docker network [" + systemConfig.ServiceConfig.NetworkName + "]")
				}
			}
		}
	},
}
