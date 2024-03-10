package cronjob

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/swarm"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"log"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

func (m Manager) HaProxyPortExposer() {
	for {
		// Fetch a random swarm manager
		swarmManagerServer, err := core.FetchSwarmManager(&m.ServiceManager.DbClient)
		if err != nil {
			log.Println(err)
			continue
		}
		// Fetch docker manager
		dockerManager, err := manager.DockerClient(context.Background(), swarmManagerServer)
		if err != nil {
			log.Println(err)
			continue
		}
		// Fetch all ingress rules with only port field
		var ingressRules []core.IngressRule
		tx := m.ServiceManager.DbClient.Select("port").Where("port IS NOT NULL").Not("protocol = ?", "udp").Find(&ingressRules)
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
		// Check if ports are changed
		exposedPorts, err := dockerManager.FetchPublishedHostPorts(m.Config.LocalConfig.ServiceConfig.HAProxyServiceName)
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
			err := dockerManager.UpdatePublishedHostPorts(m.Config.LocalConfig.ServiceConfig.HAProxyServiceName, portsUpdateRequired)
			if err != nil {
				log.Println(err)
			} else {
				log.Println("Exposed ports of haproxy service updated")
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
					err := firewallDenyPort(m.Config.SystemConfig.FirewallConfig.DenyPortCommand, port)
					if err != nil {
						log.Printf("Failed to deny port %d in firewall", port)
					} else {
						log.Printf("Port %d denied", port)
					}
				}
				// Allow exposed ports
				for port := range portsMap {
					err := firewallAllowPort(m.Config.SystemConfig.FirewallConfig.AllowPortCommand, port)
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

func firewallDenyPort(commandTemplate string, port int) error {
	// TODO: replace with ssh
	command := strings.ReplaceAll(commandTemplate, "{{PORT}}", fmt.Sprintf("%d", port))
	// Run using os package
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Run()
	if err != nil {
		return err
	} else {
		return nil
	}
}

func firewallAllowPort(commandTemplate string, port int) error {
	// TODO: replace with ssh
	command := strings.ReplaceAll(commandTemplate, "{{PORT}}", fmt.Sprintf("%d", port))
	// Run using os package
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Run()
	if err != nil {
		return err
	} else {
		return nil
	}
}
