package sslmanager

import (
	"context"

	"github.com/mholt/acmez"
	"github.com/mholt/acmez/acme"
	"github.com/redis/go-redis/v9"
)

type SSLManager struct {
	ctx         context.Context
	account     acme.Account
	client      acmez.Client
	redisClient redis.Client
	options     SSLManagerOptions
}

type SSLManagerOptions struct {
	Email                      string
	AccountPrivateKeyFilePath  string
	DomainPrivateKeyStorePath  string
	DomainFullChainStorePath string
}

type http01Solver struct {
	redisClient redis.Client
}
