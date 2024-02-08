package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func init() {
	udpProxyCmd.AddCommand(udpProxyStatusCmd)
	udpProxyCmd.AddCommand(udpProxyStartCmd)
	udpProxyCmd.AddCommand(udpProxyStopCmd)
}

var udpProxyCmd = &cobra.Command{
	Use:   "udpproxy",
	Short: "Manage UDP Proxy service",
	Long:  "Manage UDP Proxy service",
}

// Start command
var udpProxyStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start UDP Proxy service",
	Long:  "Start UDP Proxy service",
	Run: func(cmd *cobra.Command, args []string) {
		// Delete socket file if it already exists
		if checkIfFileExists(systemConfig.UDPProxyConfig.UnixSocketPath) {
			err := os.Remove(systemConfig.UDPProxyConfig.UnixSocketPath)
			if err != nil {
				printError("Failed to remove socket file > " + systemConfig.UDPProxyConfig.UnixSocketPath)
				return
			}
		}
		dockerImage := systemConfig.UDPProxyConfig.DockerImage
		// base directory for socket file
		unixSocketMountDir := filepath.Dir(systemConfig.UDPProxyConfig.UnixSocketPath)
		// Fetch hostname
		hostname, err := os.Hostname()
		if err != nil {
			printError("failed to fetch hostname")
			return
		}
		// Find out required ports
		ports := []uint{}
		dbClient, err := getDBClient()
		if err == nil {
			var ingressRules []core.IngressRule
			tx := dbClient.Select("port").Where("port IS NOT NULL").Where("protocol == udp").Find(&ingressRules)
			if tx.Error == nil {
				if ingressRules != nil {
					for _, ingressRule := range ingressRules {
						ports = append(ports, ingressRule.Port)
					}
				}
			}
		}
		// Start HAProxy service
		args1 := []string{
			"service", "create",
			"--name", systemConfig.UDPProxyConfig.ServiceName,
			"--mode", "replicated",
			"--replicas", "1",
			"--network", systemConfig.ServiceConfig.NetworkName,
			"--constraint", "node.hostname==" + hostname,
			"--mount", "type=bind,source=" + unixSocketMountDir + ",destination=/etc/udpproxy",
			"--mount", "type=bind,source=" + systemConfig.UDPProxyConfig.DataDir + ",destination=/var/lib/udpproxy",
		}
		args2 := make([]string, 0, len(ports))
		for _, port := range ports {
			args2 = append(args2, "--publish", "mode=ingress,target="+strconv.Itoa(int(port))+",published="+strconv.Itoa(int(port)))
		}
		args3 := []string{
			"--env", "SWIFTWAVE_SERVICE_ENDPOINT=" + systemConfig.ServiceConfig.AddressOfCurrentNode + ":" + strconv.Itoa(systemConfig.ServiceConfig.BindPort),
			dockerImage,
		}
		finalArgs := append(append(args1, args2...), args3...)
		dockerCmd := exec.Command("docker", finalArgs...)
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
		dockerCmd.Stdin = os.Stdin
		err = dockerCmd.Run()
		if err != nil {
			printError("Failed to start UDP Proxy service")
			return
		}
		printSuccess("Started UDP Proxy service")
	},
}

// Stop command
var udpProxyStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop UDP Proxy service",
	Long:  "Stop UDP Proxy service",
	Run: func(cmd *cobra.Command, args []string) {
		// Stop HAProxy service
		dockerCmd := exec.Command("docker", "service", "rm", systemConfig.UDPProxyConfig.ServiceName)
		err := dockerCmd.Run()
		if err != nil {
			printError("Failed to stop UDP Proxy service")
			return
		}
		printSuccess("Stopped UDP Proxy service")
	},
}

// Status command
var udpProxyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show UDP Proxy service status",
	Long:  "Show UDP Proxy service status",
	Run: func(cmd *cobra.Command, args []string) {
		// Show HAProxy service status
		dockerManager, err := containermanger.NewDockerManager(systemConfig.ServiceConfig.DockerUnixSocketPath)
		if err != nil {
			printError("Failed to connect to docker daemon")
			return
		}
		serviceDetails, err := dockerManager.GetService(systemConfig.UDPProxyConfig.ServiceName)
		if err != nil {
			printError("UDP Proxy service is not running")
			return
		}
		// Check realtime status of HAProxy service
		info, err := dockerManager.RealtimeInfoService(systemConfig.UDPProxyConfig.ServiceName, false)
		if err != nil {
			printError("Failed to get realtime info of UDP Proxy service")
			return
		}
		// Print service status
		printSuccess("UDP Proxy service is running")
		printInfo("Service : " + systemConfig.UDPProxyConfig.ServiceName)
		printInfo("Image : " + removeHashFromDockerImageName(serviceDetails.Image))
		printInfo("Running replicas : " + strconv.Itoa(info.RunningReplicas))
		color.Green("\n--------------Node Names-------------")
		for _, placementInfo := range info.PlacementInfos {
			printInfo(placementInfo.NodeName + " (" + placementInfo.NodeID + ")")
		}
		color.Green("------------------------------------")
	},
}
