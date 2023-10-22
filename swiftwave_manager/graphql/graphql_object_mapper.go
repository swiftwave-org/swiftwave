package graphql

import (
	dbmodel "github.com/swiftwave-org/swiftwave/swiftwave_manager/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_manager/graphql/model"
)

// This file contains object mappers
// 1. Convert Database models to GraphQL models > <type>ToGraphqlObject.go
// 2. Convert GraphQL models to Database models > <type>ToDatabaseObject.go

// gitCredentialToGraphqlObject : converts GitCredential to GitCredentialGraphqlObject
func gitCredentialToGraphqlObject(record *dbmodel.GitCredential) *model.GitCredential {
	return &model.GitCredential{
		ID:       int(record.ID),
		Name:     record.Name,
		Username: record.Username,
		Password: record.Password,
	}
}

// gitCredentialInputToDatabaseObject : converts GitCredentialInput to GitCredentialDatabaseObject
func gitCredentialInputToDatabaseObject(record *model.GitCredentialInput) *dbmodel.GitCredential {
	return &dbmodel.GitCredential{
		Name:     record.Name,
		Username: record.Username,
		Password: record.Password,
	}
}
