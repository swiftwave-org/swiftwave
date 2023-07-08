package Manager

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strings"

	"github.com/mholt/acmez"
	"github.com/mholt/acmez/acme"
	"gorm.io/gorm"
)

// Manager constructor

// Init Manager
func (s *Manager) Init(ctx context.Context, db gorm.DB, options ManagerOptions) error {
	s.ctx = ctx
	s.dbClient = db
	s.options = options
	s.options.DomainPrivateKeyStorePath = strings.TrimSuffix(s.options.DomainPrivateKeyStorePath, "/")
	s.options.DomainFullChainStorePath = strings.TrimSuffix(s.options.DomainFullChainStorePath, "/")
	// Migrate database
	db.AutoMigrate(&KeyAuthorizationToken{})
	db.AutoMigrate(&DomainSSLDetails{})
	// Initialize account
	acmeDirectory := "https://acme-staging-v02.api.letsencrypt.org/directory"
	if !options.IsStaging {
		acmeDirectory = "https://acme-v02.api.letsencrypt.org/directory"
	}
	s.client = acmez.Client{
		Client: &acme.Client{
			Directory: acmeDirectory,
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true, // REMOVE THIS FOR PRODUCTION USE!
					},
				},
			},
		},
		ChallengeSolvers: map[string]acmez.Solver{
			acme.ChallengeTypeHTTP01: http01Solver{
				dbClient: s.dbClient,
			},
		},
	}
	if options.AccountPrivateKeyFilePath == "" {
		return errors.New("account private key file path is not provided")
	}
	// Init acme account
	acme_account, err := initiateACMEAccount(s.ctx, &s.client, options.AccountPrivateKeyFilePath, options.Email)
	if err != nil {
		return errors.New("error while initiating acme account")
	}
	s.account = acme_account
	return nil
}
