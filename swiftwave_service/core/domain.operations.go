package core

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"gorm.io/gorm"
	"time"
)

// This file contains the operations for the Domain model.
// This functions will perform necessary validation before doing the actual database operation.

// Each function's argument format should be (ctx context.Context, db gorm.DB, ...)
// context used to pass some data to the function e.g. user id, auth info, etc.

func FindAllDomains(_ context.Context, db gorm.DB) ([]*Domain, error) {
	var domains []*Domain
	tx := db.Find(&domains)
	return domains, tx.Error
}

func FetchDomainsThoseWillExpire(_ context.Context, db gorm.DB, daysToExpire int) ([]*Domain, error) {
	var domains []*Domain
	tx := db.Where("ssl_status = ?", DomainSSLStatusIssued).Where("ssl_auto_renew = ?", true).Where("ssl_expired_at < ?", time.Now().AddDate(0, 0, daysToExpire)).Find(&domains)
	return domains, tx.Error
}

func (domain *Domain) FindById(_ context.Context, db gorm.DB, id uint) error {
	tx := db.Where("id = ?", id).First(&domain)
	return tx.Error
}

func (domain *Domain) Create(_ context.Context, db gorm.DB) error {
	err := domain.fillSSLInfo()
	if err != nil {
		return err
	}
	tx := db.Create(&domain)
	return tx.Error
}

func (domain *Domain) Update(_ context.Context, db gorm.DB) error {
	err := domain.fillSSLInfo()
	if err != nil {
		return err
	}
	tx := db.Save(&domain)
	return tx.Error
}

func (domain *Domain) Delete(_ context.Context, db gorm.DB) error {
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

func (domain *Domain) UpdateSSLStatus(_ context.Context, db gorm.DB, status DomainSSLStatus) error {
	domain.SSLStatus = status
	tx := db.Where("id = ?", domain.ID).Update("ssl_status", status)
	return tx.Error
}

func (domain *Domain) fillSSLInfo() error {
	if domain == nil || domain.SSLFullChain == "" {
		return nil
	}
	certBytes := []byte(domain.SSLFullChain)
	block, _ := pem.Decode(certBytes)
	if block == nil {
		return errors.New("failed to decode SSL full chain certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.New("failed to parse SSL full chain certificate")
	}
	domain.SSLIssuedAt = cert.NotBefore
	domain.SSLExpiredAt = cert.NotAfter
	var sslIssuer = "Unknown Issuer"
	if len(cert.Issuer.Organization) > 0 {
		sslIssuer = cert.Issuer.Organization[0]
	}
	domain.SSLIssuer = sslIssuer
	return nil
}
