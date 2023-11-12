package cronjob

import (
	"github.com/docker/docker/api/types/swarm"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"log"
	"reflect"
	"time"
)

func (m Manager) HaProxyPortExposer() {
	for {
		// Fetch all ingress rules with only port field
		var ingressRules []core.IngressRule
		tx := m.ServiceManager.DbClient.Select("port").Where("port IS NOT NULL").Find(&ingressRules)
		if tx.Error != nil {
			log.Println(tx.Error)
			continue
		}
		// Serialize port
		var portsMap = make(map[int]bool)
		for _, ingressRule := range ingressRules {
			portsMap[int(ingressRule.Port)] = true
		}
		// add 80 and 443 to ports
		portsMap[80] = true
		portsMap[443] = true
		portsMap[5555] = true
		// Check if ports are changed
		exposedPorts, err := m.ServiceManager.DockerManager.FetchPublishedHostPorts(m.ServiceConfig.HAProxyConfig.ServiceName)
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
					Protocol:      swarm.PortConfigProtocolTCP,
					PublishMode:   swarm.PortConfigPublishModeHost,
					TargetPort:    uint32(port),
					PublishedPort: uint32(port),
				})
			}
			// Update exposed ports
			err := m.ServiceManager.DockerManager.UpdatePublishedHostPorts(m.ServiceConfig.HAProxyConfig.ServiceName, portsUpdateRequired)
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Exposed ports of haproxy service updated")
			}
		}
		time.Sleep(20 * time.Second)
	}
	m.wg.Done()
}
