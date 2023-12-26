package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
)

/*
Create the HAProxy services across all the manager nodes
---
Commands to run:
swiftwave haproxy start
swiftwave haproxy stop
swiftwave haproxy status
---
Special customized image for HAProxy: https://github.com/swiftwave-org/haproxy
---
Command to start HAProxy service:
docker service create \
	--name <from_config> \
	--mode global \
	--network <from_config> \
	--mount type=bind,source=<from_config>,destination=/var/lib/haproxy \
	--mount type=bind,source=<from_config>,destination=/home/dataplaneapi.sock \
	--publish mode=host,target=80,published=80 \
	--publish mode=host,target=443,published=443 \
	--env ADMIN_USER=<from_config> \
	--env ADMIN_PASSWORD=<from_config> \
	--env SWIFTWAVE_SERVICE_ENDPOINT=<from_config> \
<image_from_config>

*/

func init() {
	haproxyCmd.AddCommand(haproxyStatusCmd)
	haproxyCmd.AddCommand(haproxyStartCmd)
	haproxyCmd.AddCommand(haproxyStopCmd)
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
		// Delete socket file if it already exists
		if checkIfFileExists(systemConfig.HAProxyConfig.UnixSocketPath) {
			err := os.Remove(systemConfig.HAProxyConfig.UnixSocketPath)
			if err != nil {
				printError("Failed to remove socket file > " + systemConfig.HAProxyConfig.UnixSocketPath)
				return
			}
		}
		// base directory for socket file
		mountDir := filepath.Dir(systemConfig.HAProxyConfig.UnixSocketPath)
		// Service endpoint
		SERVICE_ENDPOINT := systemConfig.ServiceConfig.AddressOfCurrentNode
		if systemConfig.ServiceConfig.UseTLS {
			SERVICE_ENDPOINT = "https://" + SERVICE_ENDPOINT
		} else {
			SERVICE_ENDPOINT = "http://" + SERVICE_ENDPOINT
		}
		// add port
		SERVICE_ENDPOINT = SERVICE_ENDPOINT + ":" + strconv.Itoa(systemConfig.ServiceConfig.BindPort)
		// Start HAProxy service
		dockerCmd := exec.Command("docker", "service", "create",
			"--name", systemConfig.HAProxyConfig.ServiceName,
			"--mode", "global",
			"--network", systemConfig.ServiceConfig.NetworkName,
			"--mount", "type=bind,source="+systemConfig.HAProxyConfig.DataDir+",destination=/var/lib/haproxy",
			"--mount", "type=bind,source="+mountDir+",destination=/home/",
			"--publish", "mode=host,target=80,published=80",
			"--publish", "mode=host,target=443,published=443",
			"--env", "ADMIN_USER="+systemConfig.HAProxyConfig.User,
			"--env", "ADMIN_PASSWORD="+systemConfig.HAProxyConfig.Password,
			"--env", "SWIFTWAVE_SERVICE_ENDPOINT="+SERVICE_ENDPOINT,
			systemConfig.HAProxyConfig.DockerImage)
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
		dockerCmd.Stdin = os.Stdin
		err := dockerCmd.Run()
		if err != nil {
			printError("Failed to start HAProxy service")
			return
		}
		printSuccess("Started HAProxy service")
	},
}

// Stop command
var haproxyStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop HAProxy service",
	Long:  "Stop HAProxy service",
	Run: func(cmd *cobra.Command, args []string) {
		// Stop HAProxy service
		dockerCmd := exec.Command("docker", "service", "rm", systemConfig.HAProxyConfig.ServiceName)
		err := dockerCmd.Run()
		if err != nil {
			printError("Failed to stop HAProxy service")
			return
		}
		printSuccess("Stopped HAProxy service")
	},
}

// Status command
var haproxyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show HAProxy service status",
	Long:  "Show HAProxy service status",
	Run: func(cmd *cobra.Command, args []string) {
		// Show HAProxy service status
		dockerManager, err := containermanger.NewDockerManager(systemConfig.ServiceConfig.DockerUnixSocketPath)
		if err != nil {
			printError("Failed to connect to docker daemon")
			return
		}
		serviceDetails, err := dockerManager.GetService(systemConfig.HAProxyConfig.ServiceName)
		if err != nil {
			printError("HAProxy service is not running")
			return
		}
		// Check realtime status of HAProxy service
		info, err := dockerManager.RealtimeInfoService(systemConfig.HAProxyConfig.ServiceName, false)
		if err != nil {
			printError("Failed to get realtime info of HAProxy service")
			return
		}
		// Print HAProxy service status
		printSuccess("HAProxy service is running")
		printInfo("Service : " + systemConfig.HAProxyConfig.ServiceName)
		printInfo("Image : " + removeHashFromDockerImageName(serviceDetails.Image))
		printInfo("Running replicas : " + strconv.Itoa(info.RunningReplicas))
		color.Green("\n--------------Node Names-------------")
		for _, placementInfo := range info.PlacementInfos {
			printInfo(placementInfo.NodeName + " (" + placementInfo.NodeID + ")")
		}
		color.Green("------------------------------------")
	},
}

// Private function to check if haproxy service is created
func removeHashFromDockerImageName(image string) string {
	// split at @
	s := strings.Split(image, "@")
	if len(s) == 0 {
		// no @ found
		return image
	}
	// return the first part
	return s[0]
}
