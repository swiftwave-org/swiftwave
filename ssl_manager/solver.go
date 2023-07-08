package Manager

import (
	"context"
	"errors"
	"strings"

	"github.com/mholt/acmez/acme"
)

// This file consists http-01 challenge solver

// Required for acmez.Solver interface
func (s http01Solver) Present(ctx context.Context, chal acme.Challenge) error {
	keyAuthorization := chal.KeyAuthorization
	// keyAuthorization is in the form of "token.base64url(Thumbprint(accountKey))"
	if !strings.Contains(keyAuthorization, ".") {
		return errors.New("invalid key authorization")
	}
	token := strings.Split(keyAuthorization, ".")[0]
	tx := s.dbClient.Create(&KeyAuthorizationToken{
		Token:              token,
		AuthorizationToken: keyAuthorization,
	})
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// Required for acmez.Solver interface
func (s http01Solver) CleanUp(ctx context.Context, chal acme.Challenge) error {
	keyAuthorization := chal.KeyAuthorization
	// keyAuthorization is in the form of "token.base64url(Thumbprint(accountKey))"
	if !strings.Contains(keyAuthorization, ".") {
		return errors.New("invalid key authorization")
	}
	token := strings.Split(keyAuthorization, ".")[0]
	tx := s.dbClient.Delete(&KeyAuthorizationToken{}, "token = ?", token)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// Required for http-01 verification
func (s Manager) fetchKeyAuthorization(token string) string {
	keyAuthorization := KeyAuthorizationToken{}
	tx := s.dbClient.Where("token = ?", token).Find(&keyAuthorization)
	if tx.Error != nil {
		return ""
	}
	return keyAuthorization.AuthorizationToken
}
