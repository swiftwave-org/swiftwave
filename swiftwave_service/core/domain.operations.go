package core

import (
	"context"
	"errors"
	"gorm.io/gorm"
)

// This file contains the operations for the Domain model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindAllDomains(ctx context.Context, db gorm.DB) ([]*Domain, error) {
	var domains []*Domain
	tx := db.Find(&domains)
	return domains, tx.Error
}

func (domain *Domain) FindById(ctx context.Context, db gorm.DB, id uint) error {
	tx := db.First(&domain, id)
	return tx.Error
}

func (domain *Domain) Create(ctx context.Context, db gorm.DB) error {
	tx := db.Create(&domain)
	return tx.Error
}

func (domain *Domain) Update(ctx context.Context, db gorm.DB) error {
	tx := db.Save(&domain)
	return tx.Error
}

func (domain *Domain) Delete(ctx context.Context, db gorm.DB) error {
	// Make sure there is no ingress rule or redirect rule associated with this domain
	isIngressRuleExist := db.Where("domain_id = ?", domain.ID).First(&IngressRule{}).RowsAffected > 0
	if isIngressRuleExist {
		return errors.New("there is ingress rule associated with this domain")
	}
	isRedirectRuleExist := db.Where("domain_id = ?", domain.ID).First(&RedirectRule{}).RowsAffected > 0
	if isRedirectRuleExist {
		return errors.New("there is redirect rule associated with this domain")
	}
	tx := db.Delete(&domain)
	return tx.Error
}
