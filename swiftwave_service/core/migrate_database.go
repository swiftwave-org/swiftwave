package core

import (
	"errors"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"gorm.io/gorm"
)

func MigrateDatabase(dbClient *gorm.DB) error {
	// Migrate the schema
	err := dbClient.AutoMigrate(
		&User{},
		&Domain{},
		&RedirectRule{},
		&PersistentVolume{},
		&PersistentVolumeBinding{},
		&PersistentVolumeBackup{},
		&PersistentVolumeRestore{},
		&Application{},
		&GitCredential{},
		&ImageRegistryCredential{},
		&IngressRule{},
		&EnvironmentVariable{},
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
