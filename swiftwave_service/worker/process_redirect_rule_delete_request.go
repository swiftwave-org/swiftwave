package worker

import (
	"context"
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
)

func (m Manager) RedirectRuleDelete(request RedirectRuleDeleteRequest, ctx context.Context, cancelContext context.CancelFunc) error {
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
	// create new haproxy transaction
	haproxyTransactionId, err := m.ServiceManager.HaproxyManager.FetchNewTransactionId()
	if err != nil {
		return err
	}
	// delete redirect rule
	if redirectRule.Protocol == core.HTTPProtocol {
		err = m.ServiceManager.HaproxyManager.DeleteHTTPRedirectRule(haproxyTransactionId, domain.Name)
	} else if redirectRule.Protocol == core.HTTPSProtocol {
		err = m.ServiceManager.HaproxyManager.DeleteHTTPSRedirectRule(haproxyTransactionId, domain.Name)
	} else {
		// invalid protocol
		return nil
	}
	if err != nil {
		// set status as failed and exit
		// because `DeleteHTTPRedirectRule` can fail only if haproxy not working
		deleteHaProxyTransaction(m, haproxyTransactionId)
		// requeue required as it fault of haproxy and may be resolved in next try
		return err
	}
	// commit haproxy transaction
	err = m.ServiceManager.HaproxyManager.CommitTransaction(haproxyTransactionId)
	if err != nil {
		deleteHaProxyTransaction(m, haproxyTransactionId)
		// requeue required as it fault of haproxy and may be resolved in next try
		return err
	}
	// delete redirect rule from database
	err = redirectRule.Delete(ctx, dbWithoutTx, true)
	if err != nil {
		return err
	}
	return nil
}
