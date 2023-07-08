package sslmanager

import (
	"errors"
	"time"
)

// This file consists functions to generate SSL certificate

// This will initiate ACME server to reverse verification
// and store generated certificate in preferred location
// - output: fullchain.pem
func (s SSLManager) ObtainCertificate(domain string) error {
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
	err = storeBytesToPEMFile(certificate.ChainPEM, s.options.DomainFullChainStorePath+"/"+domain+".pem")
	if err != nil {
		return errors.New("unable to store certificate to file")
	}
	// Update the creation date in redis
	tx := s.dbClient.Create(&DomainSSLDetails{
		Domain:   domain,
		CreationDate: time.Now(),
	})
	if tx.Error != nil {
		return errors.New("unable to update creation date in redis")
	}
	return nil
}
