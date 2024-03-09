package cronjob

import (
	"github.com/docker/docker/api/types/swarm"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"log"
	"reflect"
	"time"
)

func (m Manager) UDPProxyPortExposer() {
	for {
		// Fetch all ingress rules with only port field
		var ingressRules []core.IngressRule
		tx := m.ServiceManager.DbClient.Select("port").Where("port IS NOT NULL").Where("protocol = ?", "udp").Find(&ingressRules)
		if tx.Error != nil {
			log.Println(tx.Error)
			continue
		}
		// Serialize port
		var portsMap = make(map[int]bool)
		for _, ingressRule := range ingressRules {
			portsMap[int(ingressRule.Port)] = true
		}
		// Check if ports are changed
		exposedPorts, err := m.ServiceManager.DockerManager.FetchPublishedHostPorts(m.Config.UDPProxyConfig.ServiceName)
		if err != nil {
			log.Println(err)
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
			err := m.ServiceManager.DockerManager.UpdatePublishedHostPorts(m.Config.UDPProxyConfig.ServiceName, portsUpdateRequired)
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Exposed ports of udp proxy service updated")
			}
			// Update firewall
			if m.Config.ServiceConfig.FirewallEnabled {
				// Find out the ports that are unexposed
				var unexposedPorts = make([]int, 0)
				for port := range exposedPortsMap {
					if _, ok := portsMap[port]; !ok {
						unexposedPorts = append(unexposedPorts, port)
					}
				}
				// Deny unexposed ports
				for _, port := range unexposedPorts {
					err := firewallDenyPort(m.Config.ServiceConfig.FirewallDenyPortCommand, port)
					if err != nil {
						log.Printf("Failed to deny port %d in firewall", port)
					} else {
						log.Printf("Port %d denied", port)
					}
				}
				// Allow exposed ports
				for port := range portsMap {
					err := firewallAllowPort(m.Config.ServiceConfig.FirewallAllowPortCommand, port)
					if err != nil {
						log.Printf("Failed to allow port %d in firewall", port)
					} else {
						log.Printf("Port %d allowed", port)
					}
				}
			}
		}
		time.Sleep(20 * time.Second)
	}
}
