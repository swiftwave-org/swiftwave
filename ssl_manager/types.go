package Manager

import (
	"context"
	"github.com/mholt/acmez"
	"github.com/mholt/acmez/acme"
	"gorm.io/gorm"
)

type Manager struct {
	ctx      context.Context
	account  acme.Account
	client   acmez.Client
	dbClient gorm.DB
	options  ManagerOptions
}

type ManagerOptions struct {
	IsStaging         bool
	Email             string
	AccountPrivateKey string
}

type http01Solver struct {
	dbClient gorm.DB
}

// GORM Models
type KeyAuthorizationToken struct {
	Token              string `gorm:"primaryKey"`
	AuthorizationToken string
}
