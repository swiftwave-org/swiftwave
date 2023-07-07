package sslmanager

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
	token := "token_"+strings.Split(keyAuthorization, ".")[0]
	status :=  s.redisClient.Set(ctx, token, keyAuthorization, 0)
	if status.Err() != nil {
		return status.Err()
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
	token := "token_"+strings.Split(keyAuthorization, ".")[0]
	status :=  s.redisClient.Del(ctx, token)
	if status.Err() != nil {
		return status.Err()
	}
	return nil
}

// Required for http-01 verification
func (s SSLManager) fetchKeyAuthorization(token string) string {
	keyAuthorization, err := s.redisClient.Get(s.ctx, "token_"+token).Result()
	if err != nil {
		return ""
	}
	return keyAuthorization
}