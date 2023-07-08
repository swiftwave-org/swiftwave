package Manager

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// Store the private key to a file
func storePrivateKeyToFile(keyFile string, key *ecdsa.PrivateKey) error {
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
func readPrivateKeyFromFile(keyFile string) (*ecdsa.PrivateKey, error) {
	keyData, err := os.ReadFile(keyFile)
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
		return nil, errors.New("unable to parse private key")
	}
	return privateKey, nil
}

// Store byte[] to PEM file
func storeBytesToPEMFile(bytes []byte, pemFile string) error {
	pemKey := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: bytes,
	}

	// Create the PEM file
	file, err := os.Create(pemFile)
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

// Get private key for a domain
// -- Create a private key if it doesn't exist
// -- Read the private key from file if it exists

func fetchPrivateKeyForDomain(domain string, certsPrivateKeyDirectory string) (*ecdsa.PrivateKey, error) {
	privateKeyFile := certsPrivateKeyDirectory + "/" + domain + ".pem"
	privateKey, err := readPrivateKeyFromFile(privateKeyFile)
	if err == nil {
		return privateKey, nil
	} else {
		// Create a private key
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, errors.New("unable to generate private key")
		}
		// Store the private key to file
		err = storePrivateKeyToFile(privateKeyFile, privateKey)
		if err != nil {
			return nil, errors.New("unable to store private key to file")
		}
		return privateKey, nil
	}
}
