package sslmanager

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
)

// Store the private key to a file
func storeKeyToFile(keyFile string, key *ecdsa.PrivateKey) error {
	// Encode the private key to PEM format
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}

	pemKey := pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}

	// Create the PEM file
	file, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the PEM-encoded key to the file
	err = pem.Encode(file, &pemKey)
	if err != nil {
		return err
	}
	return nil
}

// Read the private key from a file
func readKeyFromFile(keyFile string)(*ecdsa.PrivateKey, error) {
	keyData, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, errors.New("unable to read account private key file")
	}

	// Parse the PEM-encoded data
	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, errors.New("invalid PEM file or key type")
	}

	// Parse the DER-encoded key data
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("unable to parse account private key")
	}
	return privateKey, nil
}