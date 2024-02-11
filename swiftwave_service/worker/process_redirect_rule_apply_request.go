package worker

import (
	"context"
	"errors"
	"github.com/swiftwave-org/swiftwave/swiftwave_service/core"
	"gorm.io/gorm"
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
	// fetch haproxy transaction
	haproxyTransactionId, err := m.ServiceManager.HaproxyManager.FetchNewTransactionId()
	if err != nil {
		return err
	}
	// add redirect
	if redirectRule.Protocol == core.HTTPProtocol {
		err = m.ServiceManager.HaproxyManager.AddHTTPRedirectRule(haproxyTransactionId, domain.Name, redirectRule.RedirectURL)
	} else {
		err = m.ServiceManager.HaproxyManager.AddHTTPSRedirectRule(haproxyTransactionId, domain.Name, redirectRule.RedirectURL)
	}
	if err != nil {
		// set status as failed and exit
		_ = redirectRule.UpdateStatus(ctx, dbWithoutTx, core.RedirectRuleStatusFailed)
		deleteHaProxyTransaction(m, haproxyTransactionId)
		// no requeue
		return nil
	}
	// commit haproxy transaction
	err = m.ServiceManager.HaproxyManager.CommitTransaction(haproxyTransactionId)
	if err != nil {
		// set status as failed and exit
		_ = redirectRule.UpdateStatus(ctx, dbWithoutTx, core.RedirectRuleStatusFailed)
		deleteHaProxyTransaction(m, haproxyTransactionId)
		// no requeue
		return nil
	}
	// set status as applied
	err = redirectRule.UpdateStatus(ctx, dbWithoutTx, core.RedirectRuleStatusApplied)
	if err != nil {
		return err
	}

	return nil
}
