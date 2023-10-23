package graphql

import (
	dbmodel "github.com/swiftwave-org/swiftwave/swiftwave_manager/core"
	"github.com/swiftwave-org/swiftwave/swiftwave_manager/graphql/model"
)

// This file contains object mappers
// 1. Convert Database models to GraphQL models > <type>ToGraphqlObject.go
// 2. Convert GraphQL models to Database models > <type>ToDatabaseObject.go

// Why _ToDatabaseObject() dont adding ID field?
// Because ID field is provided directly to Mutation or Query function

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

// imageRegistryCredentialToGraphqlObject : converts ImageRegistryCredential to ImageRegistryCredentialGraphqlObject
func imageRegistryCredentialToGraphqlObject(record *dbmodel.ImageRegistryCredential) *model.ImageRegistryCredential {
	return &model.ImageRegistryCredential{
		ID:       int(record.ID),
		URL:      record.Url,
		Username: record.Username,
		Password: record.Password,
	}
}

// imageRegistryCredentialInputToDatabaseObject : converts ImageRegistryCredentialInput to ImageRegistryCredentialDatabaseObject
func imageRegistryCredentialInputToDatabaseObject(record *model.ImageRegistryCredentialInput) *dbmodel.ImageRegistryCredential {
	return &dbmodel.ImageRegistryCredential{
		Url:      record.URL,
		Username: record.Username,
		Password: record.Password,
	}
}

// persistentVolumeToGraphqlObject : converts PersistentVolume to PersistentVolumeGraphqlObject
func persistentVolumeToGraphqlObject(record *dbmodel.PersistentVolume) *model.PersistentVolume {
	return &model.PersistentVolume{
		ID:   int(record.ID),
		Name: record.Name,
	}
}

// persistentVolumeInputToDatabaseObject : converts PersistentVolumeInput to PersistentVolumeDatabaseObject
func persistentVolumeInputToDatabaseObject(record *model.PersistentVolumeInput) *dbmodel.PersistentVolume {
	return &dbmodel.PersistentVolume{
		Name: record.Name,
	}
}
