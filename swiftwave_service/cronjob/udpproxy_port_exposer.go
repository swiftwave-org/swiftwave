package cronjob

import (
	"context"
	"github.com/docker/docker/api/types/swarm"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/logger"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"reflect"
	"time"
)

func (m Manager) UDPProxyPortExposer() {
	isFirstTime := true
	for {
		if isFirstTime {
			// wait 10 seconds
			time.Sleep(5 * time.Second)
			isFirstTime = false
		} else {
			time.Sleep(20 * time.Second)
		}
		// Fetch a random swarm manager
		swarmManagerServer, err := core.FetchSwarmManager(&m.ServiceManager.DbClient)
		if err != nil {
			logger.CronJobLoggerError.Println(err)
			continue
		}
		// Fetch docker manager
		dockerManager, err := manager.DockerClient(context.Background(), swarmManagerServer)
		if err != nil {
			logger.CronJobLoggerError.Println(err)
			continue
		}
		// Fetch all proxy servers
		proxyServers, err := core.FetchAllProxyServers(&m.ServiceManager.DbClient)
		if err != nil {
			logger.CronJobLoggerError.Println(err)
			continue
		}
		// Fetch all ingress rules with only port field
		var ingressRules []core.IngressRule
		tx := m.ServiceManager.DbClient.Select("port").Where("port IS NOT NULL").Where("protocol = ?", "udp").Find(&ingressRules)
		if tx.Error != nil {
			logger.CronJobLoggerError.Println(tx.Error)
			continue
		}
		// Serialize port
		var portsMap = make(map[int]bool)
		for _, ingressRule := range ingressRules {
			portsMap[int(ingressRule.Port)] = true
		}
		// Check if ports are changed
		exposedPorts, err := dockerManager.FetchPublishedHostPorts(m.Config.LocalConfig.ServiceConfig.UDPProxyServiceName)
		if err != nil {
			logger.CronJobLoggerError.Println(err)
			continue
		}
		exposedPortsMap := make(map[int]bool)
		for _, port := range exposedPorts {
			exposedPortsMap[port] = true
		}
		portsNotChanged := reflect.DeepEqual(exposedPortsMap, portsMap)
		if !portsNotChanged {
			var portsUpdateRequired = make([]swarm.PortConfig, 0)
			for port := range portsMap {
				portsUpdateRequired = append(portsUpdateRequired, swarm.PortConfig{
					Protocol:      swarm.PortConfigProtocolUDP,
					PublishMode:   swarm.PortConfigPublishModeHost,
					TargetPort:    uint32(port),
					PublishedPort: uint32(port),
				})
			}
			// Update exposed ports
			err := dockerManager.UpdatePublishedHostPorts(m.Config.LocalConfig.ServiceConfig.UDPProxyServiceName, portsUpdateRequired)
			if err != nil {
				logger.CronJobLoggerError.Println(err)
			} else {
				logger.CronJobLogger.Println("Exposed ports of udp proxy service updated")
			}
			// Update firewall
			if m.Config.SystemConfig.FirewallConfig.Enabled {
				// Find out the ports that are unexposed
				var unexposedPorts = make([]int, 0)
				for port := range exposedPortsMap {
					if _, ok := portsMap[port]; !ok {
						unexposedPorts = append(unexposedPorts, port)
					}
				}
				// Deny unexposed ports
				for _, port := range unexposedPorts {
					err := firewallDenyPort(proxyServers, m.Config.SystemConfig.FirewallConfig.DenyPortCommand, port)
					if err != nil {
						logger.CronJobLoggerError.Printf("Failed to deny port %d in firewall", port)
					} else {
						logger.CronJobLogger.Printf("Port %d denied", port)
					}
				}
				// Allow exposed ports
				for port := range portsMap {
					err := firewallAllowPort(proxyServers, m.Config.SystemConfig.FirewallConfig.AllowPortCommand, port)
					if err != nil {
						logger.CronJobLoggerError.Printf("Failed to allow port %d in firewall", port)
					} else {
						logger.CronJobLogger.Printf("Port %d allowed", port)
					}
				}
			}
		}
	}
}
