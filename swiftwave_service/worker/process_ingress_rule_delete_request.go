package worker

import (
	"context"
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	UDP_PROXY "github.com/swiftwave-org/swiftwave/udp_proxy_manager"
	"gorm.io/gorm"
)

func (m Manager) IngressRuleDelete(request IngressRuleDeleteRequest, ctx context.Context, cancelContext context.CancelFunc) error {
	dbWithoutTx := m.ServiceManager.DbClient
	// restricted ports
	restrictedPorts := make([]int, 0)
	for _, port := range m.Config.SystemConfig.RestrictedPorts {
		restrictedPorts = append(restrictedPorts, int(port))
	}
	// fetch ingress rule
	var ingressRule core.IngressRule
	err := ingressRule.FindById(ctx, dbWithoutTx, request.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// check status should be deleting
	if ingressRule.Status != core.IngressRuleStatusDeleting {
		// dont requeue
		return nil
	}
	// fetch the domain
	domain := core.Domain{}
	if ingressRule.Protocol == core.HTTPProtocol || ingressRule.Protocol == core.HTTPSProtocol {
		if ingressRule.DomainID == nil {
			return errors.New("domain id is nil")
		}
		err = domain.FindById(ctx, dbWithoutTx, *ingressRule.DomainID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
	}

	// fetch application
	var application core.Application
	err = application.FindById(ctx, dbWithoutTx, ingressRule.ApplicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// generate backend name
	backendName := m.ServiceManager.HaproxyManager.GenerateBackendName(application.Name, int(ingressRule.TargetPort))
	// delete ingress rule from haproxy
	// create new haproxy transaction
	haproxyTransactionId, err := m.ServiceManager.HaproxyManager.FetchNewTransactionId()
	if err != nil {
		return err
	}
	// delete ingress rule
	if ingressRule.Protocol == core.HTTPSProtocol {
		err = m.ServiceManager.HaproxyManager.DeleteHTTPSLink(haproxyTransactionId, backendName, domain.Name)
		if err != nil {
			// set status as failed and exit
			// because `DeleteHTTPSLink` can fail only if haproxy not working
			deleteHaProxyTransaction(m, haproxyTransactionId)
			// requeue required as it fault of haproxy and may be resolved in next try
			return err
		}
	} else if ingressRule.Protocol == core.HTTPProtocol {
		if ingressRule.Port == 80 {
			err = m.ServiceManager.HaproxyManager.DeleteHTTPLink(haproxyTransactionId, backendName, domain.Name)
			if err != nil {
				// set status as failed and exit
				// because `DeleteHTTPLink` can fail only if haproxy not working
				deleteHaProxyTransaction(m, haproxyTransactionId)
				// requeue required as it fault of haproxy and may be resolved in next try
				return err
			}
		} else {
			err = m.ServiceManager.HaproxyManager.DeleteTCPLink(haproxyTransactionId, backendName, int(ingressRule.Port), domain.Name, restrictedPorts)
			if err != nil {
				// set status as failed and exit
				// because `DeleteTCPLink` can fail only if haproxy not working
				deleteHaProxyTransaction(m, haproxyTransactionId)
				// requeue required as it fault of haproxy and may be resolved in next try
				return err
			}
		}
	} else if ingressRule.Protocol == core.TCPProtocol {
		err = m.ServiceManager.HaproxyManager.DeleteTCPLink(haproxyTransactionId, backendName, int(ingressRule.Port), "", restrictedPorts)
		if err != nil {
			// set status as failed and exit
			// because `DeleteTCPLink` can fail only if haproxy not working
			deleteHaProxyTransaction(m, haproxyTransactionId)
			// requeue required as it fault of haproxy and may be resolved in next try
			return err
		}
	} else if ingressRule.Protocol == core.UDPProtocol {
		err = m.ServiceManager.UDPProxyManager.Remove(UDP_PROXY.Proxy{
			Port:       int(ingressRule.Port),
			TargetPort: int(ingressRule.TargetPort),
			Service:    application.Name,
		})
		if err != nil {
			return err
		}
	} else {
		// unknown protocol
		deleteHaProxyTransaction(m, haproxyTransactionId)
		return nil
	}

	// delete backend
	backendUsedByOther := true
	var ingressRuleCheck core.IngressRule
	err = m.ServiceManager.DbClient.Where("id != ? AND application_id = ? AND target_port = ?", ingressRule.ID, ingressRule.ApplicationID, ingressRule.TargetPort).First(&ingressRuleCheck).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			backendUsedByOther = false
		}
	}
	if !backendUsedByOther {
		err = m.ServiceManager.HaproxyManager.DeleteBackend(haproxyTransactionId, backendName)
		if err != nil {
			// set status as failed and exit
			// because `DeleteBackend` can fail only if haproxy not working
			deleteHaProxyTransaction(m, haproxyTransactionId)
			// requeue required as it fault of haproxy and may be resolved in next try
			return err
		}
	}

	// commit haproxy transaction
	err = m.ServiceManager.HaproxyManager.CommitTransaction(haproxyTransactionId)
	if err != nil {
		deleteHaProxyTransaction(m, haproxyTransactionId)
		// requeue required as it fault of haproxy and may be resolved in next try
		return err
	}

	// delete ingress rule from database
	err = ingressRule.Delete(ctx, dbWithoutTx, true)
	if err != nil {
		return err
	}

	return nil
}
