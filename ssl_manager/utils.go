package Manager

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// Decode the private key from a private key string
func decodePrivateKey(key string) (*rsa.PrivateKey, error) {
	keyData := []byte(key)
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

func (s Manager) FetchIssuerName() string {
	if s.options.IsStaging {
		return "Let's Encrypt (Staging)"
	} else {
		return "Let's Encrypt"
	}
}
