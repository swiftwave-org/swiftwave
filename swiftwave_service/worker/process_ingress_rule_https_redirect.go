package worker

import (
	"context"
	"errors"
	"log"

	haproxymanager "github.com/swiftwave-org/swiftwave/haproxy_manager"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/manager"
	"gorm.io/gorm"
)

func (m Manager) IngressRuleHttpsRedirect(request IngressRuleHttpsRedirectRequest, ctx context.Context, _ context.CancelFunc) error {
	dbTx := m.ServiceManager.DbClient.Begin()
	// fetch ingress rule
	var ingressRule = &core.IngressRule{}
	err := ingressRule.FindById(ctx, *dbTx, request.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	// check status should be deleting
	if ingressRule.Status == core.IngressRuleStatusDeleting {
		// dont requeue
		return nil
	}
	if ingressRule.Protocol != core.HTTPSProtocol {
		return nil
	}
	// fetch the domain
	domain := core.Domain{}
	err = domain.FindById(ctx, *dbTx, *ingressRule.DomainID)
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
	// don't attempt if no proxy servers are active
	if len(proxyServers) == 0 {
		return errors.New("no proxy servers are active")
	}
	// fetch all haproxy managers
	var haproxyManagers []*haproxymanager.Manager
	haproxyManagers, err = manager.HAProxyClients(context.Background(), proxyServers)
	if err != nil {
		return err

	}
	// create new transaction
	// map of server ip and transaction id
	transactionIdMap := make(map[*haproxymanager.Manager]string)
	isFailed := false

	for _, haproxyManager := range haproxyManagers {
		// create new haproxy transaction
		haproxyTransactionId, err := haproxyManager.FetchNewTransactionId()
		// store transaction id
		transactionIdMap[haproxyManager] = haproxyTransactionId
		if err != nil {
			continue
		}
	}

	for haproxyManager, haproxyTransactionId := range transactionIdMap {
		var err2 error
		if request.Enabled {
			err2 = haproxyManager.EnableHTTPSRedirection(haproxyTransactionId, domain.Name)
		} else {
			err2 = haproxyManager.DisableHTTPSRedirection(haproxyTransactionId, domain.Name)
		}
		if err2 != nil {
			isFailed = true
			break
		}
	}

	if !isFailed {
		err := ingressRule.UpdateHttpsRedirectStatus(ctx, *dbTx, request.Enabled)
		if err != nil {
			isFailed = true
			log.Println("failed to update ingress rule https redirect status", err)
		}
	}

	for haproxyManager, haproxyTransactionId := range transactionIdMap {
		if !isFailed {
			err = haproxyManager.CommitTransaction(haproxyTransactionId)
			if err != nil {
				log.Println("committing haproxy transaction", haproxyTransactionId, err)
			}
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
		dbTx.Rollback()
		return nil
	} else {
		return dbTx.Commit().Error
	}
}
