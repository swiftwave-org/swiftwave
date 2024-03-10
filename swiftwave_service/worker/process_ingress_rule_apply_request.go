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

	// fetch application
	application := &core.Application{}
	err = application.FindById(ctx, dbWithoutTx, ingressRule.ApplicationID)
	if err != nil {
		return err
	}

	// fetch all proxy servers
	proxyServers, err := core.FetchProxyActiveServers(&m.ServiceManager.DbClient)
	if err != nil {
		return err
	}
	// fetch all haproxy managers
	haproxyManagers, err := manager.HAProxyClients(context.Background(), proxyServers)
	if err != nil {
		return err
	}
	// fetch all udp proxy managers
	udpProxyManagers, err := manager.UDPProxyClients(context.Background(), proxyServers)
	if err != nil {
		return err
	}
	// map of server ip and transaction id
	transactionIdMap := make(map[*haproxymanager.Manager]string)
	isFailed := false

	for _, haproxyManager := range haproxyManagers {
		// create new haproxy transaction
		haproxyTransactionId, err := haproxyManager.FetchNewTransactionId()
		if err != nil {
			isFailed = true
			break
		}
		// add to map
		transactionIdMap[haproxyManager] = haproxyTransactionId
		// generate backend name
		backendName := haproxyManager.GenerateBackendName(application.Name, int(ingressRule.TargetPort))
		// add backend
		_, err = haproxyManager.AddBackend(haproxyTransactionId, application.Name, int(ingressRule.TargetPort), int(application.Replicas))
		if err != nil {
			isFailed = true
			// set status as failed and exit
			_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
			// no requeue
			return nil
		}
		// add frontend
		if ingressRule.Protocol == core.HTTPSProtocol {
			err = haproxyManager.AddHTTPSLink(haproxyTransactionId, backendName, domain.Name)
			if err != nil {
				isFailed = true
				// set status as failed and exit
				_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
				// no requeue
				return nil
			}
		} else if ingressRule.Protocol == core.HTTPProtocol {
			// for default port 80, should use fe_http frontend due to some binding restrictions
			if ingressRule.Port == 80 {
				err = haproxyManager.AddHTTPLink(haproxyTransactionId, backendName, domain.Name)
				if err != nil {
					isFailed = true
					// set status as failed and exit
					_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
					// no requeue
					return nil
				}
			} else {
				// for other ports, use custom frontend
				err = haproxyManager.AddTCPLink(haproxyTransactionId, backendName, int(ingressRule.Port), domain.Name, haproxymanager.HTTPMode, restrictedPorts)
				if err != nil {
					isFailed = true
					// set status as failed and exit
					_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
					// no requeue
					return nil
				}
			}
		} else if ingressRule.Protocol == core.TCPProtocol {
			err = haproxyManager.AddTCPLink(haproxyTransactionId, backendName, int(ingressRule.Port), "", haproxymanager.TCPMode, restrictedPorts)
			if err != nil {
				isFailed = true
				// set status as failed and exit
				_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
				// no requeue
				return nil
			}
		} else if ingressRule.Protocol == core.UDPProtocol {
			// will be handled by udp proxy
		} else {
			isFailed = true
			// set status as failed and exit
			_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
			// no requeue
			return nil
		}
	}

	for _, udpProxyManager := range udpProxyManagers {
		if ingressRule.Protocol == core.UDPProtocol {
			err = udpProxyManager.Add(udpproxymanager.Proxy{
				Port:       int(ingressRule.Port),
				TargetPort: int(ingressRule.TargetPort),
				Service:    application.Name,
			}, restrictedPorts)
			if err != nil {
				isFailed = true
				// set status as failed and exit
				_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
				// no requeue
				return nil
			}
		}
	}

	for haproxyManager, haproxyTransactionId := range transactionIdMap {
		if !isFailed {
			// commit the haproxy transaction
			err = haproxyManager.CommitTransaction(haproxyTransactionId)
		}
		if isFailed || err != nil {
			log.Println("failed to commit haproxy transaction", err)
			err := haproxyManager.DeleteTransaction(haproxyTransactionId)
			if err != nil {
				log.Println("failed to rollback haproxy transaction", err)
			}
		}
	}
	manager.KillAllHAProxyConnections(haproxyManagers)
	manager.KillAllUDPProxyConnections(udpProxyManagers)

	if isFailed {
		// set status as failed and exit
		_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
		// no requeue
		return nil
	}

	// update status as applied
	err = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusApplied)
	if err != nil {
		// requeue because this error can lead to block stage of application
		return err
	}

	// success
	return nil
}
