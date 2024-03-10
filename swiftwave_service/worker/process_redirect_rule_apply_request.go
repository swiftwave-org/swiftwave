package worker

import (
	"context"
	"errors"
	haproxymanager "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"gorm.io/gorm"
	"log"
)

func (m Manager) RedirectRuleApply(request RedirectRuleApplyRequest, ctx context.Context, cancelContext context.CancelFunc) error {
	dbWithoutTx := m.ServiceManager.DbClient
	// fetch redirect rule
	redirectRule := &core.RedirectRule{}
	err := redirectRule.FindById(ctx, dbWithoutTx, request.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// ensure that the redirect rule is not being deleted
	if redirectRule.Status == core.RedirectRuleStatusDeleting {
		return nil
	}
	// fetch domain
	domain := &core.Domain{}
	err = domain.FindById(ctx, dbWithoutTx, redirectRule.DomainID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
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
	// map of server ip and transaction id
	transactionIdMap := make(map[*haproxymanager.Manager]string)
	isFailed := false
	defer func() {
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
	}()

	for _, haproxyManager := range haproxyManagers {
		// fetch haproxy transaction
		haproxyTransactionId, err := haproxyManager.FetchNewTransactionId()
		if err != nil {
			return err
		}
		transactionIdMap[haproxyManager] = haproxyTransactionId
		// add redirect
		if redirectRule.Protocol == core.HTTPProtocol {
			err = haproxyManager.AddHTTPRedirectRule(haproxyTransactionId, domain.Name, redirectRule.RedirectURL)
		} else {
			err = haproxyManager.AddHTTPSRedirectRule(haproxyTransactionId, domain.Name, redirectRule.RedirectURL)
		}
		if err != nil {
			// set status as failed and exit
			_ = redirectRule.UpdateStatus(ctx, dbWithoutTx, core.RedirectRuleStatusFailed)
			isFailed = true
			// no requeue
			return nil
		}
		// commit haproxy transaction
		err = haproxyManager.CommitTransaction(haproxyTransactionId)
		if err != nil {
			// set status as failed and exit
			_ = redirectRule.UpdateStatus(ctx, dbWithoutTx, core.RedirectRuleStatusFailed)
			isFailed = true
			// no requeue
			return nil
		}
	}
	// set status as applied
	err = redirectRule.UpdateStatus(ctx, dbWithoutTx, core.RedirectRuleStatusApplied)
	if err != nil {
		return err
	}

	return nil
}
