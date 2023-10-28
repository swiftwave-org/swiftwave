package core

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

func FindAllIngressRules(ctx context.Context, db gorm.DB) ([]*IngressRule, error) {
	var ingressRules []*IngressRule
	tx := db.Find(&ingressRules)
	return ingressRules, tx.Error
}

func (ingressRule *IngressRule) FindById(ctx context.Context, db gorm.DB, id uint) error {
	tx := db.First(&ingressRule, id)
	return tx.Error
}

func FindIngressRulesByDomainID(ctx context.Context, db gorm.DB, domainID uint) ([]*IngressRule, error) {
	var ingressRules []*IngressRule
	tx := db.Where("domain_id = ?", domainID).Find(&ingressRules)
	return ingressRules, tx.Error
}

func (ingressRule *IngressRule) Create(ctx context.Context, db gorm.DB) error {
	// verify if domain exist
	domain := &Domain{}
	err := domain.FindById(ctx, db, ingressRule.DomainID)
	if err != nil {
		return err
	}
	// verify if application exist
	application := &Application{}
	err = application.FindById(ctx, db, ingressRule.ApplicationID)
	if err != nil {
		return err
	}
	// verify there is no ingress rule with same domain and port
	isIngressRuleExist := db.Where("domain_id = ? AND port = ?", ingressRule.DomainID, ingressRule.Port).First(&IngressRule{}).RowsAffected > 0
	if isIngressRuleExist {
		return errors.New("there is ingress rule with same domain and port")
	}
	// verify there is no redirect rule with same domain and port
	isRedirectRuleExist := db.Where("domain_id = ? AND port = ?", ingressRule.DomainID, ingressRule.Port).First(&RedirectRule{}).RowsAffected > 0
	if isRedirectRuleExist {
		return errors.New("there is redirect rule with same domain and port")
	}

	// create record
	tx := db.Create(&ingressRule)
	return tx.Error
}

func (ingressRule *IngressRule) Update(ctx context.Context, db gorm.DB) error {
	return errors.New("update of ingress rule is not allowed")
}

func (ingressRule *IngressRule) Delete(ctx context.Context, db gorm.DB) error {
	// verify if ingress rule is not deleting
	isDeleting := ingressRule.isDeleting()
	if isDeleting {
		return errors.New("ingress rule is deleting")
	}
	// update status to deleting
	tx := db.Model(&ingressRule).Update("status", IngressRuleStatusDeleting)
	return tx.Error
}

func (ingressRule *IngressRule) isDeleting() bool {
	if ingressRule.Status == IngressRuleStatusDeleting {
		return true
	}
	return false
}
