package sslmanager

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strings"

	"github.com/mholt/acmez"
	"github.com/mholt/acmez/acme"
	"github.com/redis/go-redis/v9"
)

// SSLManager constructor

// Init SSLManager
func (s *SSLManager) Init(ctx context.Context, redisClient redis.Client, options SSLManagerOptions) error {
	s.ctx = ctx
	s.redisClient = redisClient
	s.options = options
	s.options.DomainPrivateKeyStorePath = strings.TrimSuffix(s.options.DomainPrivateKeyStorePath, "/")
	s.options.DomainFullChainStorePath = strings.TrimSuffix(s.options.DomainFullChainStorePath, "/")
	// Initialize account
	s.client = acmez.Client{
		Client: &acme.Client{
			Directory: "https://acme-staging-v02.api.letsencrypt.org/directory",
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
				redisClient: s.redisClient,
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
