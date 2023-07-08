package Manager

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// Store the private key to a file
func storePrivateKeyToFile(keyFile string, key *rsa.PrivateKey) error {
	// Encode the private key to PEM format
	keyBytes := x509.MarshalPKCS1PrivateKey(key)

	pemKey := pem.Block{
		Type:  "RSA PRIVATE KEY",
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
func readPrivateKeyFromFile(keyFile string) (*rsa.PrivateKey, error) {
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

// Store byte[] to PEM file
func storeBytesToCRTFile(bytes []byte, pemFile string) error {
	// pemKey := pem.Block{
	// 	Type:  "CERTIFICATE",
	// 	Bytes: bytes,
	// }

	data := string(bytes)

	// Create the PEM file
	file, err := os.Create(pemFile)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(data)

	// Write the PEM-encoded key to the file
	// err = pem.Encode(file, &pemKey)
	// if err != nil {
	// 	return err
	// }
	return nil
}

// Get private key for a domain
// -- Create a private key if it doesn't exist
// -- Read the private key from file if it exists

func fetchPrivateKeyForDomain(domain string, certsPrivateKeyDirectory string) (*rsa.PrivateKey, error) {
	privateKeyFile := certsPrivateKeyDirectory + "/" + domain + ".key"
	privateKey, err := readPrivateKeyFromFile(privateKeyFile)
	if err == nil {
		return privateKey, nil
	} else {
		// Create a private key
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
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
