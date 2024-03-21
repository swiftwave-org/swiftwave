package Manager

import (
	"context"
	"github.com/mholt/acmez"
	"github.com/mholt/acmez/acme"
)

// Initialize the ACME client
func initiateACMEAccount(ctx context.Context, client *acmez.Client, accountPrivateKey string, accountEmail string) (acme.Account, error) {
	// Read the private key from file
	accountPrivateRSAKey, err := decodePrivateKey(accountPrivateKey)
	if err != nil {
		return acme.Account{}, err
	}

	// ACME Client Account
	account := acme.Account{
		Contact:              []string{"mailto:" + accountEmail},
		TermsOfServiceAgreed: true,
		PrivateKey:           accountPrivateRSAKey,
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
