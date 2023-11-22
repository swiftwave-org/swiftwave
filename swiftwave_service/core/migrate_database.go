package core

import (
	"errors"
	"gorm.io/gorm"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
)

func MigrateDatabase(dbClient *gorm.DB) error {
	// Migrate the schema
	err := dbClient.AutoMigrate(
		&Domain{},
		&RedirectRule{},
		&PersistentVolume{},
		&Application{},
		&GitCredential{},
		&ImageRegistryCredential{},
		&IngressRule{},
		&EnvironmentVariable{},
		&PersistentVolumeBinding{},
		&Deployment{},
		&BuildArg{},
		&DeploymentLog{},
		&SSL.KeyAuthorizationToken{},
	)
	if err != nil {
		return errors.New("failed to migrate database \n" + err.Error())
	}
	return nil
}
