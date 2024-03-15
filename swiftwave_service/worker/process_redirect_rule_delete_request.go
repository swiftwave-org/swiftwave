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

func (m Manager) RedirectRuleDelete(request RedirectRuleDeleteRequest, ctx context.Context, _ context.CancelFunc) error {
	dbWithoutTx := m.ServiceManager.DbClient
	// fetch redirect rule
	var redirectRule core.RedirectRule
	err := redirectRule.FindById(ctx, dbWithoutTx, request.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// check status should be deleting
	if redirectRule.Status != core.RedirectRuleStatusDeleting {
		// dont requeue
		return nil
	}
	// fetch the domain
	var domain core.Domain
	err = domain.FindById(ctx, dbWithoutTx, redirectRule.DomainID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// delete redirect rule from haproxy
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
	haproxyManagers, err := manager.HAProxyClients(context.Background(), proxyServers)
	if err != nil {
		return err
	}
	// map of server ip and transaction id
	transactionIdMap := make(map[*haproxymanager.Manager]string)
	isFailed := false

	// create new haproxy transaction
	for _, haproxyManager := range haproxyManagers {
		haproxyTransactionId, err := haproxyManager.FetchNewTransactionId()
		if err != nil {
			return err
		}
		transactionIdMap[haproxyManager] = haproxyTransactionId
		// delete redirect rule
		if redirectRule.Protocol == core.HTTPProtocol {
			err = haproxyManager.DeleteHTTPRedirectRule(haproxyTransactionId, domain.Name)
		} else if redirectRule.Protocol == core.HTTPSProtocol {
			err = haproxyManager.DeleteHTTPSRedirectRule(haproxyTransactionId, domain.Name)
		} else {
			// invalid protocol
			return nil
		}
		if err != nil {
			// set status as failed and exit
			// because `DeleteHTTPRedirectRule` can fail only if haproxy not working
			//nolint:ineffassign
			isFailed = true
			// requeue required as it fault of haproxy and may be resolved in next try
			return err
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

	// delete redirect rule from database
	err = redirectRule.Delete(ctx, dbWithoutTx, true)
	if err != nil {
		return err
	}
	return nil
}
