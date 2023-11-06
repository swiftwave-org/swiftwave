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
	isIngressRuleExist := db.Where("domain_id = ? AND port = ?", redirectRule.DomainID, redirectRule.Port).First(&IngressRule{}).RowsAffected > 0
	if isIngressRuleExist {
		return errors.New("there is ingress rule with same domain and port")
	}
	// verify if there is no redirect rule with same domain and port
	isRedirectRuleExist := db.Where("domain_id = ? AND port = ?", redirectRule.DomainID, redirectRule.Port).First(&RedirectRule{}).RowsAffected > 0
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
	}
	// update record
	tx := db.Model(&redirectRule).Update("status", RedirectRuleStatusDeleting)
	return tx.Error
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
