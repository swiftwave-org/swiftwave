package cronjob

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/swarm"
	containermanger "github.com/swiftwave-org/swiftwave/container_manager"
	"github.com/swiftwave-org/swiftwave/ssh_toolkit"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"strings"
	"time"
)

func (m Manager) SyncProxy() {
	for {
		m.syncProxy()
		time.Sleep(1 * time.Minute)
	}
}

func (m Manager) syncProxy() {
	// create context
	ctx := context.Background()
	// fetch all servers hostnames
	servers, err := core.FetchAllServers(&m.ServiceManager.DbClient)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch all proxy servers", err.Error())
		return
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
	if len(servers) == 0 {
		// delete haproxy and udpproxy services
		err = dockerClient.RemoveService(m.Config.LocalConfig.ServiceConfig.HAProxyServiceName)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to remove haproxy service", err.Error())
		} else {
			logger.CronJobLogger.Println("Removed haproxy service")
		}
		err = dockerClient.RemoveService(m.Config.LocalConfig.ServiceConfig.UDPProxyServiceName)
		if err != nil {
			logger.CronJobLoggerError.Println("Failed to remove udpproxy service", err.Error())
		} else {
			logger.CronJobLogger.Println("Removed udpproxy service")
		}
		return
	}
	// prepare placement constraints
	var placementConstraints []string
	for _, proxyServer := range servers {
		if !proxyServer.ProxyConfig.Enabled {
			placementConstraints = append(placementConstraints, "node.hostname!="+proxyServer.HostName)
		}
	}
	// haproxy
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
		if !isSameMap(haproxyService.Env, haProxyEnvironmentVariables) || strings.Compare(haproxyService.Image, m.Config.SystemConfig.HAProxyConfig.Image) != 0 || !isListSame(haproxyService.PlacementConstraints, placementConstraints) {
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
		} else {
			logger.CronJobLogger.Println("No change in haproxy service")
		}
	}
	// udp proxy
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
		} else {
			logger.CronJobLogger.Println("No change in udpproxy service")
		}
	}

	// PORT EXPOSER

	// fetch all exposed tcp ports
	tcpPorts, err := core.FetchAllExposedTCPPorts(ctx, m.ServiceManager.DbClient)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch all exposed tcp ports", err.Error())
		return
	}
	// add port 80 and 443
	tcpPorts = append(tcpPorts, 80, 443)
	tcpPorts = removeDuplicatesInt(tcpPorts)
	tcpPortsRule := make([]swarm.PortConfig, 0)
	for _, port := range tcpPorts {
		tcpPortsRule = append(tcpPortsRule, swarm.PortConfig{
			Protocol:      swarm.PortConfigProtocolTCP,
			PublishMode:   swarm.PortConfigPublishModeHost,
			TargetPort:    uint32(port),
			PublishedPort: uint32(port),
		})
	}
	// fetch all exposed udp ports
	udpPorts, err := core.FetchAllExposedUDPPorts(ctx, m.ServiceManager.DbClient)
	udpPorts = removeDuplicatesInt(udpPorts)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch all exposed udp ports", err.Error())
		return
	}
	udpPortsRule := make([]swarm.PortConfig, 0)
	for _, port := range udpPorts {
		udpPortsRule = append(udpPortsRule, swarm.PortConfig{
			Protocol:      swarm.PortConfigProtocolUDP,
			PublishMode:   swarm.PortConfigPublishModeHost,
			TargetPort:    uint32(port),
			PublishedPort: uint32(port),
		})
	}

	// fetch exposed tcp ports of haproxy service
	existingTcpPortRules, err := dockerClient.FetchPublishedPortRules(m.Config.LocalConfig.ServiceConfig.HAProxyServiceName)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch exposed tcp ports of haproxy service", err.Error())
		return
	} else {
		// check if exposed tcp ports are changed
		if !isPortListSame(existingTcpPortRules, tcpPortsRule) {
			// update exposed tcp ports
			err = dockerClient.UpdatePublishedHostPorts(m.Config.LocalConfig.ServiceConfig.HAProxyServiceName, tcpPortsRule)
			if err != nil {
				logger.CronJobLoggerError.Println("Failed to update exposed tcp ports of haproxy service", err.Error())
			} else {
				logger.CronJobLogger.Println("Updated exposed tcp ports of haproxy service")
			}
		} else {
			logger.CronJobLogger.Println("No change in exposed tcp ports of haproxy service")
		}
	}

	// fetch exposed udp ports of udpproxy service
	existingUdpPortRules, err := dockerClient.FetchPublishedPortRules(m.Config.LocalConfig.ServiceConfig.UDPProxyServiceName)
	if err != nil {
		logger.CronJobLoggerError.Println("Failed to fetch exposed udp ports of udpproxy service", err.Error())
		return
	} else {
		// check if exposed udp ports are changed
		if !isPortListSame(existingUdpPortRules, udpPortsRule) {
			// update exposed udp ports
			err = dockerClient.UpdatePublishedHostPorts(m.Config.LocalConfig.ServiceConfig.UDPProxyServiceName, udpPortsRule)
			if err != nil {
				logger.CronJobLoggerError.Println("Failed to update exposed udp ports of udpproxy service", err.Error())
			} else {
				logger.CronJobLogger.Println("Updated exposed udp ports of udpproxy service")
			}
		} else {
			logger.CronJobLogger.Println("No change in exposed udp ports of udpproxy service")
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

func removeDuplicatesInt(list []int) []int {
	keys := make(map[int]bool)
	var listWithoutDuplicates []int
	for _, entry := range list {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			listWithoutDuplicates = append(listWithoutDuplicates, entry)
		}
	}
	return listWithoutDuplicates
}

func isPortListSame(list1 []swarm.PortConfig, list2 []swarm.PortConfig) bool {
	if len(list1) != len(list2) {
		return false
	}
	for _, item1 := range list1 {
		found := false
		for _, item2 := range list2 {
			if item1.PublishedPort == item2.PublishedPort &&
				item1.TargetPort == item2.TargetPort &&
				item1.Protocol == item2.Protocol &&
				item1.PublishMode == item2.PublishMode {
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
