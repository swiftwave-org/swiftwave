package Manager

import (
	"errors"
	"time"
)

// This file consists functions to generate SSL certificate

// This will initiate ACME server to reverse verification
// and store generated certificate in preferred location
// - output: fullchain.crt
func (s Manager) ObtainCertificate(domain string) error {
	// Check if the domain is pointing to the server
	if !s.VerifyDomain(domain) {
		return errors.New("domain is not pointing to the server")
	}
	// Generate private key
	privateKey, err := fetchPrivateKeyForDomain(domain, s.options.DomainPrivateKeyStorePath)
	if err != nil {
		return errors.New("unable to fetch private key for domain")
	}
	certs, err := s.client.ObtainCertificate(s.ctx, s.account, privateKey, []string{domain})
	if err != nil {
		return errors.New("unable to obtain certificate")
	}
	// Get the certificate
	certificate := certs[0]
	// Store the certificate to file
	err = storeBytesToCRTFile(certificate.ChainPEM, s.options.DomainFullChainStorePath+"/"+domain+".crt")
	if err != nil {
		return errors.New("unable to store certificate to file")
	}
	// Update the creation date in redis
	// -- check if the domain is already in database
	var domainSSLDetails DomainSSLDetails
	tx := s.dbClient.Where("domain = ?", domain).First(&domainSSLDetails)
	if tx.Error != nil {
		tx := s.dbClient.Create(&DomainSSLDetails{
			Domain:       domain,
			CreationDate: time.Now(),
		})
		if tx.Error != nil {
			return errors.New("unable to create entry in database")
		} else {
			return nil
		}
	} else {
		// Update the creation date
		tx = s.dbClient.Model(&domainSSLDetails).Update("creation_date", time.Now())
		if tx.Error != nil {
			return errors.New("unable to update creation date in database")
		}
		return nil
	}
}
