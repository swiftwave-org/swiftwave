package core

import (
	"errors"
	SSL "github.com/swiftwave-org/swiftwave/ssl_manager"
	"gorm.io/gorm"
)

func MigrateDatabase(dbClient *gorm.DB) error {
	// Migrate the schema
	err := dbClient.AutoMigrate(
		&SystemConfig{},
		&SystemLog{},
		&Server{},
		&ServerLog{},
		&User{},
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
		&PersistentVolumeBackup{},
		&PersistentVolumeRestore{},
	)
	if err != nil {
		return errors.New("failed to migrate database \n" + err.Error())
	}
	return nil
}
