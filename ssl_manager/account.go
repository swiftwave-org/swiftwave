package Manager

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"os"
	"strings"

	"github.com/mholt/acmez"
	"github.com/mholt/acmez/acme"
)

// Initialize the ACME client
func initiateACMEAccount(ctx context.Context, client *acmez.Client, AccountPrivateKeyFilePath string, accountEmail string) (acme.Account, error) {
	// Read the private key from file
	accountPrivateKey, err := fetchAccountPrivateKey(AccountPrivateKeyFilePath)
	if err != nil {
		return acme.Account{}, err
	}

	// ACME Client Account
	account := acme.Account{
		Contact:              []string{"mailto:" + accountEmail},
		TermsOfServiceAgreed: true,
		PrivateKey:           accountPrivateKey,
	}

	var finalAccount acme.Account
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
func fetchAccountPrivateKey(AccountPrivateKeyFilePath string) (*rsa.PrivateKey, error) {
	if !strings.HasSuffix(AccountPrivateKeyFilePath, ".key") {
		return nil, errors.New("invalid account private key file path. file must be .key file")
	}
	// If file does not exist, create a new private key
	if _, err := os.Stat(AccountPrivateKeyFilePath); os.IsNotExist(err) {
		accountPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return nil, errors.New("unable to generate account private key")
		}
		err = storePrivateKeyToFile(AccountPrivateKeyFilePath, accountPrivateKey)
		if err != nil {
			return nil, err
		}
	}

	// Read the private key from file
	accountPrivateKey, err := readPrivateKeyFromFile(AccountPrivateKeyFilePath)
	if err != nil {
		return nil, err
	}
	return accountPrivateKey, nil
}
