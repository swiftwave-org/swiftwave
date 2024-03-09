package cmd

import (
	"github.com/spf13/cobra"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	swiftwave "github.com/swiftwave-org/swiftwave/swiftwave_service"
	"os"
	"os/exec"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start swiftwave service",
	Long:  `Start swiftwave service`,
	Run: func(cmd *cobra.Command, args []string) {
		binaryPath, err := os.Executable()
		if err != nil {
			printError("Failed to get swiftwave binary path")
			return
		}
		if !isHaproxyRunning() {
			printInfo("Starting HAProxy service")
			c := exec.Command(binaryPath, "haproxy", "start")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				printError("Failed to start HAProxy service")
				return
			}
		}
		if !isUDPProxyRunning() {
			printInfo("Starting UDPProxy service")
			c := exec.Command(binaryPath, "udpproxy", "start")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				printError("Failed to start UDPProxy service")
				return
			}
		}
		swiftwave.Start(config)
	},
}

func isHaproxyRunning() bool {
	dockerManager, err := containermanger.NewDockerManager(config.ServiceConfig.DockerUnixSocketPath)
	if err != nil {
		printError("Failed to connect to docker daemon")
		return false
	}
	_, err = dockerManager.GetService(config.HAProxyConfig.ServiceName)
	return err == nil
}

func isUDPProxyRunning() bool {
	dockerManager, err := containermanger.NewDockerManager(config.ServiceConfig.DockerUnixSocketPath)
	if err != nil {
		printError("Failed to connect to docker daemon")
		return false
	}
	_, err = dockerManager.GetService(config.UDPProxyConfig.ServiceName)
	return err == nil
}
