package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Prepare the environment for swiftwave",
	Long:  "Prepare the environment for swiftwave",
	Run: func(cmd *cobra.Command, args []string) {
		var dir string
		// Create service.ssl_certificate_dir if it doesn't exist
		dir = systemConfig.ServiceConfig.SSLCertificateDir
		if checkIfFolderExists(dir) {
			printSuccess("Swiftwave Service Certificate directory [" + dir + "] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create Swiftwave Service Certificate [" + dir + "]")
			} else {
				printSuccess("Created Swiftwave Service Certificate [" + dir + "]")
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
		// Create udp_proxy.unix_socket_path base directory if it doesn't exist
		dir = systemConfig.UDPProxyConfig.UnixSocketPath
		dir = filepath.Dir(dir)
		if checkIfFolderExists(dir) {
			printSuccess("UDP Proxy unix socket directory [" + dir + "] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create UDP Proxy unix socket directory [" + dir + "]")
			} else {
				printSuccess("Created UDP Proxy unix socket directory [" + dir + "]")
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
		// Create udp_proxy.data_dir if it doesn't exist
		dir = systemConfig.UDPProxyConfig.DataDir
		if checkIfFolderExists(dir) {
			printSuccess("UDP Proxy data directory [" + dir + "] already exists")
		} else {
			err := createFolder(dir)
			if err != nil {
				printError("Failed to create UDP Proxy data directory [" + dir + "]")
			} else {
				printSuccess("Created UDP Proxy data directory [" + dir + "]")
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
