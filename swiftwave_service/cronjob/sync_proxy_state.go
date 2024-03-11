package cronjob

import (
	"context"
	"fmt"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"time"
)

func (m Manager) SyncProxy() {
	isFirstTime := true
	for {
		if isFirstTime {
			time.Sleep(5 * time.Minute)
			isFirstTime = false
		} else {
			time.Sleep(50 * time.Minute)
		}
		m.syncProxy()
	}
}

func (m Manager) syncProxy() {
	// create context
	ctx := context.Background()
	// fetch all proxy servers hostnames
	proxyServers, err := core.FetchAllServers(&m.ServiceManager.DbClient)
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
	fmt.Println(placementConstraints)
	fmt.Println("HERE")

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
