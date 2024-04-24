package worker

import (
	"context"
	"errors"
	haproxymanager "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	udpproxymanager "github.com/swiftwave-org/swiftwave/udp_proxy_manager"
	"gorm.io/gorm"
	"log"
)

func (m Manager) IngressRuleApply(request IngressRuleApplyRequest, ctx context.Context, _ context.CancelFunc) error {
	dbWithoutTx := m.ServiceManager.DbClient
	// restricted ports
	restrictedPorts := make([]int, 0)
	for _, port := range m.Config.SystemConfig.RestrictedPorts {
		restrictedPorts = append(restrictedPorts, int(port))
	}
	// fetch ingress rule
	ingressRule := &core.IngressRule{}
	err := ingressRule.FindById(ctx, dbWithoutTx, request.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if ingressRule.Status == core.IngressRuleStatusDeleting {
		return nil
	}
	domain := &core.Domain{}
	if ingressRule.Protocol == core.HTTPSProtocol || ingressRule.Protocol == core.HTTPProtocol {
		// fetch domain
		if ingressRule.DomainID == nil {
			return errors.New("domain id is nil")
		}
		err = domain.FindById(ctx, dbWithoutTx, *ingressRule.DomainID)
		if err != nil {
			return err
		}
	}

	// service name
	serviceName := ""
	var serviceReplicas uint = 1

	if ingressRule.TargetType == core.ApplicationIngressRule {
		// fetch application
		application := &core.Application{}
		err = application.FindById(ctx, dbWithoutTx, *ingressRule.ApplicationID)
		if err != nil {
			return err
		}
		serviceName = application.Name
		serviceReplicas = application.Replicas
	} else if ingressRule.TargetType == core.ExternalServiceIngressRule {
		serviceName = ingressRule.ExternalService
	} else {
		return nil
	}

	// fetch all proxy servers
	proxyServers, err := core.FetchProxyActiveServers(&m.ServiceManager.DbClient)
	if err != nil {
		return err
	}
	// don't attempt if no proxy servers are active
	if len(proxyServers) == 0 {
		return errors.New("no proxy servers are active")
	}
	// fetch all haproxy managers
	var haproxyManagers []*haproxymanager.Manager
	if isHAProxyAccessRequired(ingressRule) {
		haproxyManagers, err = manager.HAProxyClients(context.Background(), proxyServers)
		if err != nil {
			return err
		}
	}
	// fetch all udp proxy managers
	var udpProxyManagers []*udpproxymanager.Manager
	if isUDProxyAccessRequired(ingressRule) {
		udpProxyManagers, err = manager.UDPProxyClients(context.Background(), proxyServers)
		if err != nil {
			return err
		}
	}
	// map of server ip and transaction id
	transactionIdMap := make(map[*haproxymanager.Manager]string)
	var isFailed bool

	for _, haproxyManager := range haproxyManagers {
		// check if ingress rules is not udp based
		if ingressRule.Protocol == core.UDPProtocol {
			continue
		}
		// backend protocol
		backendProtocol := ingressRuleProtocolToBackendProtocol(ingressRule.Protocol)
		// create new haproxy transaction
		haproxyTransactionId, err := haproxyManager.FetchNewTransactionId()
		if err != nil {
			isFailed = true
			break
		}
		// add to map
		transactionIdMap[haproxyManager] = haproxyTransactionId
		// generate backend name
		backendName := haproxyManager.GenerateBackendName(backendProtocol, serviceName, int(ingressRule.TargetPort))
		// add backend
		_, err = haproxyManager.AddBackend(haproxyTransactionId, backendProtocol, serviceName, int(ingressRule.TargetPort), int(serviceReplicas))
		if err != nil {
			isFailed = true
			break
		}
		// add frontend
		if ingressRule.Protocol == core.HTTPSProtocol {
			err = haproxyManager.AddHTTPSLink(haproxyTransactionId, backendName, domain.Name)
			if err != nil {
				isFailed = true
				break
			}
		} else if ingressRule.Protocol == core.HTTPProtocol {
			// for default port 80, should use fe_http frontend due to some binding restrictions
			if ingressRule.Port == 80 {
				err = haproxyManager.AddHTTPLink(haproxyTransactionId, backendName, domain.Name)
				if err != nil {
					isFailed = true
					break
				}
			} else {
				// for other ports, use custom frontend
				err = haproxyManager.AddTCPLink(haproxyTransactionId, backendName, int(ingressRule.Port), domain.Name, haproxymanager.HTTPMode, restrictedPorts)
				if err != nil {
					isFailed = true
					break
				}
			}
		} else if ingressRule.Protocol == core.TCPProtocol {
			err = haproxyManager.AddTCPLink(haproxyTransactionId, backendName, int(ingressRule.Port), "", haproxymanager.TCPMode, restrictedPorts)
			if err != nil {
				isFailed = true
				break
			}
		} else {
			isFailed = true
			break
		}
	}

	for _, udpProxyManager := range udpProxyManagers {
		if ingressRule.Protocol == core.UDPProtocol {
			err = udpProxyManager.Add(udpproxymanager.Proxy{
				Port:       int(ingressRule.Port),
				TargetPort: int(ingressRule.TargetPort),
				Service:    serviceName,
			}, restrictedPorts)
			if err != nil {
				isFailed = true
				break
			}
		}
	}

	for haproxyManager, haproxyTransactionId := range transactionIdMap {
		if !isFailed {
			// commit the haproxy transaction
			err = haproxyManager.CommitTransaction(haproxyTransactionId)
		}
		if isFailed || err != nil {
			isFailed = true
			log.Println("failed to commit haproxy transaction", err)
			err := haproxyManager.DeleteTransaction(haproxyTransactionId)
			if err != nil {
				log.Println("failed to rollback haproxy transaction", err)
			}
		}
	}

	if isFailed {
		return ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
	} else {
		// update status as applied
		return ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusApplied)
	}
}
