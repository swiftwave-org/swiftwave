package Manager

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// This file consists functions to generate SSL certificate

// This will initiate ACME server to reverse verification
// and store generated certificate in preferred location
// - return fullchain of the certificate, error
func (s Manager) ObtainCertificate(domain string, privateKeyStr string) (string, error) {
	// Check if the domain is pointing to the server
	if !s.VerifyDomain(domain) {
		return "", errors.New("domain is not pointing to the server")
	}

	// Parse the PEM-encoded data
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return "", errors.New("invalid PEM file or key type")
	}

	// Parse the DER-encoded key data
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", errors.New("unable to parse private key for domain")
	}
	certs, err := s.client.ObtainCertificate(s.ctx, s.account, privateKey, []string{domain})
	if err != nil {
		return "", errors.New("unable to obtain certificate")
	}
	// Get the certificate
	certificate := certs[0]
	fullchain := certificate.ChainPEM
	// Convert byte[] to string
	fullchainStr := string(fullchain)
	return fullchainStr, nil
}
