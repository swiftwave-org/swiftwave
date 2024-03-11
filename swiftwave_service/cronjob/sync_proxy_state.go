package cronjob

import (
	"context"
	"errors"
	"fmt"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/config"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func (m Manager) SyncProxy() {
	isFirstTime := true
	for {
		if isFirstTime {
			time.Sleep(5 * time.Minute)
			isFirstTime = false
		} else {
			time.Sleep(20 * time.Minute)
		}
		m.syncProxy()
		time.Sleep(20 * time.Minute)
	}
}

func (m Manager) syncProxy() {
	// create context
	ctx := context.Background()
	// try to generate default haproxy configuration
	err := generateDefaultHAProxyConfiguration(m.Config)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to generate default haproxy configuration", err.Error())
		return
	}
	// fetch all proxy servers hostnames
	proxyServers, err := core.FetchAllProxyServers(&m.ServiceManager.DbClient)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch all proxy servers", err.Error())
		return
	}
	if len(proxyServers) == 0 {
		return
	}
	// prepare placement constraints
	var placementConstraints []string
	for _, proxyServer := range proxyServers {
		if !proxyServer.ProxyConfig.Enabled {
			placementConstraints = append(placementConstraints, "node.hostname!="+proxyServer.HostName)
		}
	}

	// fetch a swarm manager
	swarmManager, err := core.FetchSwarmManager(&m.ServiceManager.DbClient)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch swarm manager", err.Error())
		return
	}
	// create conn over ssh
	conn, err := ssh_toolkit.NetConnOverSSH("unix", swarmManager.DockerUnixSocketPath, 5, swarmManager.IP, 22, "root", m.Config.SystemConfig.SshPrivateKey, 20)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to create conn over ssh", err.Error())
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to close conn", err.Error())
		}
	}()
	// create docker client
	dockerClient, err := containermanger.New(ctx, conn)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to create docker client", err.Error())
		return
	}
	haProxyEnvironmentVariables := map[string]string{
		"ADMIN_USERNAME":             m.Config.SystemConfig.HAProxyConfig.Username,
		"ADMIN_PASSWORD":             m.Config.SystemConfig.HAProxyConfig.Password,
		"SWIFTWAVE_SERVICE_ENDPOINT": fmt.Sprintf("%s:%d", m.Config.LocalConfig.ServiceConfig.ManagementNodeAddress, m.Config.LocalConfig.ServiceConfig.BindPort),
	}
	// Try to fetch info about haproxy service
	haproxyService, err := dockerClient.GetService(m.Config.LocalConfig.ServiceConfig.HAProxyServiceName)
	if err != nil {
		// create haproxy service
		err = dockerClient.CreateService(containermanger.Service{
			Name:                 m.Config.LocalConfig.ServiceConfig.HAProxyServiceName,
			Image:                m.Config.SystemConfig.HAProxyConfig.Image,
			DeploymentMode:       containermanger.DeploymentModeGlobal,
			PlacementConstraints: placementConstraints,
			Env:                  haProxyEnvironmentVariables,
			Networks:             []string{m.Config.SystemConfig.NetworkName},
			VolumeBinds: []containermanger.VolumeBind{
				{
					Source: m.Config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath,
					Target: "/etc/haproxy",
				},
				{
					Source: m.Config.LocalConfig.ServiceConfig.HAProxyUnixSocketDirectory,
					Target: "/home",
				},
			},
		}, "", "")
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to create haproxy service", err.Error())
		} else {
			logger.CronJobLogger.Println("Created haproxy service")
		}
	} else {
		// check if env variables, image or placement constraints have changed
		if !isSameMap(haproxyService.Env, haProxyEnvironmentVariables) || haproxyService.Image != m.Config.SystemConfig.HAProxyConfig.Image || !isListSame(haproxyService.PlacementConstraints, placementConstraints) {
			logger.CronJobLogger.Println("Updating haproxy service")
			// update haproxy service
			haproxyService.Env = haProxyEnvironmentVariables
			haproxyService.Image = m.Config.SystemConfig.HAProxyConfig.Image
			haproxyService.PlacementConstraints = placementConstraints
			err = dockerClient.UpdateService(haproxyService)
			if err != nil {
				logger.CronJobLoggerError.Println("Failed to update haproxy service", err.Error())
			} else {
				logger.CronJobLogger.Println("Updated haproxy service")
			}
		}
	}
	udpProxyEnvironmentVariables := map[string]string{
		"SWIFTWAVE_SERVICE_ENDPOINT": fmt.Sprintf("%s:%d", m.Config.LocalConfig.ServiceConfig.ManagementNodeAddress, m.Config.LocalConfig.ServiceConfig.BindPort),
	}
	// Try to fetch info about udpproxy service
	udpproxyService, err := dockerClient.GetService(m.Config.LocalConfig.ServiceConfig.UDPProxyServiceName)
	if err != nil {
		// create udpproxy service
		err = dockerClient.CreateService(containermanger.Service{
			Name:                 m.Config.LocalConfig.ServiceConfig.UDPProxyServiceName,
			Image:                m.Config.SystemConfig.UDPProxyConfig.Image,
			DeploymentMode:       containermanger.DeploymentModeGlobal,
			PlacementConstraints: placementConstraints,
			Env:                  udpProxyEnvironmentVariables,
			Networks:             []string{m.Config.SystemConfig.NetworkName},
			VolumeBinds: []containermanger.VolumeBind{
				{
					Source: m.Config.LocalConfig.ServiceConfig.UDPProxyUnixSocketDirectory,
					Target: "/etc/udpproxy",
				},
				{
					Source: m.Config.LocalConfig.ServiceConfig.UDPProxyDataDirectoryPath,
					Target: "/var/lib/udpproxy",
				},
			},
		}, "", "")
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to create udpproxy service", err.Error())
		} else {
			logger.CronJobLogger.Println("Created udpproxy service")
		}
	} else {
		// check if env variables, image or placement constraints have changed
		if !isSameMap(udpproxyService.Env, udpProxyEnvironmentVariables) || udpproxyService.Image != m.Config.SystemConfig.UDPProxyConfig.Image || !isListSame(udpproxyService.PlacementConstraints, placementConstraints) {
			// update udpproxy service
			udpproxyService.Env = udpProxyEnvironmentVariables
			udpproxyService.Image = m.Config.SystemConfig.UDPProxyConfig.Image
			udpproxyService.PlacementConstraints = placementConstraints
			err = dockerClient.UpdateService(udpproxyService)
			if err != nil {
				logger.CronJobLoggerError.Println("Failed to update udpproxy service", err.Error())
			} else {
				logger.CronJobLogger.Println("Updated udpproxy service")
			}
		}
	}

}

