package core

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

// This file contains the operations for the RedirectRule model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindAllRedirectRules(ctx context.Context, db gorm.DB) ([]*RedirectRule, error) {
	var redirectRules []*RedirectRule
	tx := db.Find(&redirectRules)
	return redirectRules, tx.Error
}

func (redirectRule *RedirectRule) FindById(ctx context.Context, db gorm.DB, id uint) error {
	tx := db.Where("id = ?", id).First(&redirectRule)
	return tx.Error
}

func FindRedirectRulesByDomainID(ctx context.Context, db gorm.DB, domainID uint) ([]*RedirectRule, error) {
	var redirectRules []*RedirectRule
	tx := db.Where("domain_id = ?", domainID).Find(&redirectRules)
	return redirectRules, tx.Error
}

func (redirectRule *RedirectRule) Create(ctx context.Context, db gorm.DB) error {
	// verify if domain exist
	domain := &Domain{}
	err := domain.FindById(ctx, db, redirectRule.DomainID)
	if err != nil {
		return err
	}
	// verify if there is no ingress rule with same domain and port
	if redirectRule.Protocol != HTTPProtocol && redirectRule.Protocol != HTTPSProtocol {
		return errors.New("invalid protocol")
	}
	var isIngressRuleExist bool
	/*
	 * For redirect rule with HTTP protocol, port will be anyhow 80
	 * For redirect rule with HTTPS protocol, port will be anyhow 443
	 * So, we can check if there is any ingress rule with same domain and port
	 */
	if redirectRule.Protocol == HTTPProtocol {
		isIngressRuleExist = db.Where("domain_id = ? AND protocol = ? AND port = ?", redirectRule.DomainID, redirectRule.Protocol, 80).First(&IngressRule{}).RowsAffected > 0
	} else if redirectRule.Protocol == HTTPSProtocol {
		isIngressRuleExist = db.Where("domain_id = ? AND protocol = ? AND port = ?", redirectRule.DomainID, redirectRule.Protocol, 443).First(&IngressRule{}).RowsAffected > 0
	}
	if isIngressRuleExist {
		return errors.New("there is ingress rule with same domain and port")
	}
	// verify if there is no redirect rule with same domain and protocl
	isRedirectRuleExist := db.Where("domain_id = ? AND protocol = ?", redirectRule.DomainID, redirectRule.Protocol).First(&RedirectRule{}).RowsAffected > 0
	if isRedirectRuleExist {
		return errors.New("there is redirect rule with same domain and port")
	}
	// create record
	tx := db.Create(&redirectRule)
	return tx.Error
}

func (redirectRule *RedirectRule) Update(ctx context.Context, db gorm.DB) error {
	return errors.New("update of redirect rule is not allowed")
}

func (redirectRule *RedirectRule) Delete(ctx context.Context, db gorm.DB, force bool) error {
	if !force {
		// verify if redirect rule is not deleting
		if redirectRule.isDeleting() {
			return errors.New("redirect rule is already deleting")
		}
		// update record
		tx := db.Model(&redirectRule).Update("status", RedirectRuleStatusDeleting)
		return tx.Error
	} else {
		// delete record
		tx := db.Delete(&redirectRule)
		return tx.Error
	}
}

func (redirectRule *RedirectRule) isDeleting() bool {
	if redirectRule.Status == RedirectRuleStatusDeleting {
		return true
	}
	return false
}

func (redirectRule *RedirectRule) UpdateStatus(ctx context.Context, db gorm.DB, status RedirectRuleStatus) error {
	// update record
	tx := db.Model(&redirectRule).Update("status", status)
	return tx.Error

}
