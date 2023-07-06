package sslmanager

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"os"
	"strings"

	"github.com/mholt/acmez/acme"
)

// Initialize the ACME client
func initiateACMEAccount(ctx context.Context, client *acme.Client, accountPrivateKeyFilePath string, accountEmail string) (acme.Account, error){
	// Read the private key from file
	accountPrivateKey, err := fetchAccountPrivateKey(accountPrivateKeyFilePath)
	if err != nil {
		return acme.Account{}, err
	}

	// ACME Client Account
	account := acme.Account{
		Contact:              []string{"mailto:"+accountEmail},
		TermsOfServiceAgreed: true,
		PrivateKey:           accountPrivateKey,
	}

	var finalAccount acme.Account;
	// Try to fetch the account from server using same private key
	fetchedAccount, err := client.GetAccount(ctx, account)
	if err == nil {
		finalAccount = fetchedAccount
	} else {
		// If account does not exist, create a new one
		finalAccount, err = client.NewAccount(ctx, account)
		if err != nil {
			return acme.Account{}, err
		}
	}
	// Return the account
	return finalAccount, nil
}

// Fetch the account private key from file
// If file does not exist, it will create a new private key and store it in file
func fetchAccountPrivateKey(accountPrivateKeyFilePath string) (*ecdsa.PrivateKey, error){
	if !strings.HasSuffix(accountPrivateKeyFilePath, ".pem") {
		return nil, errors.New("invalid account private key file path. file must be .pem file")
	}
	// If file does not exist, create a new private key
	if _, err := os.Stat(accountPrivateKeyFilePath); os.IsNotExist(err) {
		accountPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, errors.New("unable to generate account private key")
		}
		err = storeKeyToFile(accountPrivateKeyFilePath, accountPrivateKey)
		if err != nil {
			return nil, err
		}
	}

	// Read the private key from file
	accountPrivateKey, err := readKeyFromFile(accountPrivateKeyFilePath)
	if err != nil {
		return nil, err
	}
	return accountPrivateKey, nil
}