// private function
func isListSame(list1 []string, list2 []string) bool {
	// order does not matter
	if len(list1) != len(list2) {
		return false
	}
	for _, item1 := range list1 {
		found := false
		for _, item2 := range list2 {
			if item1 == item2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func isSameMap(map1 map[string]string, map2 map[string]string) bool {
	if len(map1) != len(map2) {
		return false
	}
	for key, value1 := range map1 {
		value2, ok := map2[key]
		if !ok {
			return false
		}
		if value1 != value2 {
			return false
		}
	}
	return true
}

// generateDefaultHAProxyConfiguration : Generate default haproxy configuration (skip if already exists)
func generateDefaultHAProxyConfiguration(config *config.Config) error {
	// Check if the directory exists
	if _, err := os.Stat(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist > %s", config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath)
	}
	baseUrl, err := generateHAProxyConfigDownloadBaseUrl(config)
	if err != nil {
		return err
	}

	// Check if `haproxy.cfg` file exists
	if !checkIfFileExists(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath + "/haproxy.cfg") {
		content, err := downloadContent(baseUrl + "/haproxy.cfg")
		if err != nil {
			return err
		} else {
			log.Println("Downloaded `haproxy.cfg` file")
			err = writeContent(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath+"/haproxy.cfg", content)
			if err != nil {
				return err
			} else {
				log.Println("Created `haproxy.cfg` file")
			}
		}
	}

	// Check if `dataplaneapi.yaml` file exists
	if !checkIfFileExists(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath + "/dataplaneapi.yaml") {
		content, err := downloadContent(baseUrl + "/dataplaneapi.yaml")
		if err != nil {
			return err
		} else {
			log.Println("Downloaded `dataplaneapi.yaml` file")
			content = strings.ReplaceAll(content, "ADMIN_USERNAME", config.SystemConfig.HAProxyConfig.Username)
			content = strings.ReplaceAll(content, "ADMIN_PASSWORD", config.SystemConfig.HAProxyConfig.Password)
			err = writeContent(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath+"/dataplaneapi.yaml", content)
			if err != nil {
				return err
			} else {
				log.Println("Created `dataplaneapi.yaml` file")
			}
		}
	}

	// Create `ssl` directory if it does not exist
	if _, err := os.Stat(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath + "/ssl"); os.IsNotExist(err) {
		err := os.MkdirAll(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath+"/ssl", os.ModePerm)
		if err != nil {
			return err
		} else {
			log.Println("Created `ssl` directory")
		}
	}

	// Check if `ssl/default.pem` file exists
	if !checkIfFileExists(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath + "/ssl/default.pem") {
		content, err := downloadContent(baseUrl + "/default.pem")
		if err != nil {
			return err
		} else {
			log.Println("Downloaded `ssl/default.pem` file")
			err = writeContent(config.LocalConfig.ServiceConfig.HAProxyDataDirectoryPath+"/ssl/default.pem", content)
			if err != nil {
				return err
			} else {
				log.Println("Created `ssl/default.pem` file")
			}
		}
	}
	return nil
}

func generateHAProxyConfigDownloadBaseUrl(config *config.Config) (string, error) {
	if config == nil {
		return "", errors.New("config is nil")
	}
	splitString := strings.Split(config.SystemConfig.HAProxyConfig.Image, ":")
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
			log.Println("Failed to close response body")
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

func checkIfFileExists(file string) bool {
	cmd := exec.Command("ls", file)
	err := cmd.Run()
	return err == nil
}
