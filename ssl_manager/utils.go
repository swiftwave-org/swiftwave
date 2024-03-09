package Manager

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// Decode the private key from a private key string
func decodePrivateKey(keyFile string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, errors.New("unable to read account private key file")
	}

	// Parse the PEM-encoded data
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid PEM file or key type")
	}

	// Parse the DER-encoded key data
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("unable to parse private key")
	}
	return privateKey, nil
}

// Fetch SSL Issuer's Name
func (s Manager) FetchIssuerName() string {
	if s.options.IsStaging {
		return "Let's Encrypt (Staging)"
	} else {
		return "Let's Encrypt"
	}
}
