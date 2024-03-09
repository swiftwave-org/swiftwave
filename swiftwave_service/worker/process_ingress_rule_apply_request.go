package worker

import (
	"context"
	"errors"
	haproxymanager "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	UDP_PROXY "github.com/swiftwave-org/swiftwave/udp_proxy_manager"
	"gorm.io/gorm"
	"log"
)

func (m Manager) IngressRuleApply(request IngressRuleApplyRequest, ctx context.Context, cancelContext context.CancelFunc) error {
	dbWithoutTx := m.ServiceManager.DbClient
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
	// create new haproxy transaction
	haproxyTransactionId, err := m.ServiceManager.HaproxyManager.FetchNewTransactionId()
	if err != nil {
		return err
	}
	// generate backend name
	backendName := m.ServiceManager.HaproxyManager.GenerateBackendName(application.Name, int(ingressRule.TargetPort))
	// add backend
	_, err = m.ServiceManager.HaproxyManager.AddBackend(haproxyTransactionId, application.Name, int(ingressRule.TargetPort), int(application.Replicas))
	if err != nil {
		// set status as failed and exit
		_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
		deleteHaProxyTransaction(m, haproxyTransactionId)
		// no requeue
		return nil
	}
	// add frontend
	if ingressRule.Protocol == core.HTTPSProtocol {
		err = m.ServiceManager.HaproxyManager.AddHTTPSLink(haproxyTransactionId, backendName, domain.Name)
		if err != nil {
			// set status as failed and exit
			_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
			deleteHaProxyTransaction(m, haproxyTransactionId)
			// no requeue
			return nil
		}
	} else if ingressRule.Protocol == core.HTTPProtocol {
		// for default port 80, should use fe_http frontend due to some binding restrictions
		if ingressRule.Port == 80 {
			err = m.ServiceManager.HaproxyManager.AddHTTPLink(haproxyTransactionId, backendName, domain.Name)
			if err != nil {
				// set status as failed and exit
				_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
				deleteHaProxyTransaction(m, haproxyTransactionId)
				// no requeue
				return nil
			}
		} else {
			// for other ports, use custom frontend
			err = m.ServiceManager.HaproxyManager.AddTCPLink(haproxyTransactionId, backendName, int(ingressRule.Port), domain.Name, haproxymanager.HTTPMode, m.Config.ServiceConfig.RestrictedPorts)
			if err != nil {
				// set status as failed and exit
				_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
				deleteHaProxyTransaction(m, haproxyTransactionId)
				// no requeue
				return nil
			}
		}
	} else if ingressRule.Protocol == core.TCPProtocol {
		err = m.ServiceManager.HaproxyManager.AddTCPLink(haproxyTransactionId, backendName, int(ingressRule.Port), "", haproxymanager.TCPMode, m.Config.ServiceConfig.RestrictedPorts)
		if err != nil {
			// set status as failed and exit
			_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
			deleteHaProxyTransaction(m, haproxyTransactionId)
			// no requeue
			return nil
		}
	} else if ingressRule.Protocol == core.UDPProtocol {
		err = m.ServiceManager.UDPProxyManager.Add(UDP_PROXY.Proxy{
			Port:       int(ingressRule.Port),
			TargetPort: int(ingressRule.TargetPort),
			Service:    application.Name,
		}, m.Config.ServiceConfig.RestrictedPorts)
		if err != nil {
			// set status as failed and exit
			_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
			// no requeue
			return nil
		}
	} else {
		// set status as failed and exit
		_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
		deleteHaProxyTransaction(m, haproxyTransactionId)
		// no requeue
		return nil
	}

	// commit haproxy transaction
	err = m.ServiceManager.HaproxyManager.CommitTransaction(haproxyTransactionId)
	if err != nil {
		// set status as failed and exit
		_ = ingressRule.UpdateStatus(ctx, dbWithoutTx, core.IngressRuleStatusFailed)
		deleteHaProxyTransaction(m, haproxyTransactionId)
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

// private functions
func deleteHaProxyTransaction(m Manager, haproxyTransactionId string) {
	err := m.ServiceManager.HaproxyManager.DeleteTransaction(haproxyTransactionId)
	if err != nil {
		log.Println("error while deleting haproxy transaction")
	}
}
