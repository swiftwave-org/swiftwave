package cmd

import (
	"errors"
	"fmt"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/system_config"
	"io"
	"net/http"
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
		dockerImage := systemConfig.HAProxyConfig.DockerImage
		if !systemConfig.ServiceConfig.UseTLS {
			dockerImage = dockerImage + "-http"
		}
		// base directory for socket file
		unixSocketMountDir := filepath.Dir(systemConfig.HAProxyConfig.UnixSocketPath)
		err := generateDefaultHAProxyConfiguration(systemConfig)
		if err != nil {
			printError("Failed to generate default HAProxy configuration")
			printError("Error : " + err.Error())
			return
		}
		// Fetch hostname
		hostname, err := os.Hostname()
		if err != nil {
			printError("failed to fetch hostname")
			return
		}
		// Find out required ports
		ports := []uint{80, 443}
		dbClient, err := getDBClient()
		if err == nil {
			var ingressRules []core.IngressRule
			tx := dbClient.Select("port").Where("port IS NOT NULL").Where("protocol != udp").Find(&ingressRules)
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
			"--name", systemConfig.HAProxyConfig.ServiceName,
			"--mode", "replicated",
			"--replicas", "1",
			"--network", systemConfig.ServiceConfig.NetworkName,
			"--constraint", "node.hostname==" + hostname,
			"--mount", "type=bind,source=" + systemConfig.HAProxyConfig.DataDir + ",destination=/etc/haproxy",
			"--mount", "type=bind,source=" + unixSocketMountDir + ",destination=/home/",
		}
		args2 := make([]string, 0, len(ports))
		for _, port := range ports {
			args2 = append(args2, "--publish", "mode=host,protocol=tcp,target="+strconv.Itoa(int(port))+",published="+strconv.Itoa(int(port)))
		}
		args3 := []string{
			"--env", "ADMIN_USER=" + systemConfig.HAProxyConfig.User,
			"--env", "ADMIN_PASSWORD=" + systemConfig.HAProxyConfig.Password,
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

func generateDefaultHAProxyConfiguration(config *system_config.Config) error {
	// Check if the directory exists
	if _, err := os.Stat(config.HAProxyConfig.DataDir); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Directory does not exist > %s", config.HAProxyConfig.DataDir))
	}
	baseUrl, err := generateHAProxyConfigDownloadBaseUrl(config)
	if err != nil {
		return err
	}

	// Check if `haproxy.cfg` file exists
	if !checkIfFileExists(config.HAProxyConfig.DataDir + "/haproxy.cfg") {
		content, err := downloadContent(baseUrl + "/haproxy.cfg")
		if err != nil {
			return err
		} else {
			printSuccess("Downloaded `haproxy.cfg` file")
			err = writeContent(config.HAProxyConfig.DataDir+"/haproxy.cfg", content)
			if err != nil {
				return err
			} else {
				printSuccess("Created `haproxy.cfg` file")
			}
		}
	}

	// Check if `dataplaneapi.yaml` file exists
	if !checkIfFileExists(config.HAProxyConfig.DataDir + "/dataplaneapi.yaml") {
		content, err := downloadContent(baseUrl + "/dataplaneapi.yaml")
		if err != nil {
			return err
		} else {
			printSuccess("Downloaded `dataplaneapi.yaml` file")
			content = strings.ReplaceAll(content, "ADMIN_USERNAME", config.HAProxyConfig.User)
			content = strings.ReplaceAll(content, "ADMIN_PASSWORD", config.HAProxyConfig.Password)
			err = writeContent(config.HAProxyConfig.DataDir+"/dataplaneapi.yaml", content)
			if err != nil {
				return err
			} else {
				printSuccess("Created `dataplaneapi.yaml` file")
			}
		}
	}

	// Create `ssl` directory if it does not exist
	if _, err := os.Stat(config.HAProxyConfig.DataDir + "/ssl"); os.IsNotExist(err) {
		err := os.MkdirAll(config.HAProxyConfig.DataDir+"/ssl", os.ModePerm)
		if err != nil {
			return err
		} else {
			printSuccess("Created `ssl` directory")
		}
	}

	// Check if `ssl/default.pem` file exists
	if !checkIfFileExists(config.HAProxyConfig.DataDir + "/ssl/default.pem") {
		content, err := downloadContent(baseUrl + "/default.pem")
		if err != nil {
			return err
		} else {
			printSuccess("Downloaded `ssl/default.pem` file")
			err = writeContent(config.HAProxyConfig.DataDir+"/ssl/default.pem", content)
			if err != nil {
				return err
			} else {
				printSuccess("Created `ssl/default.pem` file")
			}
		}
	}
	return nil
}

func generateHAProxyConfigDownloadBaseUrl(config *system_config.Config) (string, error) {
	if config == nil {
		return "", errors.New("config is nil")
	}
	splitString := strings.Split(config.HAProxyConfig.DockerImage, ":")
	if len(splitString) < 2 {
		return "", errors.New("invalid docker image name")
	}
	version := splitString[1]
	url := "https://raw.githubusercontent.com/swiftwave-org/haproxy/main/" + version
	return url, nil
}

func downloadContent(url string) (string, error) {
	// download with GET request
	res, err := http.Get(url)
	if err != nil {
		return "", errors.New("failed to download file > " + url)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			printError("Failed to close response body")
		}
	}(res.Body)

	var fileStringBytes []byte
	// Read the body into bytes
	fileStringBytes, err = io.ReadAll(res.Body)
	if err != nil {
		return "", errors.New("failed to read response body")
	}

	// Convert bytes to string
	fileString := string(fileStringBytes)

	return fileString, nil
}

func writeContent(filePath string, content string) error {
	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	// Write the content
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	// Close the file
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}
